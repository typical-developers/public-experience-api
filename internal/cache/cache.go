package cache

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
)

func init() {
	var opts redis.Options

	if host := os.Getenv("REDIS_HOST"); host != "" {
		opts.Addr = host
	}

	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		opts.Password = password
	}

	if db := os.Getenv("REDIS_DB"); db != "" {
		dbNum, err := strconv.Atoi(db)
		if err != nil {
		} else {
			opts.DB = dbNum
		}
	}

	Client = redis.NewClient(&opts)
}

func GetCached[T any](ctx context.Context, key string, path string) *T {
	v := Client.JSONGet(ctx, key, path)

	if v.Val() == "" {
		return nil
	}

	expanded, err := v.Expanded()
	if err != nil {
		return nil
	}

	entries := expanded.([]any)
	entry := entries[0]

	if entry == nil {
		return nil
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		return nil
	}

	var t T
	err = json.Unmarshal(jsonBytes, &t)
	if err != nil {
		return nil
	}

	return &t
}

type CacheOpts struct {
	Expiry *time.Duration
}

func SetCached[T any](ctx context.Context, key string, path string, value T, opts *CacheOpts) {
	Client.JSONSet(ctx, key, path, value)

	if opts != nil {
		if opts.Expiry != nil {
			_ = Client.Expire(ctx, key, *opts.Expiry)
		}
	}
}
