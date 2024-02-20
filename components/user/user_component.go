package user

import (
	"errors"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/tx"
)

// I as a user want to create the bridging transaction
// without knowing the actual logic of the bridge
// I will pass the necessary arguments and sign the transaction

func CreateBridgingTransaction(sender string, chainId string, receiversAndAmounts map[string]uint) (transaction tx.Tx, err error) {
	if sender == "" {
		err = errors.New("sender address cannot be empty string")
		return
	}

	// UPDATETODO: This function panics - potentialy handle better
	senderAddress, err := address.NewAddress(sender)
	if err != nil {
		return
	}

	if chainId == "" {
		err = errors.New("chainId cannot be empty string")
		return
	}

	multisigAddressString, multisigFeeAddressString, multisigFee, err := GetChainData(chainId)
	if err != nil {
		return
	}

	if len(receiversAndAmounts) == 0 {
		err = errors.New("no receivers defined")
		return
	}

	// Validaton:

	// 1. Check if user has enough funds for the transaction

	// Calculate the summ of all receivers + multisig fee + potential transaction fee
	// UPDATETODO: Do some calculation for potential fee instead of 200000
	// UPDATETODO: MinUtxoValue in protocol parameters doesn't match the real value,
	// currently on preview testnet minUtxoValue is null, but realy it is "hidden".
	// It can be calculated as: minUTxoVal = (160 + sizeInBytes (TxOut)) * coinsPerUTxOByte
	minUtxoValue := uint(1000000)
	sendAmount := uint(0) + multisigFee
	potentialFee := uint(200000)
	for receiverAddress, amount := range receiversAndAmounts {
		_, err = address.NewAddress(receiverAddress)
		if err != nil {
			return
		}

		if amount < minUtxoValue {
			err = errors.New("receivers amount cannot be under 1000000 tokens")
			return
		}

		sendAmount += amount
	}

	// Instantiate transaction builder
	pr, err := getProtocolParameters(chainId)
	if err != nil {
		return
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{},
	)

	// Get the users UTXOs
	utxos, err := getUsersUTXOs(chainId, senderAddress, sendAmount, potentialFee)
	if err != nil {
		return
	}

	builder.AddInputs(utxos...)

	// 2. Check if all provided addresses are correct

	// Loop trough addresses check correctness
	// Add outputs to the transaction
	// Fill metadata values
	builder.Tx().AuxiliaryData = tx.NewAuxiliaryData()
	if chainId == "prime" {
		builder.Tx().AuxiliaryData.AddMetadataElement("chainId", "vector")
	} else {
		builder.Tx().AuxiliaryData.AddMetadataElement("chainId", "prime")
	}

	for addressString, amount := range receiversAndAmounts {
		builder.Tx().AuxiliaryData.AddMetadataTransaction(addressString, amount)
	}

	// Add multisig fee address and amount to metadata
	builder.Tx().AuxiliaryData.AddMetadataTransaction(multisigFeeAddressString, multisigFee)

	// Add transaction output for multisig address
	multisigAddress, err := address.NewAddress(multisigAddressString)
	if err != nil {
		return
	}
	builder.AddOutputs(tx.NewTxOutput(
		multisigAddress,
		uint(sendAmount),
	))

	// Query slot from a node on the network.
	// Slot is needed to compute TTL of transaction.
	slot, err := getSlotNumber(chainId)
	if err != nil {
		return
	}

	// Set TTL for 5 min into the future
	builder.SetTTL(uint32(slot) + uint32(300))

	// Route back the change to the sender address
	// This is equivalent to adding an output with the source address and change amount
	builder.AddChangeIfNeeded(senderAddress)

	// Build transaction
	transaction, err = builder.Build()
	if err != nil {
		return
	}

	return
}
