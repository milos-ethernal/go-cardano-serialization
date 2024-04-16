package node

import (
	"github.com/milos-ethernal/go-cardano-serialization/address"
	"github.com/milos-ethernal/go-cardano-serialization/protocol"
	"github.com/milos-ethernal/go-cardano-serialization/tx"
)

type Node interface {
	// UTXOs returns list of unspent transaction outputs
	UTXOs(address.Address) ([]tx.TxInput, error)

	// SubmitTx submits a cbor marshalled transaction to the cardano blockchain
	// using blockfrost or cardano-cli
	SubmitTx(tx.Tx) (string, error)

	// ProtocolParameters returns Protocol Parameters from the network
	ProtocolParameters() (protocol.Protocol, error)

	// QueryTip returns the tip of the network for use in tx building
	//
	// Using `query tip` on cardano-cli requires a synced local node
	QueryTip() (NetworkTip, error)
}
