package txhelper

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/tx"
)

const (
	MinUtxoValue = uint(1000000)
)

var (
	errReceiverAmountToLow = errors.New("receivers amount cannot be under 1000000 tokens")
)

func CreateUtxos(receivers map[string]uint, checkMin bool) ([]*tx.TxOutput, error) {
	// Calculate amount for UTXO
	receiversUTXOs := make([]*tx.TxOutput, 0, len(receivers))

	for addr, amount := range receivers {
		if checkMin && amount < MinUtxoValue {
			return nil, errReceiverAmountToLow
		}

		receiver, err := address.NewAddress(addr)
		if err != nil {
			return nil, err
		}

		receiversUTXOs = append(receiversUTXOs, tx.NewTxOutput(receiver, amount))
	}

	return receiversUTXOs, nil
}

func CreateSignersScript(signersKeyHashes [][]byte) tx.NativeScript {
	scripts := make([]tx.NativeScript, len(signersKeyHashes))

	for i, keyHash := range signersKeyHashes {
		scripts[i] = tx.NativeScript{
			Type:          0,
			KeyHash:       keyHash,
			N:             0,
			Scripts:       []tx.NativeScript{},
			IntervalValue: 0,
		}
	}

	return tx.NativeScript{
		Type:          3,
		KeyHash:       []byte{},
		N:             3,
		Scripts:       scripts,
		IntervalValue: 0,
	}
}

func GetUTXOsForAmount(utxos []tx.TxInput, amount uint, potentialFee uint) ([]*tx.TxInput, error) {
	// Loop through utxos to find first input with enough tokens
	// If we don't have this UTXO we need to use more of them
	var (
		amountSum   = uint(0)
		chosenUTXOs []*tx.TxInput
		desired     = amount + MinUtxoValue + potentialFee
	)

	for _, utxo := range utxos {
		if utxo.Amount >= desired {
			return []*tx.TxInput{&utxo}, nil
		}

		amountSum += utxo.Amount
		chosenUTXOs = append(chosenUTXOs, &utxo)

		if amountSum >= desired {
			return chosenUTXOs, nil
		}
	}

	return nil, fmt.Errorf("no enough available funds for generating transaction: %d available, %d required", amountSum, desired)
}

func GetTxInputsSum(utxos []*tx.TxInput) (sum uint) {
	for _, x := range utxos {
		sum += x.Amount
	}

	return sum
}

func GetTxOutputsSum(utxos []*tx.TxOutput) (sum uint) {
	for _, x := range utxos {
		sum += x.Amount
	}

	return sum
}

func GetSumFromRecipients(receiversAndAmounts map[string]uint) (uint, error) {
	sum := uint(0)

	for receiverAddress, amount := range receiversAndAmounts {
		if _, err := address.NewAddress(receiverAddress); err != nil {
			return 0, err
		}

		if amount < MinUtxoValue {
			return 0, errors.New("receivers amount cannot be under 1000000 tokens")
		}

		sum += amount
	}

	return sum, nil
}

func GetDummyWitnesses(cnt int) []tx.VKeyWitness {
	res := make([]tx.VKeyWitness, cnt)

	for i := 0; i < cnt; i++ {
		res[i] = tx.NewVKeyWitness(
			make([]byte, 32),
			make([]byte, 64),
		)
	}

	return res
}

// Create Witness
func CreateWitness(transaction *tx.Tx, pkSeed string) (tx.VKeyWitness, error) {
	seed, err := hex.DecodeString(pkSeed)
	if err != nil {
		return tx.VKeyWitness{}, err
	}

	pk, err := bip32.NewXPrv(seed)
	if err != nil {
		return tx.VKeyWitness{}, err
	}

	hash, err := transaction.Hash()
	if err != nil {
		return tx.VKeyWitness{}, err
	}

	publicKey := pk.Public().PublicKey()
	signature := pk.Sign(hash[:])

	return tx.NewVKeyWitness(publicKey, signature[:]), nil
}
