package batcher

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
)

func BuildAndSubmitBatchingTx(destinationChainId string, receivers map[string]uint) (txHash string, err error) {
	if destinationChainId != "prime" && destinationChainId != "vector" {
		return "", errors.New("chainId not supported, supported chainIds are prime and vector")
	}

	if len(receivers) == 0 {
		return "", errors.New("receivers map cannot be empty")
	}

	// Calculate amount for UTXO
	amountSum := uint(0)
	var receiversUTXOs []tx.TxOutput
	var receiver address.Address
	for addressString, amount := range receivers {
		if amount < uint(1000000) {
			err = errors.New("receiver amount cannot be smaller than 1000000 tokens, " + addressString + ":" + fmt.Sprint(amount))
			return
		}
		amountSum += amount

		receiver, err = address.NewAddress(addressString)
		if err != nil {
			return
		}
		receiversUTXOs = append(receiversUTXOs, *tx.NewTxOutput(receiver, amount))
	}

	err = godotenv.Load()
	if err != nil {
		return
	}

	bridgeControlledAddressString := os.Getenv("BRIDGE_ADDRESS_" + strings.ToUpper(destinationChainId))

	// Get input UTXOs
	inputs, err := getUTXOs(bridgeControlledAddressString, amountSum, destinationChainId)
	if err != nil {
		return
	}

	// Generate batching transaction
	// Instantiate transaction builder
	pr, err := getProtocolParameters(destinationChainId)
	if err != nil {
		return
	}

	seed, _ := hex.DecodeString(os.Getenv("BRIDGE_ADDRESS_KEY_" + strings.ToUpper(destinationChainId)))
	pk, err := bip32.NewXPrv(seed)
	if err != nil {
		panic(err)
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{pk},
	)

	builder.AddInputs(inputs...)

	bridgeControlledAddress, err := address.NewAddress(bridgeControlledAddressString)
	if err != nil {
		return
	}

	// Add receiver outputs
	for _, utxo := range receiversUTXOs {
		builder.AddOutputs(&utxo)
	}

	// Query slot from a node on the network.
	// Slot is needed to compute TTL of transaction.
	slot, err := getSlotNumber(destinationChainId)
	if err != nil {
		return
	}

	// Set TTL for 5 min into the future
	builder.SetTTL(uint32(slot) + uint32(300))
	builder.Tx().AuxiliaryData = tx.NewAuxiliaryData()
	builder.AddChangeIfNeeded(bridgeControlledAddress)

	txFinal, err := builder.Build()
	if err != nil {
		return
	}

	ogmios := node.NewOgmiosNode(os.Getenv("OGMIOS_NODE_ADDRESS_" + strings.ToUpper(destinationChainId)))

	txBytes, err := txFinal.Bytes()
	if err != nil {
		return
	}

	txHash, err = ogmios.SubmitTx(hex.EncodeToString(txBytes[:]))

	return
}
