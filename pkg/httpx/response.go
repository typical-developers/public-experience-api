package httpx

import (
	"encoding/json"
	"net/http"
	"strings"
)

func WriteJSON(w http.ResponseWriter, data any, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(data)
}

func WriteJSONWithETag(w http.ResponseWriter, r *http.Request, data any, statusCode int) error {
	etag, err := ETagJSON(data)
	if err != nil {
		return err
	}

	w.Header().Set("ETag", *etag)

	match := r.Header.Get("If-None-Match")
	if match == "*" {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	for _, part := range strings.Split(match, ",") {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}

		if strings.HasPrefix(token, "W/") || strings.HasPrefix(token, "w/") {
			token = strings.TrimSpace(token[2:])
		}

		if token == *etag {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
	}

	return WriteJSON(w, data, statusCode)
}
