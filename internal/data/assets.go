package data

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// RobloxAssetsUpdatedChannel is published to by the workers whenever the hosted Roblox assets change.
const RobloxAssetsUpdatedChannel = "roblox:asset:updated"

const (
	RobloxAssetStatePending = "pending"
	RobloxAssetStateHosted  = "hosted"
	RobloxAssetStateBlocked = "blocked"
	RobloxAssetStateFailed  = "failed"
)

const RobloxAssetSourceThumbnail = "thumbnail"

const RobloxAssetKindImage = "image"

type RobloxAsset struct {
	AssetID string
	// The kind of file that is hosted. Only images are hosted.
	Kind  string
	State string
	// Where the hosted file was fetched from.
	Source string
	// The key of the hosted file in the bucket.
	Key         string
	ContentType string
	// The SHA-256 of the hosted file.
	Hash string
	Size int64

	// The failed attempts since the asset was last hosted.
	Attempts      int
	LastError     string
	NextAttemptAt time.Time
	FetchedAt     time.Time
	// The last time the asset was referenced by synced data.
	LastReferencedAt time.Time
}

type AssetStore struct {
	// Roblox assets, keyed by their asset ID.
	Roblox *RedisHash[RobloxAsset]
}

func NewAssetStore(c redis.UniversalClient) *AssetStore {
	return &AssetStore{
		Roblox: NewRedisHash[RobloxAsset](c, "roblox:asset"),
	}
}
