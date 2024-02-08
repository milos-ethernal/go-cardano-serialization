package relayer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fivebinaries/go-cardano-serialization/components"
	"github.com/fivebinaries/go-cardano-serialization/internal/bech32/cbor"
	"github.com/fivebinaries/go-cardano-serialization/node"
	"github.com/fivebinaries/go-cardano-serialization/tx"
	"github.com/joho/godotenv"
)

func getOgmiosNode(chainId string) (ogmiosNode node.OgmiosNode, err error) {
	// Load env variables
	err = godotenv.Load()
	if err != nil {
		return
	}

	ogmiosNode = node.NewOgmiosNode(os.Getenv(strings.ToUpper(chainId)))

	return
}

// UPDATETODO: Read data from bridge chain
// Mocked to get all data from the files locally
func getTransactionAndWitnesses() (transaction tx.Tx, witnesses []tx.VKeyWitness, err error) {
	// To submit valid tx in the current setup we need at least 6 witnesses
	// So for the testing purrpose we will read 6 files

	filePath := filepath.Join("/tmp", "tx_and_witness_")
	if err != nil {
		return
	}

	for i := 0; i < 6; i++ {
		tr, witness, errReadFile := readFile(filePath + fmt.Sprint(i))
		if err != nil {
			err = errReadFile
			return
		}

		transaction = tr
		witnesses = append(witnesses, witness)
	}

	return
}

// UPDATETODO: Remove after getTransactionAndWitnesses() is updated
func readFile(filename string) (transaction tx.Tx, witness tx.VKeyWitness, err error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	defer file.Close()

	// Create a buffer to read the file in chunks
	buffer := make([]byte, 1024) // Read 1024 bytes at a time

	// Create a slice to store the bytes read from the file
	var data []byte

	// Loop until the end of the file is reached
	for {
		// Read from the file into the buffer
		bytesRead, err1 := file.Read(buffer)
		if err != nil {
			// Check if the error is EOF (End of File)
			if err.Error() == "EOF" {
				break // Exit the loop when EOF is encountered
			}
			err = err1
			return
		}

		// If no bytes were read, we've reached the end of the file
		if bytesRead == 0 {
			break
		}

		// Process the bytes read from the buffer and append them to the data slice
		data = append(data, buffer[:bytesRead]...)
	}

	// Unmarshal byte arrays back to structs
	var readStruct components.Submit
	err = cbor.Unmarshal(data, &readStruct)
	if err != nil {
		return
	}

	// Print the read structs
	transaction = readStruct.Transaction
	witness = readStruct.Witness

	return
}
