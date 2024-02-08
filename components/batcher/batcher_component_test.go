package batcher_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/components/batcher"
	"github.com/fivebinaries/go-cardano-serialization/internal/bech32/cbor"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/stretchr/testify/assert"
)

func TestBatcherBuildingBatchTxVector(t *testing.T) {
	tx, err := batcher.BuildBatchingTx("vector")
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, 2, len(tx.Body.Inputs))
	assert.GreaterOrEqual(t, 3, len(tx.Body.Outputs))
	assert.Equal(t, 2, len(tx.WitnessSet.Scripts))
	assert.Equal(t, 8, len(tx.WitnessSet.Witnesses))
	assert.Equal(t, uint(1), tx.AuxiliaryData.Metadata[1]["batch_nonce_id"])
	assert.NotEqual(t, uint(0), tx.Body.Fee)
}

func TestBatcherBuildingBatchTxPrime(t *testing.T) {
	tx, err := batcher.BuildBatchingTx("prime")
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, 2, len(tx.Body.Inputs))
	assert.GreaterOrEqual(t, 3, len(tx.Body.Outputs))
	assert.Equal(t, 2, len(tx.WitnessSet.Scripts))
	assert.Equal(t, 8, len(tx.WitnessSet.Witnesses))
	assert.Equal(t, uint(1), tx.AuxiliaryData.Metadata[1]["batch_nonce_id"])
	assert.NotEqual(t, uint(0), tx.Body.Fee)
}

func TestBatcherWitnessBatchingTx(t *testing.T) {
	transacion, err := batcher.BuildBatchingTx("vector")
	assert.NoError(t, err)

	witness, err := batcher.WitnessBatchingTx(*transacion, "6fbeede8a55f740152a307b6c3b3e6c787e34174c79cebde544504b2ee758a36")
	assert.NoError(t, err)

	seed, _ := hex.DecodeString("6fbeede8a55f740152a307b6c3b3e6c787e34174c79cebde544504b2ee758a36")
	pk, err := bip32.NewXPrv(seed)
	assert.NoError(t, err)

	hash, err := transacion.Hash()
	assert.NoError(t, err)

	publicKey := pk.Public().PublicKey()
	signature := pk.Sign(hash[:])
	testWitness := tx.NewVKeyWitness(publicKey, signature[:])

	assert.Equal(t, testWitness, witness)
}

func TestBatcherSubmitBatchingTxToNode(t *testing.T) {
	transaction, err := batcher.BuildBatchingTx("prime")
	assert.NoError(t, err)

	// Multisig address witnesses
	witness1, err := batcher.WitnessBatchingTx(*transaction, "6fbeede8a55f740152a307b6c3b3e6c787e34174c79cebde544504b2ee758a36")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness1)

	witness2, err := batcher.WitnessBatchingTx(*transaction, "d7faba3a4686fc6928b15a9834c3928ad2a9fe12b1409fdff741241c17fd0161")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness2)

	witness3, err := batcher.WitnessBatchingTx(*transaction, "38ab88c5cf12f5251a3b6ba3c5af2379c0c7ee26ed15de90b2c497ac5a6619a5")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness3)

	witness4, err := batcher.WitnessBatchingTx(*transaction, "1352739ad43e23d729b0d3f502804238f66cb65a5779f7082b9605c6dc24c664")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness4)

	// Multisig fee address witnesses
	witness5, err := batcher.WitnessBatchingTx(*transaction, "764c456698239d796e23029e381fb7b2b3f6fd84eb6c7898e34fa444a51eba00")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness5)

	witness6, err := batcher.WitnessBatchingTx(*transaction, "16f80fb0f3a1ea17585ebf18e78d2a0d306837ff51d0d63ee94e464bf4b88de6")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness6)

	witness7, err := batcher.WitnessBatchingTx(*transaction, "669725b0d4dede84ceb60adbdd7121f0fde3ffcdbc753afe9833fadc40caee30")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness7)

	witness8, err := batcher.WitnessBatchingTx(*transaction, "71343859f289a2d604db09e2d383271122a7760bb0d50d09c1e1028ce0181636")
	assert.NoError(t, err)
	assert.NotEmpty(t, witness8)

	transaction.WitnessSet.Witnesses = []tx.VKeyWitness{witness2, witness3, witness4, witness6, witness7, witness8}

	ogmios := node.NewOgmiosNode("http://localhost:1337")
	txHash, err := transaction.Bytes()
	assert.NoError(t, err)
	res, err := ogmios.SubmitTx(hex.EncodeToString(txHash[:]))
	assert.NoError(t, err)

	// Prints transaction hash if test PASS
	fmt.Println(res)
}

func TestSubmitBatchingTx(t *testing.T) {
	transacion, err := batcher.BuildBatchingTx("vector")
	assert.NoError(t, err)

	witness, err := batcher.WitnessBatchingTx(*transacion, "6fbeede8a55f740152a307b6c3b3e6c787e34174c79cebde544504b2ee758a36")
	assert.NoError(t, err)

	err = batcher.SubmitBatchingTx(*transacion, witness, "0")
	assert.NoError(t, err)
}

func TestAddressMarshaling(t *testing.T) {
	addr, err := address.NewAddress("addr_test1wrv5yn3vyx58zld5xahs4e0ezrcjyezldqjtch8cnyt92zcklzc25")
	assert.NoError(t, err)

	addrBytes, err := cbor.Marshal(addr.Bytes())
	assert.NoError(t, err)

	var addrFromBytes []byte
	err = cbor.Unmarshal(addrBytes, &addrFromBytes)
	assert.NoError(t, err)

	addrFromBytesAddr, err := address.NewAddressFromBytes(addrFromBytes)
	assert.NoError(t, err)

	assert.Equal(t, addr, addrFromBytesAddr)
}
