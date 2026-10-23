package static

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/apperror"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

var (
	//go:embed all:files
	files embed.FS

	root      fs.FS
	fsHandler http.Handler
)

func init() {
	fileRoot, err := fs.Sub(files, "files")
	if err != nil {
		panic(err)
	}

	root = fileRoot
	fsHandler = http.FileServer(http.FS(root))
}

// checkForFile will check if a file exists based on its uPath.
func checkForFile(uPath string) *apperror.AppError {
	path := filepath.Clean(uPath)
	path = filepath.ToSlash(path)

	if path == "." || path == "/" {
		return rest.HttpErrorNotFound
	}
	path = strings.TrimPrefix(path, "/")

	file, err := fs.Stat(root, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return rest.HttpErrorNotFound
		}

		return rest.HttpInternalServerError.WithError(err)
	}

	if file.IsDir() {
		return rest.HttpErrorNotFound
	}

	return nil
}

// ServeStatic will return an http handler for serving static files.
func ServeStatic(store *data.OaklandsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := checkForFile(r.URL.Path); err != nil {
			_ = httpx.WriteJSON(w, rest.ErrorResponse{
				Type:    err.Type,
				Message: err.Message,
			}, err.Status)

			return
		}

		parts := strings.Split(strings.Trim(path.Clean("/"+strings.ReplaceAll(r.URL.Path, "\\", "/")), "/"), "/")
		var id, form string
		switch {
		case len(parts) == 4 && parts[0] == "assets" && parts[1] == "items":
			id, form = parts[2], strings.TrimSuffix(parts[3], path.Ext(parts[3]))
		case len(parts) == 3 && parts[0] == "assets" && parts[1] != "items":
			id, form = strings.TrimSuffix(parts[2], path.Ext(parts[2])), parts[1]
		}
		var accessError *apperror.AppError
		if id != "" {
			unlockAt, err := store.ItemUnlockTimes.Get(r.Context(), id)
			switch {
			case err != nil && !errors.Is(err, redis.Nil):
				accessError = rest.HttpInternalServerError.WithError(err)
			case err == nil && time.Now().Before(unlockAt):
				accessError = rest.HttpErrorNotFound
			default:
				item, err := store.Items.Get(r.Context(), id)
				switch {
				case errors.Is(err, redis.Nil):
					accessError = rest.HttpErrorResourceNotCached
				case err != nil:
					accessError = rest.HttpInternalServerError.WithError(err)
				default:
					switch strings.ToLower(form) {
					case "store", "box", "gift":
					default:
						for _, form := range item.Forms {
							if gift, ok := form.Data.(data.OaklandsItemFormGift); ok && !gift.IsOpened() {
								accessError = rest.HttpErrorNotFound
								break
							}
						}
					}
				}
			}
		}
		if accessError != nil {
			w.Header().Set("Cache-Control", "no-store")
			_ = httpx.WriteJSON(w, rest.ErrorResponse{Type: accessError.Type, Message: accessError.Message}, accessError.Status)
			return
		}

		http.StripPrefix("/", fsHandler).ServeHTTP(w, r)
	}
}
