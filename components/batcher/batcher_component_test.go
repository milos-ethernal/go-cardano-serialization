package batcher_test

import (
	"fmt"
	"testing"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/components/batcher"
	"github.com/fivebinaries/go-cardano-serialization/internal/bech32/cbor"
	"github.com/stretchr/testify/assert"
)

func TestWrongChainId(t *testing.T) {
	txHash, err := batcher.BuildAndSubmitBatchingTx("asd", map[string]uint{})
	assert.Error(t, err)
	assert.Equal(t, "chainId not supported, supported chainIds are prime and vector", err.Error())
	assert.Equal(t, "", txHash)
}

func TestEmptyReceiversMap(t *testing.T) {
	txHash, err := batcher.BuildAndSubmitBatchingTx("prime", map[string]uint{})
	assert.Error(t, err)
	assert.Equal(t, "receivers map cannot be empty", err.Error())
	assert.Equal(t, "", txHash)
}

func TestReceiverLessThan1000000Tokens(t *testing.T) {
	txHash, err := batcher.BuildAndSubmitBatchingTx("prime", map[string]uint{"addr_test1vz9zwl6tv8qgkzxz4ck7jqye5gdfmzujcmh8vwc4fdv68qgamk2jh": 100000})
	assert.Error(t, err)
	assert.Equal(t, "receiver amount cannot be smaller than 1000000 tokens, addr_test1vz9zwl6tv8qgkzxz4ck7jqye5gdfmzujcmh8vwc4fdv68qgamk2jh:100000", err.Error())
	assert.Equal(t, "", txHash)
}

func TestSubmitTx(t *testing.T) {
	txHash, err := batcher.BuildAndSubmitBatchingTx("vector", map[string]uint{"addr_test1vq6zkfat4rlmj2nd2sylpjjg5qhcg9mk92wykaw4m2dp2rqneafvl": 1000000})
	assert.NoError(t, err)
	fmt.Println(txHash)
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
