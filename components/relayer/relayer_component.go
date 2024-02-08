package relayer

import (
	"encoding/hex"

	"github.com/fivebinaries/go-cardano-serialization/tx"
)

// Mocked for testing purposes
func SubmitTxToDestinationChain(chainId string) (submitedTxHash string, err error) {
	// Depeneding of chainId get neccessary parameters
	ogmios, err := getOgmiosNode(chainId)
	if err != nil {
		return
	}

	// Get transaction and witnesses

	var transaction tx.Tx
	var witnesses []tx.VKeyWitness

	transaction, witnesses, err = getTransactionAndWitnesses()
	if err != nil {
		return
	}

	// Combine transaction and witnesses
	transaction.WitnessSet.Witnesses = witnesses

	// Submit tx
	txBytes, err := transaction.Bytes()
	if err != nil {
		return
	}

	submitedTxHash, err = ogmios.SubmitTx(hex.EncodeToString(txBytes[:]))

	return
}
