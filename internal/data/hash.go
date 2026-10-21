package data

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"slices"

	"github.com/redis/go-redis/v9"
)

type RedisHash[T any] struct {
	client redis.UniversalClient
	key    string
}

func NewRedisHash[T any](c redis.UniversalClient, key string) *RedisHash[T] {
	return &RedisHash[T]{
		client: c,
		key:    key,
	}
}

func (r *RedisHash[T]) Keys(ctx context.Context) ([]string, error) {
	return r.client.HKeys(ctx, r.key).Result()
}

func (r *RedisHash[T]) Get(ctx context.Context, field string) (T, error) {
	var value T

	raw, err := r.client.HGet(ctx, r.key, field).Result()
	if err != nil {
		return value, err
	}

	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		var zero T
		return zero, err
	}

	return value, nil
}

func (r *RedisHash[T]) GetMany(ctx context.Context, fields ...string) (map[string]T, error) {
	result := make(map[string]T, len(fields))
	if len(fields) == 0 {
		return result, nil
	}

	values, err := r.client.HMGet(ctx, r.key, fields...).Result()
	if err != nil {
		return nil, err
	}

	for i, raw := range values {
		if raw == nil {
			continue // Field doesn't exist.
		}

		text, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf(
				"hash field %q: unexpected type %T", fields[i], raw,
			)
		}

		var value T
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			return nil, fmt.Errorf(
				"decode hash field %q: %w", fields[i], err,
			)
		}

		result[fields[i]] = value
	}

	return result, nil
}

func (r *RedisHash[T]) Set(ctx context.Context, field string, value T) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.HSet(ctx, r.key, field, raw).Err()
}

func (r *RedisHash[T]) SetMany(ctx context.Context, fieldValues map[string]T) error {
	if len(fieldValues) == 0 {
		return nil
	}

	encoded := make(map[string]any, len(fieldValues))

	for field, value := range fieldValues {
		raw, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("encode hash field %q: %w", field, err)
		}

		encoded[field] = raw
	}

	return r.client.HSet(ctx, r.key, encoded).Err()
}

func (r *RedisHash[T]) Delete(ctx context.Context, field string) error {
	return r.client.HDel(ctx, r.key, field).Err()
}

func (r *RedisHash[T]) Len(ctx context.Context) int64 {
	len, err := r.client.HLen(ctx, r.key).Result()
	if err != nil {
		return 0
	}

	return len
}

func (r *RedisHash[T]) Values(ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		values, err := r.client.HVals(ctx, r.key).Result()

		if err != nil {
			return
		}

		for _, value := range values {
			var data T
			if err := json.Unmarshal([]byte(value), &data); err != nil {
				return
			}

			if !yield(data) {
				return
			}
		}
	}
}

func (r *RedisHash[T]) Entries(ctx context.Context) iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		values, err := r.client.HGetAll(ctx, r.key).Result()
		if err != nil {
			return
		}

		for field, raw := range values {
			var value T
			if err := json.Unmarshal([]byte(raw), &value); err != nil {
				return
			}

			if !yield(field, value) {
				return
			}
		}
	}
}

func (r *RedisHash[T]) SortedValues(ctx context.Context, compare func(a, b T) int) ([]T, error) {
	values, err := r.client.HGetAll(ctx, r.key).Result()
	if err != nil {
		return nil, err
	}

	items := make([]T, 0, len(values))
	for field, raw := range values {
		var item T
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			return nil, fmt.Errorf("decode hash field %q: %w", field, err)
		}

		items = append(items, item)
	}

	slices.SortFunc(items, compare)
	return items, nil
}
