package tx

import "github.com/fxamacker/cbor/v2"

type WitnessSet struct {
	Witnesses []VKeyWitness  `cbor:"0,keyasint,omitempty"`
	Scripts   []NativeScript `cbor:"1,keyasint,omitempty"`
}

// NewTXWitness returns a pointer to a Witness created from VKeyWitnesses.
func NewTXWitnessSet(scripts []NativeScript, witnesses []VKeyWitness) *WitnessSet {
	return &WitnessSet{
		Witnesses: witnesses,
		Scripts:   scripts,
	}
}

// Bytes returns a slice of cbor Marshalled bytes.
func (ws *WitnessSet) Bytes() ([]byte, error) {
	bytes, err := cbor.Marshal(ws)
	return bytes, err
}

// VKeyWitness - Witness for use with Shelley based transactions
type VKeyWitness struct {
	_         struct{} `cbor:",toarray"`
	VKey      []byte
	Signature []byte
}

// Bytes returns a slice of cbor Marshalled bytes.
func (v *VKeyWitness) Bytes() ([]byte, error) {
	bytes, err := cbor.Marshal(v)
	return bytes, err
}

// NewVKeyWitness creates a Witness for Shelley Based transactions from a verification key and transaction signature.
func NewVKeyWitness(vkey, signature []byte) VKeyWitness {
	return VKeyWitness{
		VKey: vkey, Signature: signature,
	}
}

// BootstrapWitness for use with Byron/Legacy based transactions
type BootstrapWitness struct {
	_          struct{} `cbor:",toarray"`
	VKey       []byte
	Signature  []byte
	ChainCode  []byte
	Attributes []byte
}
