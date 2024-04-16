package tx

import (
	"crypto/ed25519"
	"fmt"
)

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

// VKeyWitness - Witness for use with Shelley based transactions
type VKeyWitness struct {
	_         struct{} `cbor:",toarray"`
	VKey      []byte
	Signature []byte
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

// GetVerificationKeyFromSigningKey retrieves verification/public key from signing/private key
func GetVerificationKeyFromSigningKey(signingKey []byte) []byte {
	return ed25519.NewKeyFromSeed(signingKey).Public().(ed25519.PublicKey)
}

func SignMessage(signingKey, verificationKey, message []byte) (result []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error: %v", r)
		}
	}()

	privateKey := make([]byte, len(signingKey)+len(verificationKey))

	copy(privateKey, signingKey)
	copy(privateKey[32:], verificationKey)

	result = ed25519.Sign(privateKey, message)

	return
}
