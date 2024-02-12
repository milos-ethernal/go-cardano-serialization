package user

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/protocol"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
)

// Get users available UTXOs
func getUsersUTXOs(chainId string, address address.Address, amount uint, potentialFee uint) (chosenUTXOs []*tx.TxInput, err error) {
	err = godotenv.Load()
	if err != nil {
		return
	}
	ogmios := node.NewOgmiosNode(os.Getenv("OGMIOS_NODE_ADDRESS_" + strings.ToUpper(chainId)))

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
func getProtocolParameters(chainId string) (protocolParams protocol.Protocol, err error) {
	err = godotenv.Load()
	if err != nil {
		return
	}
	ogmios := node.NewOgmiosNode(os.Getenv("OGMIOS_NODE_ADDRESS_" + strings.ToUpper(chainId)))
	return ogmios.ProtocolParameters()
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
