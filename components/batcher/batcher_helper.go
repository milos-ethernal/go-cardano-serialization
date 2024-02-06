package batcher

import (
	"errors"
	"os"
	"strconv"

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
func getUTXOs(addressString string, amount uint) (chosenUTXOs []tx.TxInput, err error) {
	ogmios := node.NewOgmiosNode("http://localhost:1337")

	senderAddress, err := address.NewAddress(addressString)
	if err != nil {
		return
	}

	utxos, err := ogmios.UTXOs(senderAddress)
	if err != nil {
		return []tx.TxInput{}, err
	}

	var firstMatchInput = tx.TxInput{
		Marshaler: nil,
		TxHash:    []byte{},
		Index:     0,
		Amount:    0,
	}

	// Loop through utxos to find first input with enough tokens
	// If we don't have this UTXO we need to use more of them
	var amountSum = uint(0)

	for _, utxo := range utxos {
		if utxo.Amount >= amount {
			firstMatchInput = utxo
			break
		}

		amountSum += utxo.Amount
		chosenUTXOs = append(chosenUTXOs, utxo)

		if amountSum >= amount {
			break
		}
	}

	// Check if address have sufficent amount for transaction
	if firstMatchInput.Amount != 0 {
		chosenUTXOs = []tx.TxInput{firstMatchInput}
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
