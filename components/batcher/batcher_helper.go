package batcher

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/protocol"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
)

// Currently mocked
// UPDATETODO: Query data from contract
func getConfirmedTxs(destinationChain string) (receivers map[string]uint, err error) {
	receivers = map[string]uint{"addr_test1vpe3gtplyv5ygjnwnddyv0yc640hupqgkr2528xzf5nms7qalkkln": 1000000}
	return
}

func getBatchNonceId() (uint, error) {
	return 1, nil
}

// Currently mocked to get the UTXOs directly from chain
// UPDATETODO: Query data from contract
func getUTXOs(addressString string, amount uint, chainId string) (chosenUTXOs []*tx.TxInput, err error) {
	err = godotenv.Load()
	if err != nil {
		return
	}

	ogmios := node.NewOgmiosNode(os.Getenv("OGMIOS_NODE_ADDRESS_" + strings.ToUpper(chainId)))

	senderAddress, err := address.NewAddress(addressString)
	if err != nil {
		return
	}

	utxos, err := ogmios.UTXOs(senderAddress)
	if err != nil {
		return []*tx.TxInput{}, err
	}

	// Loop through utxos to find first input with enough tokens
	// If we don't have this UTXO we need to use more of them
	potentialFee, err := strconv.ParseUint(os.Getenv("POTENTIAL_FEE"), 10, 64)
	if err != nil {
		return
	}
	var amountSum = uint(0)
	var minUtxoValue = uint(1000000)

	for _, utxo := range utxos {
		if utxo.Amount == amount+uint(potentialFee) || utxo.Amount >= amount+minUtxoValue+uint(potentialFee) {
			println("here")
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

		if amountSum == amount+uint(potentialFee) || amountSum >= amount+minUtxoValue+uint(potentialFee) {
			break
		}
	}

	return
}

// Get protocol parameters
func getProtocolParameters(chainId string) (protocolParams protocol.Protocol, err error) {
	err = godotenv.Load()
	if err != nil {
		return
	}

	ogmios := node.NewOgmiosNode(os.Getenv("OGMIOS_NODE_ADDRESS_" + strings.ToUpper(chainId)))
	protocolParams, err = ogmios.ProtocolParameters()
	return
}

// Get slot number
func getSlotNumber(chainId string) (slot uint, err error) {
	err = godotenv.Load()
	if err != nil {
		return
	}

	ogmios := node.NewOgmiosNode(os.Getenv("OGMIOS_NODE_ADDRESS_" + strings.ToUpper(chainId)))
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
