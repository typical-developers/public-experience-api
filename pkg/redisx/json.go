package redisx

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

func JSONUnwrap[T any](ctx context.Context, r *redis.Client, key string, selector string, data *T) error {
	cmd := r.JSONGet(ctx, key, selector)

	err := cmd.Err()
	if err != nil {
		return err
	}

	if cmd.Val() == "" {
		return nil
	}

	var raw any
	if err := json.Unmarshal([]byte(cmd.Val()), &raw); err != nil {
		return err
	}

	// Redis JSON.GET with a selector like "$" returns an array of matches.
	if entries, ok := raw.([]any); ok {
		if len(entries) == 0 || entries[0] == nil {
			return nil
		}

		jsonB, err := json.Marshal(entries[0])
		if err != nil {
			return err
		}

		return json.Unmarshal(jsonB, data)
	}

	jsonB, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonB, data)
}
