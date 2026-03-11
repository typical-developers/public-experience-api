package httpx

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// ETagJSON will generate a strong ETag from JSON data.
func ETagJSON(data any) (*string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	hash := md5.Sum(buf.Bytes())
	str := hex.EncodeToString(hash[:])[:32]
	etag := fmt.Sprintf(`"%s"`, str)

	return &etag, nil
}
