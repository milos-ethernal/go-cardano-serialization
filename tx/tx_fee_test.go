package tx_test

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/bip32"
	"github.com/fivebinaries/go-cardano-serialization/network"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestMultisigTxFee(t *testing.T) {
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
	// But don't set it so we can check node error output with our calculation
	calculatedFee := builder.MinFee()

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

	// Submit tx to local cardano-submit-api
	statusCode, msg := node.SubmitTx(transaction)

	nonNumberRegex := regexp.MustCompile(`[^0-9 ]+`)
	nodeFeeErrorVal := nonNumberRegex.ReplaceAllString(strings.Split(string(msg), "Coin")[1], "")

	nodeFee, err := strconv.ParseUint(strings.TrimSpace(nodeFeeErrorVal), 10, 32)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Calculated Fee = ", calculatedFee)
	fmt.Println("Node Error Fee = ", nodeFee)
	if nodeFee >= uint64(calculatedFee) {
		fmt.Println("Difference = ", nodeFee-uint64(calculatedFee))
	} else {
		fmt.Println("Difference = ", uint64(calculatedFee)-nodeFee)
	}

	assert.Equal(t, 400, statusCode)
	assert.Equal(t, calculatedFee, nodeFee)
}

func TestSimpleTxFee(t *testing.T) {
	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	cli := node.NewBlockfrostClient(
		os.Getenv("BLOCKFROST_PROJECT_ID"),
		network.TestNet(),
	)

	pr, err := cli.ProtocolParameters()
	if err != nil {
		panic(err)
	}

	seed, _ := hex.DecodeString("085de0735c76409f64a704e05eafdccd49f733a1dffea5e5bd514c6904179e948")
	pk, err := bip32.NewXPrv(seed)
	if err != nil {
		panic(err)
	}

	sender, err := address.NewAddress("addr_test1vpe3gtplyv5ygjnwnddyv0yc640hupqgkr2528xzf5nms7qalkkln")
	if err != nil {
		panic(err)
	}

	receiver, err := address.NewAddress("addr_test1vptkepz8l4ze03478cvv6ptwduyglgk6lckxytjthkvvluc3dewfd")
	if err != nil {
		panic(err)
	}

	// Get the senders available UTXOs
	utxos, err := cli.UTXOs(sender)
	if err != nil {
		panic(err)
	}

	builder := tx.NewTxBuilder(
		pr,
		[]bip32.XPrv{pk},
	)

	// Send 1000000 lovelace or 1 ADA
	sendAmount := 1000000
	var firstMatchInput tx.TxInput

	// Loop through utxos to find first input with enough ADA
	for _, utxo := range utxos {
		minRequired := sendAmount + 1000000 + 200000
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
	tip, err := cli.QueryTip()
	if err != nil {
		log.Fatal(err)
	}

	// Set TTL for 5 min into the future
	builder.SetTTL(uint32(tip.Slot) + uint32(300))

	// Route back the change to the source address
	// This is equivalent to adding an output with the source address and change amount
	builder.AddChangeIfNeeded(sender)

	// Save calculated fee and set it to 0
	// so we can check node error output with our calculation
	calculatedFee := builder.Tx().Body.Fee
	builder.Tx().SetFee(0)

	// Build loops through the witness private keys and signs the transaction body hash
	txFinal, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	transaction, _ := txFinal.Bytes()

	// Submit tx to local cardano-submit-api
	statusCode, msg := node.SubmitTx(transaction)

	nonNumberRegex := regexp.MustCompile(`[^0-9 ]+`)
	nodeFeeErrorVal := nonNumberRegex.ReplaceAllString(strings.Split(string(msg), "Coin")[3], "")

	nodeFee, err := strconv.ParseUint(strings.TrimSpace(nodeFeeErrorVal), 10, 32)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Calculated Fee = ", calculatedFee)
	fmt.Println("Node Error Fee = ", nodeFee)
	if nodeFee >= calculatedFee {
		fmt.Println("Difference = ", nodeFee-calculatedFee)
	} else {
		fmt.Println("Difference = ", calculatedFee-nodeFee)
	}

	assert.Equal(t, 400, statusCode)
	assert.Equal(t, calculatedFee, nodeFee)
}
