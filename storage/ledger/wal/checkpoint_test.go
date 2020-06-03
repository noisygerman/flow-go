package wal_test

import (
	"fmt"
	"io"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/module/metrics"
	"github.com/dapperlabs/flow-go/storage/ledger"
	"github.com/dapperlabs/flow-go/storage/ledger/mtrie"
	"github.com/dapperlabs/flow-go/storage/ledger/utils"
	realWAL "github.com/dapperlabs/flow-go/storage/ledger/wal"
	"github.com/dapperlabs/flow-go/storage/util"
	"github.com/dapperlabs/flow-go/utils/unittest"
)

var dir = "./wal"

func RunWithWALCheckpointerWithFiles(t *testing.T, names ...interface{}) {
	f := names[len(names)-1].(func(*testing.T, *realWAL.LedgerWAL, *realWAL.Checkpointer))

	fileNames := make([]string, len(names)-1)

	for i := 0; i <= len(names)-2; i++ {
		fileNames[i] = names[i].(string)
	}

	unittest.RunWithTempDir(t, func(dir string) {
		util.CreateFiles(t, dir, fileNames...)

		wal, err := realWAL.NewWAL(nil, nil, dir, 10, 9)
		require.NoError(t, err)

		checkpointer, err := wal.Checkpointer()
		require.NoError(t, err)

		f(t, wal, checkpointer)
	})
}

func Test_WAL(t *testing.T) {

	numInsPerStep := 2
	keyByteSize := 32
	valueMaxByteSize := 2 << 16 //16kB
	size := 10
	metricsCollector := &metrics.NoopCollector{}

	unittest.RunWithTempDir(t, func(dir string) {

		f, err := ledger.NewMTrieStorage(dir, size*10, metricsCollector, nil)
		require.NoError(t, err)

		var stateCommitment = f.EmptyStateCommitment()

		//saved data after updates
		savedData := make(map[string]map[string][]byte)

		// WAL segments are 32kB, so here we generate 2 keys 16kB each, times `size`
		// so we should get at least `size` segments
		for i := 0; i < size; i++ {

			keys := utils.GetRandomKeysFixedN(numInsPerStep, keyByteSize)
			values := utils.GetRandomValues(len(keys), 10, valueMaxByteSize)

			stateCommitment, err = f.UpdateRegisters(keys, values, stateCommitment)
			require.NoError(t, err)

			fmt.Printf("Updated with %x\n", stateCommitment)

			data := make(map[string][]byte, len(keys))
			for j, key := range keys {
				data[string(key)] = values[j]
			}

			savedData[string(stateCommitment)] = data
		}

		<-f.Done()

		f2, err := ledger.NewMTrieStorage(dir, (size*10)+10, metricsCollector, nil)
		require.NoError(t, err)

		// random map iteration order is a benefit here
		for stateCommitment, data := range savedData {

			keys := make([][]byte, 0, len(data))
			for keyString := range data {
				key := []byte(keyString)
				keys = append(keys, key)
			}

			fmt.Printf("Querying with %x\n", stateCommitment)

			registerValues, err := f2.GetRegisters(keys, []byte(stateCommitment))
			require.NoError(t, err)

			for i, key := range keys {
				registerValue := registerValues[i]

				assert.Equal(t, data[string(key)], registerValue)
			}
		}

		<-f2.Done()

	})
}

func Test_Checkpointing(t *testing.T) {

	numInsPerStep := 2
	keyByteSize := 4
	valueMaxByteSize := 2 << 16 //64kB
	size := 10
	metricsCollector := &metrics.NoopCollector{}

	unittest.RunWithTempDir(t, func(dir string) {

		f, err := mtrie.NewMForest(33, dir, size*10, metricsCollector, func(tree *mtrie.MTrie) error { return nil })
		require.NoError(t, err)

		var stateCommitment = f.GetEmptyRootHash()

		//saved data after updates
		savedData := make(map[string]map[string][]byte)

		t.Run("create WAL and inital trie", func(t *testing.T) {

			wal, err := realWAL.NewWAL(nil, nil, dir, size*10, 33)
			require.NoError(t, err)

			// WAL segments are 32kB, so here we generate 2 keys 64kB each, times `size`
			// so we should get at least `size` segments

			// Generate the tree and create WAL
			for i := 0; i < size; i++ {

				keys := utils.GetRandomKeysFixedN(numInsPerStep, keyByteSize)
				values := utils.GetRandomValues(len(keys), valueMaxByteSize, valueMaxByteSize)

				err = wal.RecordUpdate(stateCommitment, keys, values)
				require.NoError(t, err)

				stateCommitment, err = f.Update(keys, values, stateCommitment)
				require.NoError(t, err)

				fmt.Printf("Updated with %x\n", stateCommitment)

				data := make(map[string][]byte, len(keys))
				for j, key := range keys {
					data[string(key)] = values[j]
				}

				savedData[string(stateCommitment)] = data
			}
			err = wal.Close()
			require.NoError(t, err)

			require.FileExists(t, path.Join(dir, "00000010")) //make sure we have enough segments saved
		})

		// create a new forest and reply WAL
		f2, err := mtrie.NewMForest(33, dir, size*10, metricsCollector, func(tree *mtrie.MTrie) error { return nil })
		require.NoError(t, err)

		t.Run("replay WAL and create checkpoint", func(t *testing.T) {

			require.NoFileExists(t, path.Join(dir, "checkpoint.00000010"))

			wal2, err := realWAL.NewWAL(nil, nil, dir, size*10, 33)
			require.NoError(t, err)

			err = wal2.Replay(
				func(nodes []*mtrie.StorableNode, tries []*mtrie.StorableTrie) error {
					return fmt.Errorf("I should fail as there should be no checkpoints")
				},
				func(commitment flow.StateCommitment, keys [][]byte, values [][]byte) error {
					_, err := f2.Update(keys, values, commitment)
					return err
				},
				func(commitment flow.StateCommitment) error {
					return fmt.Errorf("I should fail as there should be no deletions")
				},
			)
			require.NoError(t, err)

			checkpointer, err := wal2.Checkpointer()
			require.NoError(t, err)

			err = checkpointer.Checkpoint(10, func() (io.WriteCloser, error) {
				return checkpointer.CheckpointWriter(10)
			})
			require.NoError(t, err)

			require.FileExists(t, path.Join(dir, "checkpoint.00000010")) //make sure we have checkpoint file

			err = wal2.Close()
			require.NoError(t, err)
		})

		f3, err := mtrie.NewMForest(33, dir, size*10, metricsCollector, func(tree *mtrie.MTrie) error { return nil })
		require.NoError(t, err)

		t.Run("read checkpoint", func(t *testing.T) {
			wal3, err := realWAL.NewWAL(nil, nil, dir, size*10, 33)
			require.NoError(t, err)

			err = wal3.Replay(
				func(nodes []*mtrie.StorableNode, tries []*mtrie.StorableTrie) error {
					return f3.LoadStorables(nodes, tries)
				},
				func(commitment flow.StateCommitment, keys [][]byte, values [][]byte) error {
					return fmt.Errorf("I should fail as there should be no updates")
				},
				func(commitment flow.StateCommitment) error {
					return fmt.Errorf("I should fail as there should be no deletions")
				},
			)
			require.NoError(t, err)

			err = wal3.Close()
			require.NoError(t, err)
		})

		t.Run("all forests contain the same data", func(t *testing.T) {
			// random map iteration order is a benefit here
			// make sure the tries has been rebuilt from WAL and another from from Checkpoint
			// f1, f2 and f3 should be identical
			for stateCommitment, data := range savedData {

				keys := make([][]byte, 0, len(data))
				for keyString := range data {
					key := []byte(keyString)
					keys = append(keys, key)
				}

				fmt.Printf("Querying with %x\n", stateCommitment)

				registerValues, err := f.Read(keys, []byte(stateCommitment))
				require.NoError(t, err)

				registerValues2, err := f2.Read(keys, []byte(stateCommitment))
				require.NoError(t, err)

				registerValues3, err := f3.Read(keys, []byte(stateCommitment))
				require.NoError(t, err)

				for i, key := range keys {
					require.Equal(t, data[string(key)], registerValues[i])
					require.Equal(t, data[string(key)], registerValues2[i])
					require.Equal(t, data[string(key)], registerValues3[i])
				}
			}
		})

		keys2 := utils.GetRandomKeysFixedN(numInsPerStep, keyByteSize)
		values2 := utils.GetRandomValues(len(keys2), valueMaxByteSize, valueMaxByteSize)

		t.Run("create segment after checkpoint", func(t *testing.T) {

			require.NoFileExists(t, path.Join(dir, "00000011"))

			//generate one more segment
			wal4, err := realWAL.NewWAL(nil, nil, dir, size*10, 33)
			require.NoError(t, err)

			err = wal4.RecordUpdate(stateCommitment, keys2, values2)
			require.NoError(t, err)

			stateCommitment, err = f.Update(keys2, values2, stateCommitment)
			require.NoError(t, err)

			err = wal4.Close()
			require.NoError(t, err)

			require.FileExists(t, path.Join(dir, "00000011")) //make sure we have extra segment
		})

		f5, err := mtrie.NewMForest(33, dir, size*10, metricsCollector, func(tree *mtrie.MTrie) error { return nil })
		require.NoError(t, err)

		t.Run("replay both checkpoint and updates after checkpoint", func(t *testing.T) {
			wal5, err := realWAL.NewWAL(nil, nil, dir, size*10, 33)
			require.NoError(t, err)

			updatesLeft := 1 // there should be only one update

			err = wal5.Replay(
				func(nodes []*mtrie.StorableNode, tries []*mtrie.StorableTrie) error {
					return f5.LoadStorables(nodes, tries)
				},
				func(commitment flow.StateCommitment, keys [][]byte, values [][]byte) error {
					if updatesLeft == 0 {
						return fmt.Errorf("more updates called then expected")
					}
					_, err := f5.Update(keys, values, commitment)
					updatesLeft--
					return err
				},
				func(commitment flow.StateCommitment) error {
					return fmt.Errorf("I should fail as there should be no deletions")
				},
			)
			require.NoError(t, err)

			err = wal5.Close()
			require.NoError(t, err)
		})

		t.Run("extra updates were applied correctly", func(t *testing.T) {
			registerValues, err := f.Read(keys2, []byte(stateCommitment))
			require.NoError(t, err)

			registerValues5, err := f5.Read(keys2, []byte(stateCommitment))
			require.NoError(t, err)

			for i, _ := range keys2 {
				require.Equal(t, values2[i], registerValues[i])
				require.Equal(t, values2[i], registerValues5[i])
			}
		})

	})
}