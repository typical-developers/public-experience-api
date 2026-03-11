package redisx

import (
	"fmt"
	"time"
)

func ParseTime(v any) (*time.Time, error) {
	if v == nil {
		return nil, nil
	}

	var timestamp string
	switch value := v.(type) {
	case string:
		timestamp = value
	case []byte:
		timestamp = string(value)
	}

	if timestamp == "" {
		return nil, nil
	}

	if t, err := time.Parse(time.RFC3339, timestamp); err != nil {
		return nil, fmt.Errorf("invalid timestamp: %s", timestamp)
	} else {
		return &t, nil
	}
}
