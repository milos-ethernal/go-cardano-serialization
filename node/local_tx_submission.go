package node

import (
	"bytes"
	"io"
	"net/http"
)

func SubmitTx(transaction []byte) (int, string) {
	// Set up the URL for the cardano-submit-api
	url := "http://localhost:8090/api/submit/tx"

	// Create a new HTTP client
	client := &http.Client{}

	// Send the POST request with the CBOR-encoded transaction data
	resp, err := client.Post(url, "application/cbor", bytes.NewBuffer(transaction))
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(body)
}
