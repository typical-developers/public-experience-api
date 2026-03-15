package redisx

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// JSONCmdUnwrap will unwrap the value from a redis JSONCmd.
// This will only work with "$" selectors.
func JSONCmdUnwrap[T any](cmd *redis.JSONCmd, data *T) error {
	var wrapped []T

	result, err := cmd.Result()
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(result), &wrapped); err != nil {
		return err
	}

	if len(wrapped) == 0 {
		return redis.Nil
	}

	*data = wrapped[0]
	return nil
}

// JSONUnwrap will get a key-value from Redis and decode the JSON into a parameter value.
//
// Deprecated: JSONUnwrap is deprecated. Use JSONCmdUnwrap instead. Calling this method will use JSONCmdUnwrap under the hood.
func JSONUnwrap[T any](ctx context.Context, r *redis.Client, key string, selector string, data *T) error {
	cmd := r.JSONGet(ctx, key, selector)
	return JSONCmdUnwrap(cmd, data)
}
