package components

import "github.com/fivebinaries/go-cardano-serialization/tx"

type Submit struct {
	_           struct{} `cbor:",toarray"`
	Transaction tx.Tx
	Witness     tx.VKeyWitness
}
