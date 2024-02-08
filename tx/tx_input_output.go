package tx

import (
	"encoding/hex"
	"errors"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fxamacker/cbor/v2"
)

type TxInput struct {
	cbor.Marshaler

	TxHash []byte
	Index  uint16
	Amount uint
}

// NewTxInput creates and returns a *TxInput from Transaction Hash(Hex Encoded), Transaction Index and Amount.
func NewTxInput(txHash string, txIx uint16, amount uint) *TxInput {
	hash, _ := hex.DecodeString(txHash)

	return &TxInput{
		TxHash: hash,
		Index:  txIx,
		Amount: amount,
	}
}

func (txI *TxInput) MarshalCBOR() ([]byte, error) {
	type arrayInput struct {
		_      struct{} `cbor:",toarray"`
		TxHash []byte
		Index  uint16
	}
	input := arrayInput{
		TxHash: txI.TxHash,
		Index:  txI.Index,
	}
	return cbor.Marshal(input)
}

// UnmarshalCBOR implements the cbor.Unmarshaler interface for TxInput.
func (txI *TxInput) UnmarshalCBOR(data []byte) error {
	// Define a struct to decode the CBOR data into
	var input struct {
		_      struct{} `cbor:",toarray"`
		TxHash []byte
		Index  uint16
	}

	// Unmarshal the CBOR data into the input struct
	if err := cbor.Unmarshal(data, &input); err != nil {
		return err
	}

	// Set values to the TxInput struct
	txI.TxHash = input.TxHash
	txI.Index = input.Index

	return nil
}

type TxOutput struct {
	_       struct{} `cbor:",toarray"`
	Address address.Address
	Amount  uint
}

// MarshalCBOR implements the cbor.Marshaler interface for TxOutput.
func (t *TxOutput) MarshalCBOR() ([]byte, error) {
	type cborOutput struct {
		_       struct{} `cbor:",toarray"`
		Address []byte
		Amount  uint
	}
	output := cborOutput{
		Address: t.Address.Bytes(), // Assuming Bytes() method returns byte slice representation of the address
		Amount:  t.Amount,
	}
	return cbor.Marshal(output)
}

func (t *TxOutput) UnmarshalCBOR(data []byte) error {
	// Decode CBOR data into a slice of interfaces
	var cborData []interface{}
	if err := cbor.Unmarshal(data, &cborData); err != nil {
		return err
	}

	// Ensure the CBOR data has at least 2 elements (address and amount)
	if len(cborData) < 2 {
		return errors.New("invalid CBOR data for TxOutput")
	}

	// Unmarshal the address
	addressBytes, ok := cborData[0].([]byte)
	if !ok {
		return errors.New("invalid address bytes in CBOR data for TxOutput")
	}

	address, err := address.NewAddressFromBytes(addressBytes)
	if err != nil {
		return err
	}
	t.Address = address

	// Unmarshal the amount
	amount, ok := cborData[1].(uint64)
	if !ok {
		return errors.New("invalid amount in CBOR data for TxOutput")
	}
	t.Amount = uint(amount)

	return nil
}

func NewTxOutput(addr address.Address, amount uint) *TxOutput {
	return &TxOutput{
		Address: addr,
		Amount:  amount,
	}
}
