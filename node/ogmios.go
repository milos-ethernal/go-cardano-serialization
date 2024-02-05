package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fivebinaries/go-cardano-serialization/address"
	"github.com/fivebinaries/go-cardano-serialization/protocol"
	"github.com/fivebinaries/go-cardano-serialization/tx"
)

func NewOgmiosNode(address string) ogmiosNode {
	return ogmiosNode{
		address: address,
	}
}

// Implement Node interface

// UTXOs queries the ogmios for Unspent Transaction Outputs belonging to an address.
func (o *ogmiosNode) UTXOs(address address.Address) (txIs []tx.TxInput, err error) {
	query := queryLedgerStateUtxo{
		Jsonrpc: "2.0",
		Method:  "queryLedgerState/utxo",
		Params: queryLedgerStateUtxoParams{
			Addresses: []string{address.String()},
		},
		Id: nil,
	}

	// Create a new HTTP client
	client := &http.Client{}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return
	}
	// Send the POST request with the CBOR-encoded transaction data
	resp, err := client.Post(o.address, "application/json", bytes.NewBuffer(queryBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var responseData queryLedgerStateUtxoResponse
	// Unmarshal the JSON into the struct
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return
	}

	for _, utxo := range responseData.Result {
		txIs = append(txIs, *tx.NewTxInput(utxo.Transaction.Id, uint16(utxo.Index), utxo.Value.Ada.Lovelace))
	}

	return
}

func (o *ogmiosNode) ProtocolParameters() (protocolParams protocol.Protocol, err error) {
	query := queryLedgerStateProtocolParameters{
		Jsonrpc: "2.0",
		Method:  "queryLedgerState/protocolParameters",
		Id:      nil,
	}

	// Create a new HTTP client
	client := &http.Client{}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return
	}
	// Send the POST request with the JSON-encoded query
	resp, err := client.Post(o.address, "application/json", bytes.NewBuffer(queryBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var responseData queryLedgerStateProtocolParametersResponse
	// Unmarshal the JSON into the struct
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return
	}

	protocolParams = protocol.Protocol{
		TxFeePerByte: responseData.Result.MinFeeCoefficient,
		TxFeeFixed:   responseData.Result.MinFeeConstant.Ada.Lovelace,
		MaxTxSize:    responseData.Result.MaxTransactionSize.Bytes,
		ProtocolVersion: protocol.ProtocolVersion{
			Major: uint8(responseData.Result.Version.Major),
			Minor: uint8(responseData.Result.Version.Minor)},
		MinUTXOValue: responseData.Result.MinUtxoDepositCoefficient,
	}

	return
}

// Define the function for querying tip
func (o *ogmiosNode) QueryTip() (tip NetworkTip, err error) {
	query := queryLedgerStateTip{
		Jsonrpc: "2.0",
		Method:  "queryLedgerState/tip",
		Id:      nil,
	}

	// Create a new HTTP client
	client := &http.Client{}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return
	}
	// Send the POST request with the JSON-encoded query
	resp, err := client.Post(o.address, "application/json", bytes.NewBuffer(queryBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Unmarshal the JSON into the struct
	var responseData queryLedgerStateTipResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return
	}

	// Repack the response into NetworkTip
	tip.Slot = responseData.Result.Slot
	// Block is not provided in the response, so it remains 0

	return
}

// Define the function for submitting a transaction
func (o *ogmiosNode) SubmitTx(txCborString string) (string, error) {
	// Create a new HTTP client
	client := &http.Client{}

	// Create the request body
	requestBody := submitTransaction{
		Jsonrpc: "2.0",
		Method:  "submitTransaction",
		Params: submitTransactionParams{
			Transaction: submitTransactionParamsTransaction{
				CBOR: txCborString,
			},
		},
		Id: nil,
	}

	// Marshal the request body
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Send the POST request with the JSON-encoded transaction
	resp, err := client.Post(o.address, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal the JSON into the struct
	var responseData submitTransactionResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return "", err
	}

	// Check for errors in the response
	if responseData.Error.Code != 0 {
		// Combine and return error information
		errorMessage := fmt.Sprintf("Code: %d, Message: %s, MissingScripts: %v",
			responseData.Error.Code, responseData.Error.Message, responseData.Error.Data.MissingScripts)
		return errorMessage, nil
	}

	// Return the transaction ID on success
	return responseData.Result.Transaction.ID, nil
}
