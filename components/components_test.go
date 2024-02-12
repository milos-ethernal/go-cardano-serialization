package components_test

import (
	"fmt"
	"testing"

	"github.com/fivebinaries/go-cardano-serialization/components/batcher"
	"github.com/fivebinaries/go-cardano-serialization/components/relayer"
	"github.com/fivebinaries/go-cardano-serialization/components/txhelper"
	"github.com/stretchr/testify/assert"
)

func TestBatcherAndRelayerComponents(t *testing.T) {
	transaction, err := batcher.BuildBatchingTx("prime")
	assert.NoError(t, err)

	// Multisig address witnesses
	witness2, err := txhelper.CreateWitness(transaction, "d7faba3a4686fc6928b15a9834c3928ad2a9fe12b1409fdff741241c17fd0161")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness2)
	batcher.SubmitBatchingTx(*transaction, witness2, "0")

	witness3, err := txhelper.CreateWitness(transaction, "38ab88c5cf12f5251a3b6ba3c5af2379c0c7ee26ed15de90b2c497ac5a6619a5")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness3)
	batcher.SubmitBatchingTx(*transaction, witness3, "1")

	witness4, err := txhelper.CreateWitness(transaction, "1352739ad43e23d729b0d3f502804238f66cb65a5779f7082b9605c6dc24c664")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness4)
	batcher.SubmitBatchingTx(*transaction, witness4, "2")

	// Multisig fee address witnesses
	witness6, err := txhelper.CreateWitness(transaction, "16f80fb0f3a1ea17585ebf18e78d2a0d306837ff51d0d63ee94e464bf4b88de6")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness6)
	batcher.SubmitBatchingTx(*transaction, witness6, "3")

	witness7, err := txhelper.CreateWitness(transaction, "669725b0d4dede84ceb60adbdd7121f0fde3ffcdbc753afe9833fadc40caee30")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness7)
	batcher.SubmitBatchingTx(*transaction, witness7, "4")

	witness8, err := txhelper.CreateWitness(transaction, "71343859f289a2d604db09e2d383271122a7760bb0d50d09c1e1028ce0181636")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness8)
	batcher.SubmitBatchingTx(*transaction, witness8, "5")

	txHash, err := relayer.SubmitTxToDestinationChain("prime")
	assert.NoError(t, err)
	assert.NotEqual(t, "", txHash)

	// Print tx hash
	fmt.Println(txHash)
}
