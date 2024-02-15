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
	const (
		clusterCnt = 1
	)

	var (
		errors      [clusterCnt]error
		wg          sync.WaitGroup
		baseLogsDir string = path.Join("../..", fmt.Sprintf("e2e-logs-cardano-%d", time.Now().Unix()), t.Name())
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
				cardanofw.WithNodesCount(4),
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

			errors[id] = cluster.WaitForBlockWithState(10, time.Second*200)
		}()
	}

	wg.Wait()

	for i := 0; i < clusterCnt; i++ {
		assert.NoError(t, errors[i])
	}
}
