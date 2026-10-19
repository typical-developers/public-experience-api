package data

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisValue[T any] struct {
	client redis.UniversalClient
	key    string
}

func NewRedisValue[T any](c redis.UniversalClient, key string) *RedisValue[T] {
	return &RedisValue[T]{
		client: c,
		key:    key,
	}
}

func (r *RedisValue[T]) Get(ctx context.Context) (T, error) {
	var value T

	raw, err := r.client.Get(ctx, r.key).Result()
	if err != nil {
		return value, err
	}

	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return *new(T), err
	}

	return value, nil
}

func (r *RedisValue[T]) Set(ctx context.Context, value T) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.key, raw, 0).Err()
}

func (r *RedisValue[T]) Delete(ctx context.Context) error {
	return r.client.Del(ctx, r.key).Err()
}
