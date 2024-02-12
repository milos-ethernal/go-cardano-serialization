package batcher

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/components"
	"github.com/fivebinaries/go-cardano-serialization/components/txhelper"
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
	receiversUTXOs, err := txhelper.CreateUtxos(receivers, false)
	if err != nil {
		return nil, err
	}

	multisigAddressString, multisigFeeAddressString, _, err := GetChainData(destinationChain)
	if err != nil {
		return
	}

	txParamsHelper := txhelper.NewTxParamsHelper("")
	outputsSum := txhelper.GetTxOutputsSum(receiversUTXOs)

	// Get input UTXOs: currently mocked to get the UTXOs directly from chain
	// UPDATETODO: Query data from contract
	allUtxos, err := txParamsHelper.GetUTXOs(multisigAddressString)
	if err != nil {
		return
	}

	// Get input UTXOs: currently mocked to get the UTXOs directly from chain
	// UPDATETODO: Query data from contract
	allFeeUtxos, err := txParamsHelper.GetUTXOs(multisigFeeAddressString)
	if err != nil {
		return
	}

	inputs, err := txhelper.GetUTXOsForAmount(allUtxos, outputsSum, 0)
	if err != nil {
		return
	}

	// Add multisig fee input
	feeInputs, err := txhelper.GetUTXOsForAmount(allFeeUtxos, outputsSum, 0)
	if err != nil {
		return
	}

	// Generate batching transaction
	// Instantiate transaction builder
	pr, err := txParamsHelper.GetProtocolParameters()
	if err != nil {
		return
	}

	// Query slot from a node on the network.
	// Slot is needed to compute TTL of transaction.
	slot, err := txParamsHelper.GetSlotNumber()
	if err != nil {
		return
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{},
	)

	builder.AddInputs(inputs...)

	multisigAddress, err := address.NewAddress(multisigAddressString)
	if err != nil {
		return
	}

	// Add change for multisig address if it exists
	// In case of adding 0 output node throws an error on transaction submit
	inputSum := txhelper.GetTxInputsSum(inputs)
	if diff := inputSum - outputsSum; diff > 0 {
		builder.AddOutputs(tx.NewTxOutput(
			multisigAddress,
			diff,
		))
	}

	builder.AddInputs(feeInputs...)

	// Add receiver outputs
	for _, utxo := range receiversUTXOs {
		builder.AddOutputs(utxo)
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

	multisigScript := txhelper.CreateSignersScript([][]byte{
		firstSignerKeyHash, secondSignerKeyHash,
		thirdSignerKeyHash, fourthSignerKeyHash,
	})

	// Create script of fee multisig address
	firstFeeSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_FEE_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH1"))
	secondFeeSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_FEE_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH2"))
	thirdFeeSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_FEE_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH3"))
	fourthFeeSignerKeyHash, _ := hex.DecodeString(os.Getenv("MULTISIG_FEE_ADDRESS_" + strings.ToUpper(destinationChain) + "_KEYHASH4"))

	multisigFeeScript := txhelper.CreateSignersScript([][]byte{
		firstFeeSignerKeyHash, secondFeeSignerKeyHash,
		thirdFeeSignerKeyHash, fourthFeeSignerKeyHash,
	})

	// UPDATETODO: Get this parameter from chain
	// Add metadata batch nonce id
	builder.Tx().AuxiliaryData = tx.NewAuxiliaryData()

	id, err := getBatchNonceId()
	if err != nil {
		return
	}

	builder.Tx().AuxiliaryData.AddMetadataElement("batch_nonce_id", id)

	// Set arbitrary value for witnesses
	builder.Tx().WitnessSet.Witnesses = append(builder.Tx().WitnessSet.Witnesses, txhelper.GetDummyWitnesses((int(multisigScript.N)+1)*2)...)

	// Set multisig NativeScript
	builder.Tx().WitnessSet.Scripts = []tx.NativeScript{multisigScript, multisigFeeScript}

	// Calculate fee
	multisigFeeAddress, err := address.NewAddress(multisigFeeAddressString)
	if err != nil {
		return
	}

	builder.AddChangeIfNeeded(multisigFeeAddress)

	return builder.Tx(), err
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
