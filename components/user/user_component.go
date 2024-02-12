package user

import (
	"errors"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/components/txhelper"
	"github.com/fivebinaries/go-cardano-serialization/tx"
)

// I as a user want to create the bridging transaction
// without knowing the actual logic of the bridge
// I will pass the necessary arguments and sign the transaction

func CreateBridgingTransaction(sender string, chainId string, receiversAndAmounts map[string]uint) (transaction tx.Tx, err error) {
	if sender == "" {
		return tx.Tx{}, errors.New("sender address cannot be empty string")
	}

	// UPDATETODO: This function panics - potentialy handle better
	senderAddress, err := address.NewAddress(sender)
	if err != nil {
		return
	}

	if chainId == "" {
		return tx.Tx{}, errors.New("chainId cannot be empty string")
	}

	multisigAddressString, multisigFeeAddressString, multisigFee, err := GetChainData(chainId)
	if err != nil {
		return
	}

	if len(receiversAndAmounts) == 0 {
		return tx.Tx{}, errors.New("no receivers defined")
	}

	// Validaton:

	// 1. Check if user has enough funds for the transaction

	// Calculate the summ of all receivers + multisig fee + potential transaction fee
	// UPDATETODO: Do some calculation for potential fee instead of 200000
	// UPDATETODO: MinUtxoValue in protocol parameters doesn't match the real value,
	// currently on preview testnet minUtxoValue is null, but realy it is "hidden".
	// It can be calculated as: minUTxoVal = (160 + sizeInBytes (TxOut)) * coinsPerUTxOByte
	sendAmount, err := txhelper.GetSumFromRecipients(receiversAndAmounts)
	if err != nil {
		return tx.Tx{}, err
	}

	sendAmount += multisigFee

	const potentialFee = uint(200000)

	nodeUrl, err := getNodeUrl(chainId)
	if err != nil {
		return tx.Tx{}, err
	}

	txParamsHelper := txhelper.NewTxParamsHelper(nodeUrl)

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

	allUtxos, err := txParamsHelper.GetUTXOs(senderAddress.String())
	if err != nil {
		return
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{},
	)

	// Get the users UTXOs
	utxos, err := txhelper.GetUTXOsForAmount(allUtxos, sendAmount, potentialFee)
	if err != nil {
		return
	}

	builder.AddInputs(utxos...)

	// 2. Check if all provided addresses are correct

	// Loop trough addresses check correctness
	// Add outputs to the transaction
	// Fill metadata values
	builder.Tx().AuxiliaryData = tx.NewAuxiliaryData()
	builder.Tx().AuxiliaryData.AddMetadataElement("chainId", chainId)

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

	// Set TTL for 5 min into the future
	builder.SetTTL(uint32(slot) + uint32(300))

	// Route back the change to the sender address
	// This is equivalent to adding an output with the source address and change amount
	builder.AddChangeIfNeeded(senderAddress)

	// Build transaction
	return builder.Build()
}
