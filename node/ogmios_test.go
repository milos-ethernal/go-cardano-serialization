package node_test

import (
	"log"
	"os"
	"testing"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/network"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestQueryUTXO(t *testing.T) {
	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	blockfrostApi := node.NewBlockfrostClient(
		os.Getenv("BLOCKFROST_PROJECT_ID"),
		network.TestNet(),
	)

	sender, err := address.NewAddress("addr_test1wqklxqkgu755t8lxv6haj6aymqhzuljxc8wmpc546ulslks5tr7ya")
	if err != nil {
		panic(err)
	}

	resBlockfrost, err := blockfrostApi.UTXOs(sender)
	if err != nil {
		panic(err)
	}

	ogmios := node.NewOgmiosNode("http://localhost:1337")

	resOgmios, err := ogmios.UTXOs(sender)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.ElementsMatch(t, resBlockfrost, resOgmios)
}

func TestQueryProtocolParams(t *testing.T) {
	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	blockfrostApi := node.NewBlockfrostClient(
		os.Getenv("BLOCKFROST_PROJECT_ID"),
		network.TestNet(),
	)

	// Get protocol parameters
	prBlockfrost, err := blockfrostApi.ProtocolParameters()
	if err != nil {
		panic(err)
	}

	ogmios := node.NewOgmiosNode("http://localhost:1337")

	prOgmios, err := ogmios.ProtocolParameters()
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, prBlockfrost, prOgmios)
}

func TestQueryTip(t *testing.T) {
	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	blockfrostApi := node.NewBlockfrostClient(
		os.Getenv("BLOCKFROST_PROJECT_ID"),
		network.TestNet(),
	)

	// Get protocol parameters
	prBlockfrost, err := blockfrostApi.QueryTip()
	if err != nil {
		panic(err)
	}

	ogmios := node.NewOgmiosNode("http://localhost:1337")

	prOgmios, err := ogmios.QueryTip()
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, prBlockfrost.Slot, prOgmios.Slot)
}

func TestSubmitTx(t *testing.T) {
	// This test sends invalid tx with wrong TTL so the submission is supposed to fail
	// We test for expected error message
	ogmios := node.NewOgmiosNode("http://localhost:1337")

	msg, err := ogmios.SubmitTx("84a500818258201efc048ed50063d950f347da7c1a84a59a60b66442aff5e8904af7eba821a54f01018282581d702df302c8e7a9459fe666afd96ba4d82e2e7e46c1ddb0e295d73f0fda1a000f424082581d702df302c8e7a9459fe666afd96ba4d82e2e7e46c1ddb0e295d73f0fda1b000000025345ebb1021a00029f95031a026482da075820ae6e71b2ab4a70a68a67c9fad1fd4002171abc7740c953f525e8da70148e5a84a0f5d90103a100a101a267636861696e49646269646c7472616e73616374696f6e7381a1783f616464725f74657374317670746b65707a386c347a65303334373863767636707477647579676c676b366c636b7879746a74686b76766c75633364657766641a000f4240")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, "Code: 3118, Message: The transaction is outside of its validity interval. It was either submitted too early or too late. A transaction that has a lower validity bound can only be accepted by the ledger (and make it to the mempool) if the ledger's current slot is greater than the specified bound. The upper bound works similarly, as a time to live. The field 'data.currentSlot' contains the current slot as known of the ledger (this may be different from the current network slot if the ledger is still catching up). The field 'data.validityInterval' is a reminder of the validity interval provided with the transaction., MissingScripts: []", msg)
}
