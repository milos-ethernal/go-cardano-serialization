package tx

import (
	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/fees"
	"github.com/fivebinaries/go-cardano-serialization/protocol"
)

// TxBuilder - used to create, validate and sign transactions.
type TxBuilder struct {
	tx       *Tx
	xprvs    []bip32.XPrv
	protocol protocol.Protocol
}

// Sign adds a private key to create signature for witness
func (tb *TxBuilder) Sign(xprv bip32.XPrv) {
	tb.xprvs = append(tb.xprvs, xprv)
}

// Build creates hash of transaction, signs the hash using supplied witnesses and adds them to the transaction.
func (tb *TxBuilder) Build() (tx Tx, err error) {
	if tx.WitnessSet != nil {
		return *tb.tx, nil
	}

	hash, err := tb.tx.Hash()
	if err != nil {
		return tx, err
	}

	txKeys := []VKeyWitness{}
	for _, prv := range tb.xprvs {
		publicKey := prv.Public().PublicKey()
		signature := prv.Sign(hash[:])

		txKeys = append(txKeys, NewVKeyWitness(publicKey, signature[:]))
	}

	tb.tx.WitnessSet = NewTXWitnessSet(
		[]NativeScript{}, txKeys,
	)

	return *tb.tx, nil
}

// Tx returns a pointer to the transaction
func (tb *TxBuilder) Tx() (tx *Tx) {
	return tb.tx
}

// AddChangeIfNeeded calculates the excess change from UTXO inputs - outputs and adds it to the transaction body.
func (tb *TxBuilder) AddChangeIfNeeded(addr address.Address) {
	// change is amount in utxo minus outputs minus fee
	tb.tx.SetFee(tb.MinFee())
	totalI, totalO := tb.GetTotalInputOutputs()

	change := totalI - totalO - uint(tb.tx.Body.Fee)
	tb.tx.AddOutputs(
		NewTxOutput(
			addr,
			change,
		),
	)
}

// SetTTL sets the time to live for the transaction.
func (tb *TxBuilder) SetTTL(ttl uint32) {
	tb.tx.Body.TTL = ttl
}

func (tb TxBuilder) GetTotalInputOutputs() (inputs, outputs uint) {
	for _, inp := range tb.tx.Body.Inputs {
		inputs += inp.Amount
	}
	for _, out := range tb.tx.Body.Outputs {
		outputs += uint(out.Amount)
	}

	return
}

// MinFee calculates the minimum fee for the provided transaction.
func (tb TxBuilder) MinFee() (fee uint) {
	feeTx := Tx{
		Body: &TxBody{
			Inputs:  tb.tx.Body.Inputs,
			Outputs: tb.tx.Body.Outputs,
			Fee:     tb.tx.Body.Fee,
			TTL:     tb.tx.Body.TTL,
		},
		WitnessSet: tb.tx.WitnessSet,
		Valid:      true,
		Metadata:   tb.tx.Metadata,
	}
	feeTx.CalculateAuxiliaryDataHash()
	if len(feeTx.WitnessSet.Witnesses) == 0 {
		vWitness := NewVKeyWitness(
			make([]byte, 32),
			make([]byte, 64),
		)
		feeTx.WitnessSet.Witnesses = append(feeTx.WitnessSet.Witnesses, vWitness)
	}

	// Not realy sure about this part of the function
	// commented for further discusion

	// totalI, totalO := tb.GetTotalInputOutputs()

	// if totalI != (totalO) {
	// 	inner_addr, _ := address.NewAddress("addr_test1qqe6zztejhz5hq0xghlf72resflc4t2gmu9xjlf73x8dpf88d78zlt4rng3ccw8g5vvnkyrvt96mug06l5eskxh8rcjq2wyd63")

	// 	feeTx.Body.Outputs = append(feeTx.Body.Outputs, NewTxOutput(inner_addr, (totalI-totalO-200000)))

	// }

	lfee := fees.NewLinearFee(tb.protocol.TxFeePerByte, tb.protocol.TxFeeFixed)
	// The fee may have increased enough to increase the number of bytes, so do one more pass
	fee, _ = feeTx.Fee(lfee)
	feeTx.Body.Fee = uint64(fee)
	fee, _ = feeTx.Fee(lfee)

	return
}

// AddInputs adds inputs to the transaction body
func (tb *TxBuilder) AddInputs(inputs ...*TxInput) {
	tb.tx.AddInputs(inputs...)
}

// AddOutputs add outputs to the transaction body
func (tb *TxBuilder) AddOutputs(outputs ...*TxOutput) {
	tb.tx.AddOutputs(outputs...)
}

// NewTxBuilder returns pointer to a new TxBuilder.
func NewTxBuilder(pr protocol.Protocol, xprvs []bip32.XPrv) *TxBuilder {
	return &TxBuilder{
		tx:       NewTx(),
		xprvs:    xprvs,
		protocol: pr,
	}
}
