package e2e

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fivebinaries/go-cardano-serialization/e2e/cardanofw"
	"github.com/stretchr/testify/assert"
)

// Download Cardano executables from https://github.com/IntersectMBO/cardano-node/releases/tag/8.1.2 and unpack tar.gz file
// Add directory where unpacked files are located to the $PATH (in example bellow `~/Apps/cardano`)
// eq add line `export PATH=$PATH:~/Apps/cardano` to  `~/.bashrc`
func TestE2E_CardanoTwoClustersBasic(t *testing.T) {
	t.Parallel()
	const (
		clusterCnt = 2
	)

	var (
		errors [clusterCnt]error
		wg     sync.WaitGroup
		// baseLogsDir string = path.Join("../..", fmt.Sprintf("e2e-logs-cardano-%d", time.Now().Unix()), t.Name())
		baseLogsDir string = path.Join("../..", "e2e-logs-cardano")
	)

	for i := 0; i < clusterCnt; i++ {
		id := i
		wg.Add(1)

		go func() {
			defer wg.Done()

			logsDir := fmt.Sprintf("%s/%d", baseLogsDir, id)
			if err := cardanofw.CreateDirSafe(logsDir, 0750); err != nil {
				errors[id] = err

				return
			}

			cluster, err := cardanofw.NewCardanoTestCluster(t,
				cardanofw.WithNodesCount(3),
				cardanofw.WithStartTimeDelay(time.Second*5),
				cardanofw.WithPort(10000*(id+1)+3001),
				cardanofw.WithLogsDir(logsDir),
				cardanofw.WithNetworkMagic(42+(id+1)*100))
			if err != nil {
				errors[id] = err

				return
			}

			defer cluster.Stop()

			t.Log("Waiting for sockets to be ready", "id", id+1, "sockets", strings.Join(cluster.GetSockets(), ", "))
			if errors[id] = cluster.WaitForReady(time.Second * 100); errors[id] != nil {
				return
			}

			t.Log("Waiting for blocks", "id", id+1)

			if id == 0 {
				err = cluster.SetEnvVariables("PRIME")
				if err != nil {
					return
				}
			} else if id == 1 {
				err = cluster.SetEnvVariables("VECTOR")
				if err != nil {
					return
				}
			}

			// Blocks WaitForBlockWithState from stoping when desired number of blocks is reached
			if errors[id] = cluster.StartOgmiosOnNode(uint((id+1)*10000 + 3001)); errors[id] != nil {
				return
			}

			errors[id] = cluster.WaitForBlockWithState(300, time.Hour*1)
		}()
	}

	wg.Wait()

	for i := 0; i < clusterCnt; i++ {
		assert.NoError(t, errors[i])
	}
}

// func TestE2E_BridgingTransactions(t *testing.T) {
// 	t.Parallel()
// 	time.Sleep(time.Second * 245)

// 	var baseLogsDir string = path.Join("../..", fmt.Sprintf("e2e-logs-result-%d", time.Now().Unix()), t.Name())
// 	if err := cardanofw.CreateDirSafe(baseLogsDir, 0750); err != nil {
// 		assert.NoError(t, err)
// 		return
// 	}
// 	filename := fmt.Sprintf("%s/bridging_results.txt", baseLogsDir)
// 	err := os.WriteFile(filename, []byte{}, 0644)
// 	assert.NoError(t, err)

// 	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
// 	assert.NoError(t, err)

// 	// PRIME -> VECTOR
// 	// 1st STEP: (initiated by user)
// 	// PRIME_SENDER -> PRIME_MULTISIG
// 	// Define VECTOR_SENDER as a receiver of bridged funds
// 	type RequestBody1 struct {
// 		PrivKey       string `json:"priv_key"`
// 		SenderAddress string `json:"sender_address"`
// 		RecvAddress   string `json:"recv_address"`
// 		Amount        int    `json:"amount"`
// 		ChainID       string `json:"chainId"`
// 	}

// 	// Create the request body
// 	requestBody1 := RequestBody1{
// 		PrivKey:       os.Getenv("SENDER_KEY_PRIME"),
// 		SenderAddress: os.Getenv("SENDER_ADDRESS_PRIME"),
// 		RecvAddress:   os.Getenv("SENDER_ADDRESS_VECTOR"),
// 		Amount:        1000000,
// 		ChainID:       "prime",
// 	}

// 	_, err = f.WriteString(fmt.Sprintf("%s\n", os.Getenv("SENDER_KEY_PRIME")))
// 	assert.NoError(t, err)
// 	_, err = f.WriteString(fmt.Sprintf("%s\n", os.Getenv("SENDER_ADDRESS_PRIME")))
// 	assert.NoError(t, err)

// 	// Marshal the request body to JSON
// 	requestBodyBytes, err := json.Marshal(requestBody1)
// 	assert.NoError(t, err)

// 	// Send HTTP POST request
// 	resp, err := http.Post("http://localhost:8000/api/createAndSignBridgingTx", "application/json", bytes.NewBuffer(requestBodyBytes))
// 	assert.NoError(t, err)
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err := io.ReadAll(resp.Body)
// 	assert.NoError(t, err)

// 	res := fmt.Sprintf("Succesfully submited user tx to PRIME %s\n", string(body))
// 	_, err = f.WriteString(res)
// 	assert.NoError(t, err)

// 	// 2nd STEP: (initiated by relayer)
// 	// VECTOR_MULTISIG -> VECTOR_SENDER
// 	type RequestBody2 struct {
// 		ChainID  string `json:"chainId"`
// 		RecvAddr string `json:"recv_addr"`
// 		Amount   int    `json:"amount"`
// 	}

// 	// Create the request body
// 	requestBody2 := RequestBody2{
// 		ChainID:  "vector",
// 		RecvAddr: os.Getenv("SENDER_ADDRESS_VECTOR"),
// 		Amount:   1000000,
// 	}

// 	// Marshal the request body to JSON
// 	requestBodyBytes, err = json.Marshal(requestBody2)
// 	assert.NoError(t, err)

// 	// Send HTTP POST request
// 	resp, err = http.Post("http://localhost:8000/api/createAndSignBatchingTx", "application/json", bytes.NewBuffer(requestBodyBytes))
// 	assert.NoError(t, err)
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err = io.ReadAll(resp.Body)
// 	assert.NoError(t, err)
// 	res = fmt.Sprintf("Succesfully submited batching tx to VECTOR %s\n", string(body))
// 	_, err = f.WriteString(res)
// 	assert.NoError(t, err)

// 	time.Sleep(time.Second * 2)

// 	// VECTOR -> PRIME
// 	// 1st STEP: (initiated by user)
// 	// VECTOR_SENDER -> VECTOR_MULTISIG
// 	// Define PRIME_SENDER as a receiver of bridged funds

// 	// Create the request body
// 	requestBody1 = RequestBody1{
// 		PrivKey:       os.Getenv("SENDER_KEY_VECTOR"),
// 		SenderAddress: os.Getenv("SENDER_ADDRESS_VECTOR"),
// 		RecvAddress:   os.Getenv("SENDER_ADDRESS_PRIME"),
// 		Amount:        1000000,
// 		ChainID:       "vector",
// 	}

// 	// Marshal the request body to JSON
// 	requestBodyBytes, err = json.Marshal(requestBody1)
// 	assert.NoError(t, err)

// 	// Send HTTP POST request
// 	resp, err = http.Post("http://localhost:8000/api/createAndSignBridgingTx", "application/json", bytes.NewBuffer(requestBodyBytes))
// 	assert.NoError(t, err)
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err = io.ReadAll(resp.Body)
// 	assert.NoError(t, err)
// 	res = fmt.Sprintf("Succesfully submited user tx to VECTOR %s\n", string(body))
// 	_, err = f.WriteString(res)
// 	assert.NoError(t, err)

// 	// 2nd STEP: (initiated by relayer)
// 	// PRIME_MULTISIG -> PRIME_SENDER
// 	// Create the request body
// 	requestBody2 = RequestBody2{
// 		ChainID:  "prime",
// 		RecvAddr: os.Getenv("SENDER_ADDRESS_PRIME"),
// 		Amount:   1000000,
// 	}

// 	// Marshal the request body to JSON
// 	requestBodyBytes, err = json.Marshal(requestBody2)
// 	assert.NoError(t, err)

// 	// Send HTTP POST request
// 	resp, err = http.Post("http://localhost:8000/api/createAndSignBatchingTx", "application/json", bytes.NewBuffer(requestBodyBytes))
// 	assert.NoError(t, err)
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err = io.ReadAll(resp.Body)
// 	assert.NoError(t, err)

// 	res = fmt.Sprintf("Succesfully submited bridging tx to PRIME %s\n", string(body))
// 	_, err = f.WriteString(res)
// 	assert.NoError(t, err)

// 	f.Close()
// }
