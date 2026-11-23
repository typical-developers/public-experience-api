package assets

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/internal/data"
)

// RobloxIndex keeps the hosted Roblox assets in memory, keyed by their asset ID,
// so handlers don't call Redis for every asset in a response.
type RobloxIndex struct {
	redis *redis.Client
	store *data.AssetStore

	snap atomic.Pointer[map[string]data.RobloxAsset]
}

func NewRobloxIndex(redis *redis.Client, store *data.AssetStore) *RobloxIndex {
	ix := &RobloxIndex{
		redis: redis,
		store: store,
	}

	ix.snap.Store(&map[string]data.RobloxAsset{})

	return ix
}

// Start does one synchronous refresh (so the index is populated before you serve traffic),
// then refreshes whenever the workers publish an update to the hosted assets.
// It also refreshes every interval, in case an update was missed.
func (ix *RobloxIndex) Start(ctx context.Context, interval time.Duration) error {
	if err := ix.Refresh(ctx); err != nil {
		return fmt.Errorf("initial roblox asset index refresh: %w", err)
	}

	updates := ix.redis.Subscribe(ctx, data.RobloxAssetsUpdatedChannel)
	messages := updates.Channel()

	go func() {
		defer updates.Close()

		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-messages:
			case <-t.C:
			}

			if err := ix.Refresh(ctx); err != nil {
				log.Printf("roblox asset index refresh: %v", err)
			}
		}
	}()

	return nil
}

// Refresh loads the hosted assets and atomically swaps in a new snapshot.
func (ix *RobloxIndex) Refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ids, err := ix.store.Roblox.Keys(ctx)
	if err != nil {
		return err
	}

	assets, err := ix.store.Roblox.GetMany(ctx, ids...)
	if err != nil {
		return err
	}

	hosted := make(map[string]data.RobloxAsset, len(assets))
	for id, asset := range assets {
		if asset.State == data.RobloxAssetStateHosted {
			hosted[id] = asset
		}
	}

	ix.snap.Store(&hosted)
	return nil
}

// Get returns the hosted asset for a Roblox asset ID, if it is hosted.
func (ix *RobloxIndex) Get(id string) (data.RobloxAsset, bool) {
	asset, ok := (*ix.snap.Load())[id]
	return asset, ok
}
