package rest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

type etagRecorder struct {
	http.ResponseWriter
	buf    bytes.Buffer
	status int
}

func (r *etagRecorder) WriteHeader(code int) {
	r.status = code
}

func (r *etagRecorder) Write(b []byte) (int, error) {
	return r.buf.Write(b)
}

func etagMatch(header, etag string) bool {
	want := strings.TrimPrefix(etag, "W/")
	for _, c := range strings.Split(header, ",") {
		c = strings.TrimSpace(c)
		if c == "*" || strings.TrimPrefix(c, "W/") == want {
			return true
		}
	}
	return false
}

func ETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		writer := &etagRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(writer, r)

		if writer.status == http.StatusOK {
			sum := sha256.Sum256(writer.buf.Bytes())
			etag := `W/"` + hex.EncodeToString(sum[:16]) + `"`
			w.Header().Set("ETag", etag)

			if etagMatch(r.Header.Get("If-None-Match"), etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.WriteHeader(writer.status)
		_, _ = w.Write(writer.buf.Bytes())
	})
}
