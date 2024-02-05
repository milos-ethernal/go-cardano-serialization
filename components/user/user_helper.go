package user

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/protocol"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
)

// Get users available UTXOs
func getUsersUTXOs(address address.Address) ([]tx.TxInput, error) {
	ogmios := node.NewOgmiosNode("http://localhost:1337")
	return ogmios.UTXOs(address)
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
		log.Fatalf("err loading: %v", err)
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
