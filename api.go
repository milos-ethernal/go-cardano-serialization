package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fivebinaries/go-cardano-serialization/components/batcher"
	"github.com/fivebinaries/go-cardano-serialization/components/user"
)

type BatchingTxRequest struct {
	ChainID  string `json:"chainId"`
	RecvAddr string `json:"recv_addr"`
	Amount   uint64 `json:"amount"`
}

type BridgingTxRequest struct {
	PrivKey       string `json:"priv_key"`
	SenderAddress string `json:"sender_address"`
	RecvAddress   string `json:"recv_address"`
	Amount        uint64 `json:"amount"`
	ChainID       string `json:"chainId"`
}

func main() {
	http.HandleFunc("/createAndSignBatchingTx", createAndSignBatchingTx)
	http.HandleFunc("/createAndSignBridgingTx", createAndSignBridgingTx)
	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func createAndSignBatchingTx(w http.ResponseWriter, r *http.Request) {
	var req BatchingTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	txHash, err := batcher.BuildAndSubmitBatchingTx(req.ChainID, map[string]uint{req.RecvAddr: uint(req.Amount)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(txHash))
	fmt.Fprintf(w, "Successfully created and signed batching transaction\n")
}

func createAndSignBridgingTx(w http.ResponseWriter, r *http.Request) {
	var req BridgingTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	unsgignedTx, err := user.CreateBridgingTransaction(req.SenderAddress, req.ChainID, map[string]uint{req.RecvAddress: uint(req.Amount)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	txHash, err := user.SignAndSubmitTransaction(unsgignedTx, req.PrivKey, req.ChainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(txHash))
	fmt.Fprintf(w, "Successfully created and signed bridging transaction\n")
}
