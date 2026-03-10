package httpx

import (
	"encoding/json"
	"net/http"
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

	match := r.Header.Get("If-None-Match")
	if *etag != match {
		w.Header().Set("ETag", *etag)
		return WriteJSON(w, data, statusCode)
	}

	w.WriteHeader(http.StatusNotModified)
	return nil
}
