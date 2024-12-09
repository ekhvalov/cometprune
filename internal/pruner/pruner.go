package pruner

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"

	db "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func Prune(path string, keepBlocks int64) error {
	if err := validate(path, keepBlocks); err != nil {
		return err
	}

	options := &opt.Options{
		DisableSeeksCompaction: true,
	}

	blockStoreDB, err := db.NewGoLevelDBWithOpts("blockstore", path, options)
	if err != nil {
		return fmt.Errorf("blockstore open error: %w", err)
	}
	defer blockStoreDB.Close()
	blockStore := store.NewBlockStore(blockStoreDB)
	size := blockStore.Size()
	if size < keepBlocks {
		fmt.Printf("Block store size is %d, so nothing to do.\n", size)
		return nil
	}
	pruneHeight := blockStore.Height() - keepBlocks

	stateDB, err := db.NewGoLevelDBWithOpts("state", path, options)
	if err != nil {
		return fmt.Errorf("state open error: %w", err)
	}
	defer stateDB.Close()

	minHeight := blockStore.Base()

	errGroup := new(errgroup.Group)
	errGroup.Go(func() error {
		return pruneBlockStore(blockStoreDB, blockStore, pruneHeight)
	})
	errGroup.Go(func() error {
		return pruneStateStore(stateDB, minHeight, pruneHeight)
	})

	return errGroup.Wait()
}

func validate(path string, keepBlocks int64) error {
	if path == "" {
		return fmt.Errorf("validation error: %w", errors.New("path required"))
	}

	if keepBlocks <= 0 {
		return fmt.Errorf("validation error: %w", errors.New("keep-blocks must be greater than zero"))
	}

	return nil
}

func pruneBlockStore(blockStoreDB *db.GoLevelDB, blockStore *store.BlockStore, pruneHeight int64) error {
	fmt.Printf("Block store: pruning started\n")
	prunedBlocks, err := blockStore.PruneBlocks(pruneHeight)
	if err != nil {
		return fmt.Errorf("could not prune blocks: %w", err)
	}
	fmt.Printf("Block store: pruning finished; %d block(s) pruned\n", prunedBlocks)

	fmt.Println("Block store: compacting started")
	if err := blockStoreDB.Compact(nil, nil); err != nil {
		return fmt.Errorf("could not compact block store: %w", err)
	}
	fmt.Println("Block store: compacting finished")

	return nil
}

func pruneStateStore(stateDB db.DB, fromHeight, toHeight int64) error {
	stateStore := state.NewStore(stateDB, state.StoreOptions{DiscardABCIResponses: false})

	fmt.Printf("State store: pruning started\n")
	if err := stateStore.PruneStates(fromHeight, toHeight); err != nil {
		return fmt.Errorf("could not prune state store: %w", err)
	}
	fmt.Printf("State store: pruning finished; %d state(s) pruned\n", toHeight-fromHeight)

	fmt.Println("State store: compacting started")
	if err := stateDB.Compact(nil, nil); err != nil {
		return fmt.Errorf("could not compact state store: %w", err)
	}
	fmt.Println("State store: compacting finished")

	return nil
}
