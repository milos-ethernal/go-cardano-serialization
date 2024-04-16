package tx

import (
	"reflect"

	"github.com/milos-ethernal/go-cardano-serialization/internal/bech32/cbor"
)

type MetadataElement interface{}

// Metadata represents the transaction metadata.
type Metadata map[uint]map[string]MetadataElement

// AuxiliaryData is the auxiliary data in the transaction.
type AuxiliaryData struct {
	Metadata           Metadata    `cbor:"0,keyasint,omitempty"`
	NativeScripts      interface{} `cbor:"1,keyasint,omitempty"`
	PlutusScripts      interface{} `cbor:"2,keyasint,omitempty"`
	PreferAlonzoFormat bool        `cbor:"2,keyasint,omitempty"`
}

func NewAuxiliaryData() *AuxiliaryData {
	return &AuxiliaryData{
		Metadata:           make(map[uint]map[string]MetadataElement),
		NativeScripts:      nil,
		PlutusScripts:      nil,
		PreferAlonzoFormat: false,
	}
}

func (d *AuxiliaryData) AddMetadataElement(key string, value MetadataElement) {
	if d.Metadata[1] == nil {
		d.Metadata[1] = make(map[string]MetadataElement)
	}

	d.Metadata[1][key] = value
}

func (d *AuxiliaryData) AddMetadataTransaction(address string, amount uint) {
	if d.Metadata[1] == nil {
		d.Metadata[1] = make(map[string]MetadataElement)
	}
	if d.Metadata[1]["transactions"] == nil {
		d.Metadata[1]["transactions"] = []map[string]uint{}
	}

	if transactionsSlice, ok := d.Metadata[1]["transactions"].([]map[string]uint); ok {
		transactionsSlice = append(transactionsSlice, map[string]uint{address: amount})
		d.Metadata[1]["transactions"] = transactionsSlice
	} else {
		panic("Wrong format: transactions field of metadata is expected to be []map[string]uint")
	}
}

// MarshalCBOR implements cbor.Marshaler
func (d *AuxiliaryData) MarshalCBOR() ([]byte, error) {
	type auxiliaryData AuxiliaryData

	// Register tag 259 for maps
	tags, err := d.tagSet(auxiliaryData{})
	if err != nil {
		return nil, err
	}

	em, err := cbor.CanonicalEncOptions().EncModeWithTags(tags)
	if err != nil {
		return nil, err
	}

	return em.Marshal(auxiliaryData(*d))
}

// UnmarshalCBOR implements cbor.Unmarshaler
func (d *AuxiliaryData) UnmarshalCBOR(data []byte) error {
	type auxiliaryData AuxiliaryData

	// Register tag 259 for maps
	tags, err := d.tagSet(auxiliaryData{})
	if err != nil {
		return err
	}

	dm, err := cbor.DecOptions{
		MapKeyByteString: cbor.MapKeyByteStringWrap,
	}.DecModeWithTags(tags)
	if err != nil {
		return err
	}

	var dd auxiliaryData
	if err := dm.Unmarshal(data, &dd); err != nil {
		return err
	}
	d.Metadata = dd.Metadata

	return nil
}

func (d *AuxiliaryData) tagSet(contentType interface{}) (cbor.TagSet, error) {
	tags := cbor.NewTagSet()
	err := tags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(contentType),
		259)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
