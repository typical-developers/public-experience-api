package httpx

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// ETagJSON will generate a strong ETag from JSON data.
func ETagJSON(data any) (*string, error) {
	jsonb, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(jsonb)
	str := hex.EncodeToString(hash[:])[:32]
	etag := fmt.Sprintf(`"%s"`, str)

	return &etag, nil
}
