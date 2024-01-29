package tx_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/network"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/protocol"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _  = runtime.Caller(0)
	packagepath = filepath.Dir(b)
)

var (
	// update   = flag.Bool("update", false, "update .golden files")
	generate = flag.Bool("gen", false, "generate .golden files")
)

type utxoIn struct {
	TxHash         string `json:"txHash"`
	TxIndex        uint   `json:"txIndex"`
	AmountLovelace uint   `json:"amountLovelace"`
}

type txDetails struct {
	ReceiverAddress string `json:"receiverAddress"`
	AmountLovelace  uint   `json:"amountLovelace"`
	ChangeAddress   string `json:"changeAddress"`
	SlotNo          uint   `json:"slotNo"`
	UtxoIn          utxoIn `json:"utxoIn"`
}

type txScenario struct {
	Description string
	GoldenFile  string
	AddrProc    func(*address.BaseAddress) address.Address
}

func createRootKey() bip32.XPrv {
	rootKey := bip32.FromBip39Entropy(
		[]byte{214, 64, 138, 69, 145, 210, 32, 51, 202, 45, 90, 151, 33, 194, 153, 176, 188, 94, 94, 186, 67, 118, 194, 227, 207, 157, 54, 49, 34, 12, 83, 93},
		[]byte{},
	)
	return rootKey
}

func harden(num uint) uint32 {
	return uint32(0x80000000 + num)
}

func generateBaseAddress(net *network.NetworkInfo) (addr *address.BaseAddress, utxoPrvKey bip32.XPrv, err error) {
	rootKey := createRootKey()
	accountKey := rootKey.Derive(harden(1852)).Derive(harden(1815)).Derive(harden(0))

	utxoPrvKey = accountKey.Derive(0).Derive(0)
	utxoPubKey := utxoPrvKey.Public()
	utxoPubKeyHash := utxoPubKey.PublicKey().Hash()

	stakeKey := accountKey.Derive(2).Derive(0).Public()
	stakeKeyHash := stakeKey.PublicKey().Hash()

	addr = address.NewBaseAddress(
		net,
		&address.StakeCredential{
			Kind:    address.KeyStakeCredentialType,
			Payload: utxoPubKeyHash[:],
		},
		&address.StakeCredential{
			Kind:    address.KeyStakeCredentialType,
			Payload: stakeKeyHash[:],
		})
	return
}

func getTxDetails(fp string) (txD txDetails) {
	data, err := readJson(fp)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(data, &txD); err != nil {
		log.Fatal("Err", err)
	}

	return

}

func readJson(fp string) (data []byte, err error) {
	file, err := os.Open(fp)
	if err != nil {
		return
	}

	defer file.Close()

	data, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}

	return
}

func TestBlockfrostAPI(t *testing.T) {
	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	blockfrostApi := node.NewBlockfrostClient(
		os.Getenv("BLOCKFROST_PROJECT_ID"),
		network.TestNet(),
	)

	pr, err := blockfrostApi.ProtocolParameters()
	if err != nil {
		panic(err)
	}

	assert.NotEqual(t, 0, pr.MaxTxSize)
}

func TestAuxiliaryData(t *testing.T) {
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
	pr, err := blockfrostApi.ProtocolParameters()
	if err != nil {
		panic(err)
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{},
	)

	builder.Tx().AuxiliaryData = tx.NewAuxiliaryData()
	builder.Tx().AuxiliaryData.AddMetadataElement("string", "value")

	assert.Equal(t, "value", builder.Tx().AuxiliaryData.Metadata[1]["string"])
}

func TestMultisigTx(t *testing.T) {
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
	pr, err := blockfrostApi.ProtocolParameters()
	if err != nil {
		panic(err)
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{},
	)

	// Multisig address
	sender, err := address.NewAddress("addr_test1wqklxqkgu755t8lxv6haj6aymqhzuljxc8wmpc546ulslks5tr7ya")
	if err != nil {
		panic(err)
	}

	receiver, err := address.NewAddress("addr_test1vz9zwl6tv8qgkzxz4ck7jqye5gdfmzujcmh8vwc4fdv68qgamk2jh")
	if err != nil {
		panic(err)
	}

	// Get the senders available UTXOs
	utxos, err := blockfrostApi.UTXOs(sender)
	if err != nil {
		panic(err)
	}

	// Send 1000000 lovelace or 1 ADA
	sendAmount := 1000000
	var firstMatchInput tx.TxInput

	// Loop through utxos to find first input with enough ADA
	for _, utxo := range utxos {
		minRequired := sendAmount + 200000
		if utxo.Amount >= uint(minRequired) {
			firstMatchInput = utxo
		}
	}

	// Add the transaction Input / UTXO
	builder.AddInputs(&firstMatchInput)

	// Add a transaction output with the receiver's address and amount of 1 ADA
	builder.AddOutputs(tx.NewTxOutput(
		receiver,
		uint(sendAmount),
	))

	// Query tip from a node on the network. This is to get the current slot
	// and compute TTL of transaction.
	tip, err := blockfrostApi.QueryTip()
	if err != nil {
		log.Fatal(err)
	}

	// Set TTL for 5 min into the future
	builder.SetTTL(uint32(tip.Slot) + uint32(300))

	// Create script of multisig address
	firstSignerKeyHash, _ := hex.DecodeString("d8f3f9ee291c253b7c12f4103f91f73026ec32690ad9bc99cc95f8f1")
	secondSignerKeyHash, _ := hex.DecodeString("86b45d41aee0a41bc3c099d3108f251b4318a28f883e19abefb618c8")
	thirdSignerKeyHash, _ := hex.DecodeString("159bf228e41bc1e2b5fd1f347627db28111848f6b044ab1dc8bf5f57")

	multisigScript := tx.NativeScript{
		Type:    3,
		KeyHash: []byte{},
		N:       2,
		Scripts: []tx.NativeScript{
			{
				Type:          0,
				KeyHash:       firstSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
			{
				Type:          0,
				KeyHash:       secondSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
			{
				Type:          0,
				KeyHash:       thirdSignerKeyHash,
				N:             0,
				Scripts:       []tx.NativeScript{},
				IntervalValue: 0,
			},
		},
		IntervalValue: 0,
	}

	// Add metadata
	builder.Tx().AuxiliaryData = tx.NewAuxiliaryData()
	builder.Tx().AuxiliaryData.AddMetadataElement("string", "test")

	// Calculate fee before signing
	// Fee is a part of TxBody
	// Using AddChangeIfNeeded() will change tx that is being signed
	// Without added witnesses it will calculate fee wrong

	// Set arbitrary value for witnesses
	for i := 0; i < int(multisigScript.N)+1; i++ {
		vWitness := tx.NewVKeyWitness(
			make([]byte, 32),
			make([]byte, 64),
		)
		builder.Tx().WitnessSet.Witnesses = append(builder.Tx().WitnessSet.Witnesses, vWitness)
	}

	// Set multisig NativeScript
	builder.Tx().WitnessSet.Scripts = []tx.NativeScript{multisigScript}

	// Set multisig change output to difference between input and output amounts for fee calculation
	// If 0 is set instead of difference between total input and output
	// the fee calculation will be lower and tx won't pass
	totalI, totalO := builder.GetTotalInputOutputs()
	builder.AddOutputs(tx.NewTxOutput(
		sender,
		uint(totalI-totalO),
	))

	// Calculate fee
	builder.Tx().SetFee(builder.MinFee())

	// Update multisig change to = input - fee - output
	change := totalI - totalO - uint(builder.Tx().Body.Fee)

	builder.Tx().Body.Outputs[len(builder.Tx().Body.Outputs)-1].Amount = change

	// Multisig signing

	privKeys := []bip32.XPrv{}

	seed, _ := hex.DecodeString("6fbeede8a55f740152a307b6c3b3e6c787e34174c79cebde544504b2ee758a36")
	pk, err := bip32.NewXPrv(seed)
	if err != nil {
		panic(err)
	}
	privKeys = append(privKeys, pk)

	seed, _ = hex.DecodeString("d7faba3a4686fc6928b15a9834c3928ad2a9fe12b1409fdff741241c17fd0161")
	pk, err = bip32.NewXPrv(seed)
	if err != nil {
		panic(err)
	}
	privKeys = append(privKeys, pk)

	seed, _ = hex.DecodeString("38ab88c5cf12f5251a3b6ba3c5af2379c0c7ee26ed15de90b2c497ac5a6619a5")
	pk, err = bip32.NewXPrv(seed)
	if err != nil {
		panic(err)
	}
	privKeys = append(privKeys, pk)

	hash, _ := builder.Tx().Hash()
	txKeys := []tx.VKeyWitness{}
	for _, prv := range privKeys {
		publicKey := prv.Public().PublicKey()
		signature := prv.Sign(hash[:])
		txKeys = append(txKeys, tx.NewVKeyWitness(publicKey, signature[:]))
	}

	builder.Tx().WitnessSet.Witnesses = txKeys

	txFinal := builder.Tx()

	transaction, _ := txFinal.Bytes()

	// Set up the URL for the cardano-submit-api
	url := "http://localhost:8090/api/submit/tx"

	// Create a new HTTP client
	client := &http.Client{}

	// Send the POST request with the CBOR-encoded transaction data
	resp, err := client.Post(url, "application/cbor", bytes.NewBuffer(transaction))
	if err != nil {
		log.Fatal("Error:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response:", err)
		return
	}

	// Print the response status code and body
	fmt.Println("Status Code:", resp.Status)
	fmt.Println("Response Body:", string(body))

	assert.Equal(t, 202, resp.StatusCode)
}

func TestTxBuilderRaw(t *testing.T) {
	createRootKey()

	packagepath = strings.Replace(packagepath, "\\", "/", -1)
	basepath := packagepath[:strings.LastIndex(packagepath, "/")]

	pr, err := protocol.LoadProtocol(filepath.Join(basepath, "testdata", "protocol", "protocol.json"))
	if err != nil {
		log.Fatal(err)
	}

	txD := getTxDetails(filepath.Join(basepath, "testdata", "transaction", "tx_builder", "json", "raw_tx.json"))

	txScenarios := []txScenario{
		{
			Description: "Transaction with base address marshalling",
			GoldenFile:  "raw_tx_base.golden",
			AddrProc:    func(addr *address.BaseAddress) address.Address { return addr },
		},
		{
			Description: "Transaction with enterprise address marshalling",
			GoldenFile:  "raw_tx_ent.golden",
			AddrProc:    func(addr *address.BaseAddress) address.Address { return addr.ToEnterprise() },
		},
	}

	for _, sc := range txScenarios {
		t.Run(sc.Description, func(t *testing.T) {
			builder := tx.NewTxBuilder(
				*pr,
				[]bip32.XPrv{},
			)
			addr, utxoPrv, err := generateBaseAddress(network.MainNet())
			if err != nil {
				log.Fatal(err)
			}

			builder.AddInputs(
				tx.NewTxInput(
					txD.UtxoIn.TxHash,
					uint16(txD.UtxoIn.TxIndex),
					txD.UtxoIn.AmountLovelace,
				),
			)

			builder.AddOutputs(
				tx.NewTxOutput(
					sc.AddrProc(addr),
					txD.AmountLovelace,
				),
			)

			changeAddr, err := address.NewAddress(txD.ChangeAddress)
			if err != nil {
				log.Fatal(err)
			}
			builder.SetTTL(uint32(txD.SlotNo))
			builder.AddChangeIfNeeded(changeAddr)

			builder.Sign(
				utxoPrv,
			)
			txFinal, err := builder.Build()
			if err != nil {
				t.Fatal(err)
			}

			txHex, err := txFinal.Hex()
			if err != nil {
				log.Fatal(err)
			}
			golden := filepath.Join(basepath, "testdata", "transaction", "tx_builder", "golden", sc.GoldenFile)
			assert.Equal(t, txHex, ReadOrGenerateGoldenFile(t, golden, txFinal))
		})
	}
}

func WriteGoldenFile(t *testing.T, path string, data []byte) {
	t.Helper()
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

func ReadOrGenerateGoldenFile(t *testing.T, path string, txF tx.Tx) string {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if *generate {
			txHex, err := txF.Hex()
			if err != nil {
				t.Fatal("golden-gen: Failed to hex encode transaction")
			}
			if err != nil {
				t.Fatal("golden-gen: Failed to hex encode transaction")
			}
			WriteGoldenFile(t, path, []byte(txHex))
			return txHex
		}
		t.Fatalf("golden-read: Missing golden file. Run `go test -args -gen` to generate it.")
	case err != nil:
		t.Fatal(err)
	}
	return string(b)
}
