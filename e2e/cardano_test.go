package e2e

import (
	"fmt"
	"os"
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
				cardanofw.WithPort(3210+id*100),
				cardanofw.WithLogsDir(logsDir),
				cardanofw.WithNetworkMagic(42+id))
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
			if errors[id] = cluster.StartOgmiosOnNode(uint((id+1)*1000 + 300)); errors[id] != nil {
				return
			}

			errors[id] = cluster.WaitForBlockWithState(300, time.Hour*1)
		}()
	}

	wg.Wait()

	fmt.Print("sender address on prime = ")
	fmt.Println(os.Getenv("SENDER_ADDRESS_PRIME"))
	fmt.Print("sender key on prime = ")
	fmt.Println(os.Getenv("SENDER_KEY_PRIME"))
	fmt.Print("multisig address on prime = ")
	fmt.Println(os.Getenv("MULTISIG_ADDRESS_PRIME"))
	fmt.Println(os.Getenv("MULTISIG_FEE_ADDRESS_PRIME"))
	fmt.Println(os.Getenv("BRIDGE_ADDRESS_PRIME"))
	fmt.Print("multisig key on prime = ")
	fmt.Println(os.Getenv("BRIDGE_ADDRESS_KEY_PRIME"))

	fmt.Println()
	fmt.Print("sender address on vector = ")
	fmt.Println(os.Getenv("SENDER_ADDRESS_VECTOR"))
	fmt.Print("sender key on vector = ")
	fmt.Println(os.Getenv("SENDER_KEY_VECTOR"))
	fmt.Print("multisig address on vector = ")
	fmt.Println(os.Getenv("MULTISIG_ADDRESS_VECTOR"))
	fmt.Println(os.Getenv("MULTISIG_FEE_ADDRESS_VECTOR"))
	fmt.Println(os.Getenv("BRIDGE_ADDRESS_VECTOR"))
	fmt.Print("multisig key on vector = ")
	fmt.Println(os.Getenv("BRIDGE_ADDRESS_KEY_VECTOR"))

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
// 	unsgignedTx, err := user.CreateBridgingTransaction(os.Getenv("SENDER_ADDRESS_PRIME"), "prime", map[string]uint{os.Getenv("SENDER_ADDRESS_VECTOR"): uint(1000000)})
// 	if err != nil {
// 		res := fmt.Sprintf("ERROR %s\n", err.Error())
// 		_, err = f.WriteString(res)
// 		assert.NoError(t, err)
// 	}

// 	txHash, err := user.SignAndSubmitTransaction(unsgignedTx, os.Getenv("SENDER_KEY_PRIME"), "prime")
// 	if err != nil {
// 		res := fmt.Sprintf("ERROR %s\n", err.Error())
// 		_, err = f.WriteString(res)
// 		assert.NoError(t, err)
// 	}
// 	res := fmt.Sprintf("Succesfully submited user tx to PRIME %s\n", txHash)
// 	_, err = f.WriteString(res)
// 	assert.NoError(t, err)

// 	// 2nd STEP: (initiated by relayer)
// 	// VECTOR_MULTISIG -> VECTOR_SENDER
// 	txHash, err = batcher.BuildAndSubmitBatchingTx("vector", map[string]uint{os.Getenv("SENDER_ADDRESS_VECTOR"): uint(1000000)})
// 	if err != nil {
// 		res := fmt.Sprintf("ERROR %s\n", err.Error())
// 		_, err = f.WriteString(res)
// 		assert.NoError(t, err)
// 	}
// 	res = fmt.Sprintf("Succesfully submited batching tx to VECTOR %s\n", txHash)
// 	_, err = f.WriteString(res)
// 	assert.NoError(t, err)

// 	time.Sleep(time.Second * 2)

// 	// VECTOR -> PRIME
// 	// 1st STEP: (initiated by user)
// 	// VECTOR_SENDER -> VECTOR_MULTISIG
// 	// Define PRIME_SENDER as a receiver of bridged funds
// 	unsgignedTx, err = user.CreateBridgingTransaction(os.Getenv("SENDER_ADDRESS_VECTOR"), "vector", map[string]uint{os.Getenv("SENDER_ADDRESS_PRIME"): uint(1000000)})
// 	if err != nil {
// 		res := fmt.Sprintf("ERROR %s\n", err.Error())
// 		_, err = f.WriteString(res)
// 		assert.NoError(t, err)
// 	}

// 	txHash, err = user.SignAndSubmitTransaction(unsgignedTx, os.Getenv("SENDER_KEY_VECTOR"), "vector")
// 	if err != nil {
// 		res := fmt.Sprintf("ERROR %s\n", err.Error())
// 		_, err = f.WriteString(res)
// 		assert.NoError(t, err)
// 	}
// 	res = fmt.Sprintf("Succesfully submited user tx to VECTOR %s\n", txHash)
// 	_, err = f.WriteString(res)
// 	assert.NoError(t, err)

// 	// 2nd STEP: (initiated by relayer)
// 	// PRIME_MULTISIG -> PRIME_SENDER
// 	txHash, err = batcher.BuildAndSubmitBatchingTx("prime", map[string]uint{os.Getenv("SENDER_ADDRESS_PRIME"): uint(1000000)})
// 	if err != nil {
// 		res := fmt.Sprintf("ERROR %s\n", err.Error())
// 		_, err = f.WriteString(res)
// 		assert.NoError(t, err)
// 	}
// 	res = fmt.Sprintf("Succesfully submited bridging tx to PRIME %s\n", txHash)
// 	_, err = f.WriteString(res)
// 	assert.NoError(t, err)

// 	f.Close()
// }
