package batcher

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/components"
	"github.com/fivebinaries/go-cardano-serialization/internal/bech32/cbor"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
)

// Builds batching transaction
func BuildBatchingTx(destinationChain string) (transaction *tx.Tx, err error) {
	// UPDATETODO: chainId validation

	// Get all confirmed transactions
	receivers, err := getConfirmedTxs(destinationChain)
	if err != nil {
		return
	}

	// Calculate amount for UTXO
	amountSum := uint(0)
	var receiversUTXOs []tx.TxOutput
	var receiver address.Address
	for addressString, amount := range receivers {
		amountSum += amount

		receiver, err = address.NewAddress(addressString)
		if err != nil {
			return
		}
		receiversUTXOs = append(receiversUTXOs, *tx.NewTxOutput(receiver, amount))
	}

	multisigAddressString, multisigFeeAddressString, _, err := GetChainData(destinationChain)
	if err != nil {
		return
	}

	// Get input UTXOs
	inputs, err := getUTXOs(multisigAddressString, amountSum)
	if err != nil {
		return
	}

	// Generate batching transaction
	// Instantiate transaction builder
	pr, err := getProtocolParameters()
	if err != nil {
		return
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{},
	)

	inputSum := uint(0)
	for _, utxo := range inputs {
		inputSum += utxo.Amount
		builder.AddInputs(&utxo)
	}

	multisigAddress, err := address.NewAddress(multisigAddressString)
	if err != nil {
		return
	}

	// Add change for multisig address if it exists
	// In case of adding 0 output node throws an error on transaction submit
	if inputSum-amountSum > 0 {
		builder.AddOutputs(tx.NewTxOutput(
			multisigAddress,
			inputSum-amountSum,
		))
	}

	// Add multisig fee input
	feeInputs, err := getUTXOs(multisigFeeAddressString, amountSum)
	if err != nil {
		return
	}
	for _, utxo := range feeInputs {
		builder.AddInputs(&utxo)
	}

	// Add receiver outputs
	for _, utxo := range receiversUTXOs {
		builder.AddOutputs(&utxo)
	}

	// Query slot from a node on the network.
	// Slot is needed to compute TTL of transaction.
	slot, err := getSlotNumber()
	if err != nil {
		return
	}

	// Set TTL for 5 min into the future
	builder.SetTTL(uint32(slot) + uint32(300))

	err = godotenv.Load()
	if err != nil {
		return
	}

	// Create script of multisig address
	firstSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH1"))
	secondSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH2"))
	thirdSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH3"))
	fourthSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH4"))

	multisigScript := tx.NativeScript{
		Type:    3,
		KeyHash: []byte{},
		N:       3,
		Scripts: []tx.NativeScript{
			{
				Type:          0,
				KeyHash:       firstSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
			{
				Type:          0,
				KeyHash:       secondSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
			{
				Type:          0,
				KeyHash:       thirdSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
			{
				Type:          0,
				KeyHash:       fourthSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
		},
		IntervalValue: 0,
	}

	// Create script of fee multisig address
	firstFeeSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_FEE_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH1"))
	secondFeeSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_FEE_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH2"))
	thirdFeeSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_FEE_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH3"))
	fourthFeeSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_FEE_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH4"))

	multisigFeeScript := tx.NativeScript{
		Type:    3,
		KeyHash: []byte{},
		N:       3,
		Scripts: []tx.NativeScript{
			{
				Type:          0,
				KeyHash:       firstFeeSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
			{
				Type:          0,
				KeyHash:       secondFeeSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
			{
				Type:          0,
				KeyHash:       thirdFeeSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
			{
				Type:          0,
				KeyHash:       fourthFeeSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
		},
		IntervalValue: 0,
	}

	// UPDATETODO: Get this parameter from chain
	// Add metadata batch nonce id
	builder.Tx().AuxiliaryData = tx.NewAuxiliaryData()

	id, err := getBatchNonceId()
	if err != nil {
		return
	}
	builder.Tx().AuxiliaryData.AddMetadataElement("batch_nonce_id", id)

	// Set arbitrary value for witnesses
	for i := 0; i < (int(multisigScript.N)+1)*2; i++ {
		vWitness := tx.NewVKeyWitness(
			make([]byte, 32),
			make([]byte, 64),
		)
		builder.Tx().WitnessSet.Witnesses = append(builder.Tx().WitnessSet.Witnesses, vWitness)
	}

	// Set multisig NativeScript
	builder.Tx().WitnessSet.Scripts = []tx.NativeScript{multisigScript, multisigFeeScript}

	// Calculate fee
	multisigFeeAddress, err := address.NewAddress(multisigFeeAddressString)
	if err != nil {
		return
	}

	totalI, totalO := builder.GetTotalInputOutputs()
	builder.AddOutputs(tx.NewTxOutput(
		multisigFeeAddress,
		uint(totalI-totalO),
	))

	// Calculate fee
	builder.Tx().SetFee(builder.MinFee())

	// Update multisig change to = input - fee - output
	change := totalI - totalO - uint(builder.Tx().Body.Fee)

	builder.Tx().Body.Outputs[len(builder.Tx().Body.Outputs)-1].Amount = change

	transaction = builder.Tx()
	return
}

// Witness batching transaction
func WitnessBatchingTx(transaction tx.Tx, pkSeed string) (witness tx.VKeyWitness, err error) {
	seed, err := hex.DecodeString(pkSeed)
	if err != nil {
		return
	}

	pk, err := bip32.NewXPrv(seed)
	if err != nil {
		return
	}

	hash, _ := transaction.Hash()
	publicKey := pk.Public().PublicKey()
	signature := pk.Sign(hash[:])
	witness = tx.NewVKeyWitness(publicKey, signature[:])

	return
}

// UPDATETODO: Submit data to bridge chain
// Mocked to write to file instead to bridge chain for testing purposes
// UPDATETODO: Remove id parameter when updateing
func SubmitBatchingTx(transaction tx.Tx, witness tx.VKeyWitness, id string) (err error) {
	transaction.WitnessSet.Witnesses = []tx.VKeyWitness{}

	writeToFile := components.Submit{
		Transaction: transaction,
		Witness:     witness,
	}

	bytesToWrite, err := cbor.Marshal(writeToFile)
	if err != nil {
		return
	}

	// Write byte arrays to a file
	file, err := os.Create(filepath.Join("/tmp", "tx_and_witness_"+id))
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(bytesToWrite)
	if err != nil {
		return
	}

	return nil
}
