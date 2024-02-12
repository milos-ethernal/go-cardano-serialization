package txhelper

import (
	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/protocol"
	"github.com/fivebinaries/go-cardano-serialization/tx"
)

const defaultNodeUrl = "http://localhost:1337"

type TxParamsHelper interface {
	GetProtocolParameters() (protocol.Protocol, error)
	GetUTXOs(addressString string) ([]tx.TxInput, error)
	GetSlotNumber() (uint, error)
}

type TxParamsHelperImpl struct {
	nodeUrl string
}

var _ TxParamsHelper = (*TxParamsHelperImpl)(nil)

func NewTxParamsHelper(nodeUrl string) *TxParamsHelperImpl {
	if len(nodeUrl) == 0 {
		nodeUrl = defaultNodeUrl
	}

	return &TxParamsHelperImpl{
		nodeUrl: nodeUrl,
	}
}

// Get protocol parameters
func (th TxParamsHelperImpl) GetProtocolParameters() (protocol.Protocol, error) {
	ogmios := node.NewOgmiosNode(th.nodeUrl)
	return ogmios.ProtocolParameters()
}

// Get all available Utxos for address
func (th TxParamsHelperImpl) GetUTXOs(addressString string) ([]tx.TxInput, error) {
	ogmios := node.NewOgmiosNode(th.nodeUrl)

	senderAddress, err := address.NewAddress(addressString)
	if err != nil {
		return nil, err
	}

	utxos, err := ogmios.UTXOs(senderAddress)
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

// Get slot number
func (th TxParamsHelperImpl) GetSlotNumber() (uint, error) {
	ogmios := node.NewOgmiosNode(th.nodeUrl)
	tip, err := ogmios.QueryTip()
	if err != nil {
		return 0, err
	}

	return tip.Slot, nil
}
