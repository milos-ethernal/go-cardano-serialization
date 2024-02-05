package user_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/fivebinaries/go-cardano-serialization/bip32"
	user "github.com/fivebinaries/go-cardano-serialization/components"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/stretchr/testify/assert"
)

func TestCreateBridgingTransactionEmptySender(t *testing.T) {
	_, err := user.CreateBridgingTransaction("", "", make(map[string]uint))
	assert.Error(t, err)
	assert.Equal(t, "sender address cannot be empty string", err.Error())
}

func TestCreateBridgingTransactionWrongSenderAddressString(t *testing.T) {
	assert.Panics(t, func() { user.CreateBridgingTransaction("sender", "", make(map[string]uint)) }, "The code did not panic")
}

func TestCreateBridgingTransactionEmptyChainId(t *testing.T) {
	_, err := user.CreateBridgingTransaction("addr_test1vpe3gtplyv5ygjnwnddyv0yc640hupqgkr2528xzf5nms7qalkkln", "", make(map[string]uint))
	assert.Error(t, err)
	assert.Equal(t, "chainId cannot be empty string", err.Error())
}

func TestCreateBridgingTransactionUnsupportedChainId(t *testing.T) {
	_, err := user.CreateBridgingTransaction("addr_test1vpe3gtplyv5ygjnwnddyv0yc640hupqgkr2528xzf5nms7qalkkln", "id", make(map[string]uint))
	assert.Error(t, err)
	assert.Equal(t, "unsupported chainId, supported chainIds are prime and vector", err.Error())
}

func TestCreateBridgingTransactionEmptyMap(t *testing.T) {
	_, err := user.CreateBridgingTransaction("addr_test1vpe3gtplyv5ygjnwnddyv0yc640hupqgkr2528xzf5nms7qalkkln", "vector", make(map[string]uint))
	assert.Error(t, err)
	assert.Equal(t, "no receivers defined", err.Error())
}

func TestCreateBridgingTransactionNoEnoughFunds(t *testing.T) {
	receiversMap := make(map[string]uint)
	receiversMap["addr_test1vptkepz8l4ze03478cvv6ptwduyglgk6lckxytjthkvvluc3dewfd"] = 10000000001

	_, err := user.CreateBridgingTransaction("addr_test1vpe3gtplyv5ygjnwnddyv0yc640hupqgkr2528xzf5nms7qalkkln", "vector", receiversMap)
	assert.Error(t, err)
}

func TestCreateBridgingTransactionPass(t *testing.T) {
	receiversMap := make(map[string]uint)
	receiversMap["addr_test1vptkepz8l4ze03478cvv6ptwduyglgk6lckxytjthkvvluc3dewfd"] = 1000000

	output, err := user.CreateBridgingTransaction("addr_test1vpe3gtplyv5ygjnwnddyv0yc640hupqgkr2528xzf5nms7qalkkln", "vector", receiversMap)
	assert.NoError(t, err)

	assert.NotEqual(t, 0, len(output.Body.Inputs))

	multisigAddressString, multisigFeeAddressString, multisigFee, err := user.GetChainData("vector")
	assert.NoError(t, err)

	// Multisig address
	assert.Equal(t, multisigAddressString, output.Body.Outputs[0].Address.String())
	assert.Equal(t, uint(1000000)+multisigFee, output.Body.Outputs[0].Amount)

	assert.NotEqual(t, 0, output.Body.Fee)
	assert.NotEqual(t, 0, output.Body.TTL)

	assert.Equal(t, 0, len(output.WitnessSet.Witnesses))

	assert.Equal(t, "vector", output.AuxiliaryData.Metadata[1]["chainId"])

	receiver := map[string]interface{}{"address": "addr_test1vptkepz8l4ze03478cvv6ptwduyglgk6lckxytjthkvvluc3dewfd", "amount": uint(1000000)}
	feePayer := map[string]interface{}{"address": multisigFeeAddressString, "amount": multisigFee}
	metadataMap := []map[string]interface{}{receiver, feePayer}

	assert.Equal(t, metadataMap, output.AuxiliaryData.Metadata[1]["transactions"])
}

func TestCreateBridgingTransactionSubmit(t *testing.T) {
	receiversMap := make(map[string]uint)
	receiversMap["addr_test1vptkepz8l4ze03478cvv6ptwduyglgk6lckxytjthkvvluc3dewfd"] = 1000000

	output, err := user.CreateBridgingTransaction("addr_test1vpe3gtplyv5ygjnwnddyv0yc640hupqgkr2528xzf5nms7qalkkln", "vector", receiversMap)
	assert.NoError(t, err)

	// Sign the unsigned transaction
	seed, _ := hex.DecodeString("085de0735c76409f64a704e05eafdccd49f733a1dffea5e5bd514c6904179e948")
	pk, err := bip32.NewXPrv(seed)
	assert.NoError(t, err)

	txHash, err := output.Hash()
	assert.NoError(t, err)
	publicKey := pk.Public().PublicKey()
	signature := pk.Sign(txHash[:])
	output.WitnessSet.Witnesses = []tx.VKeyWitness{tx.NewVKeyWitness(publicKey, signature[:])}

	txHex, err := output.Hex()
	assert.NoError(t, err)

	ogmiosNode := node.NewOgmiosNode("http://localhost:1337")
	res, err := ogmiosNode.SubmitTx(txHex)
	assert.NoError(t, err)

	assert.NotEqual(t, "", res)
	fmt.Println(res)
}
