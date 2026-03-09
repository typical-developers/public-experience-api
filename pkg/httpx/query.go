package httpx

import "net/http"

// QueryGet will get the value from a query key.
func QueryGet(r *http.Request, key string, fallback ...string) string {
	if v := r.URL.Query().Get(key); v != "" {
		return v
	}

	if len(fallback) > 0 && fallback[0] != "" {
		return fallback[0]
	}

	return ""
}
