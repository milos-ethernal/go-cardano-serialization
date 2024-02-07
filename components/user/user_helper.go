package user

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/protocol"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
)

// Get users available UTXOs
func getUsersUTXOs(address address.Address, amount uint, potentialFee uint) (chosenUTXOs []*tx.TxInput, err error) {
	ogmios := node.NewOgmiosNode("http://localhost:1337")

	utxos, err := ogmios.UTXOs(address)
	if err != nil {
		return []*tx.TxInput{}, err
	}

	// Loop through utxos to find first input with enough tokens
	// If we don't have this UTXO we need to use more of them
	var amountSum = uint(0)
	var minUtxoValue = uint(1000000)

	for _, utxo := range utxos {
		if utxo.Amount >= amount+potentialFee+minUtxoValue {
			chosenUTXOs = []*tx.TxInput{&utxo}
			break
		}

		amountSum += utxo.Amount
		chosenUTXOs = append(chosenUTXOs, &tx.TxInput{
			Marshaler: nil,
			TxHash:    utxo.TxHash,
			Index:     utxo.Index,
			Amount:    utxo.Amount,
		})

		if amountSum >= amount+potentialFee+minUtxoValue {
			break
		}
	}

	if amountSum < amount+potentialFee+minUtxoValue {
		err = errors.New("no enough available funds for generating transaction " + fmt.Sprint(amountSum) + " available but " + fmt.Sprint(amount+potentialFee+minUtxoValue) + " required")
		return
	}
	return
}

// Get protocol parameters
func getProtocolParameters() (protocol.Protocol, error) {
	ogmios := node.NewOgmiosNode("http://localhost:1337")
	return ogmios.ProtocolParameters()
}

// Get slot number
func getSlotNumber() (slot uint, err error) {
	ogmios := node.NewOgmiosNode("http://localhost:1337")
	tip, err := ogmios.QueryTip()
	if err != nil {
		return
	}
	slot = tip.Slot
	return
}

// Get neccessary data for transaction creation
func GetChainData(chainId string) (multisigAddress string, multisigFeeAddress string, multisigFee uint, err error) {
	// UPDATETODO: Query data from contract

	// Load env variables
	err = godotenv.Load()
	if err != nil {
		return
	}

	if chainId == "prime" {
		multisigAddress = os.Getenv("MULTISIG_ADDRESS_PRIME")
		multisigFeeAddress = os.Getenv("MULTISIG_FEE_ADDRESS_PRIME")

		fee := uint64(0)
		fee, err = strconv.ParseUint(os.Getenv("MULTISIG_FEE_PRIME"), 10, 64)
		if err != nil {
			return
		}
		multisigFee = uint(fee)
		return
	} else if chainId == "vector" {
		multisigAddress = os.Getenv("MULTISIG_ADDRESS_VECTOR")
		multisigFeeAddress = os.Getenv("MULTISIG_FEE_ADDRESS_VECTOR")

		fee := uint64(0)
		fee, err = strconv.ParseUint(os.Getenv("MULTISIG_FEE_VECTOR"), 10, 64)
		if err != nil {
			return
		}
		multisigFee = uint(fee)
		return
	} else {
		err = errors.New("unsupported chainId, supported chainIds are prime and vector")
		return
	}

}
