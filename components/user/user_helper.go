package user

import (
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
)

// Get protocol parameters
func getNodeUrl(chainId string) (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	return os.Getenv("OGMIOS_NODE_ADDRESS_" + strings.ToUpper(chainId)), nil
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

// Inteded for use in MVP version without wallet
func SignTransaction(transacion tx.Tx, seedString string) (tx.Tx, error) {
	seed, _ := hex.DecodeString(seedString)
	pk, err := bip32.NewXPrv(seed)
	if err != nil {
		return transacion, err
	}

	txHash, err := transacion.Hash()
	if err != nil {
		return transacion, err
	}

	publicKey := pk.Public().PublicKey()
	signature := pk.Sign(txHash[:])
	transacion.WitnessSet.Witnesses = []tx.VKeyWitness{tx.NewVKeyWitness(publicKey, signature[:])}

	return transacion, nil
}
