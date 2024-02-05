package user

import (
	"errors"
	"fmt"

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

	// Get the users available UTXOs
	utxos, err := getUsersUTXOs(senderAddress)
	if err != nil {
		return
	}

	// Calculate the summ of all receivers + multisig fee + potential transaction fee
	// UPDATETODO: Do some calculation for potential fee instead of 200000
	sendAmount := uint(0) + multisigFee
	potentialFee := uint(200000)
	for _, value := range receiversAndAmounts {
		sendAmount += value
	}

	// Firstly check if we have input that can satisfy whole amount
	var firstMatchInput = tx.TxInput{
		Marshaler: nil,
		TxHash:    []byte{},
		Index:     0,
		Amount:    0,
	}

	// Loop through utxos to find first input with enough tokens
	// If we don't have this UTXO we need to use more of them
	var amountSum = uint(0)
	var chosenUTXOs []tx.TxInput

	for _, utxo := range utxos {
		if utxo.Amount >= sendAmount+potentialFee {
			firstMatchInput = utxo
			break
		}

		amountSum += utxo.Amount
		chosenUTXOs = append(chosenUTXOs, utxo)

		if amountSum >= sendAmount+potentialFee {
			break
		}
	}

	// Check if address have sufficent amount for transaction
	if firstMatchInput.Amount == 0 && amountSum < sendAmount+potentialFee {
		err = errors.New("no enough available funds for generating transaction " + fmt.Sprint(amountSum) + " available but " + fmt.Sprint(sendAmount+potentialFee) + " required")
		return
	}

	// Instantiate transaction builder
	pr, err := getProtocolParameters()
	if err != nil {
		return
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{},
	)

	// Add inputs to transaction
	if firstMatchInput.Amount != 0 {
		builder.AddInputs(&firstMatchInput)
	} else {
		for _, utxo := range chosenUTXOs {
			builder.AddInputs(&utxo)
		}
	}

	// 2. Check if all provided addresses are correct

	// Loop trough addresses check correctness
	// Add outputs to the transaction
	// Fill metadata values
	builder.Tx().AuxiliaryData = tx.NewAuxiliaryData()
	builder.Tx().AuxiliaryData.AddMetadataElement("chainId", chainId)

	for addressString, amount := range receiversAndAmounts {
		_, err = address.NewAddress(addressString)
		if err != nil {
			return
		}

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
	slot, err := getSlotNumber()
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
