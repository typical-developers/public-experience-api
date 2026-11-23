package assets

import (
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/config"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/internal/storage"
)

func ServeAssets(client *s3.Client, store *data.OaklandsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		filename := chi.URLParam(r, "filename")

		if id == "" || filename == "" ||
			id == "." || id == ".." ||
			strings.ContainsAny(id+filename, "/\\") {
			render.Status(r, rest.HttpErrorNotFound.Status)
			render.JSON(w, r, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			})
			return
		}

		extension := path.Ext(filename)
		form := strings.TrimSuffix(filename, extension)
		if form == "" {
			render.Status(r, rest.HttpErrorNotFound.Status)
			render.JSON(w, r, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			})
			return
		}

		switch strings.ToLower(extension) {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".avif", ".svg":
		default:
			render.Status(r, rest.HttpErrorNotFound.Status)
			render.JSON(w, r, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			})
			return
		}

		unlockTime, err := store.ItemUnlockTimes.Get(r.Context(), id)
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("get unlock time for %q: %v", id, err)
			http.Error(w, "Unable to check asset visibility", http.StatusInternalServerError)
			return
		}
		if err == nil && time.Now().Before(unlockTime) {
			render.Status(r, rest.HttpErrorNotFound.Status)
			render.JSON(w, r, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			})
			return
		}

		item, err := store.Items.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				render.Status(r, rest.HttpErrorNotFound.Status)
				render.JSON(w, r, rest.ErrorResponse{
					Type:    rest.HttpErrorNotFound.Type,
					Message: rest.HttpErrorNotFound.Message,
				})
				return
			}

			log.Printf("get item %q: %v", id, err)
			http.Error(w, "Unable to check asset visibility", http.StatusInternalServerError)
			return
		}

		// An unopened gift can expose its wrapper forms,
		// but cannot expose the item or other unboxed forms.
		switch strings.ToLower(form) {
		case "store", "box", "gift":
		default:
			for _, itemForm := range item.Forms {
				if gift, ok := itemForm.Data.(data.OaklandsItemFormGift); ok && !gift.IsOpened() {
					render.Status(r, rest.HttpErrorNotFound.Status)
					render.JSON(w, r, rest.ErrorResponse{
						Type:    rest.HttpErrorNotFound.Type,
						Message: rest.HttpErrorNotFound.Message,
					})
					return
				}
			}
		}

		key := "items/" + id + "/" + filename

		object, err := client.GetObject(r.Context(), &s3.GetObjectInput{
			Bucket: aws.String(config.C.S3.Bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			var noSuchKey *types.NoSuchKey
			var notFound *types.NotFound
			if errors.As(err, &noSuchKey) || errors.As(err, &notFound) {
				render.Status(r, rest.HttpErrorNotFound.Status)
				render.JSON(w, r, rest.ErrorResponse{
					Type:    rest.HttpErrorNotFound.Type,
					Message: rest.HttpErrorNotFound.Message,
				})
				return
			}

			log.Printf("fetch asset %q: %v", key, err)
			http.Error(w, "Unable to fetch asset", http.StatusBadGateway)
			return
		}
		defer object.Body.Close()

		contentType := aws.ToString(object.ContentType)
		if contentType == "" {
			contentType = mime.TypeByExtension(strings.ToLower(extension))
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "public, max-age=60")

		if _, err := io.Copy(w, object.Body); err != nil {
			log.Printf("stream asset %q: %v", key, err)
		}
	}
}

// ServeRobloxAssets will serve the hosted file of a Roblox asset.
func ServeRobloxAssets(bucket *storage.Bucket, index *RobloxIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		filename := chi.URLParam(r, "filename")

		// Only the current file of a hosted asset is served, which also rejects any other path.
		asset, ok := index.Get(id)
		if !ok || asset.Key != "roblox/"+id+"/"+filename {
			render.Status(r, rest.HttpErrorNotFound.Status)
			render.JSON(w, r, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			})
			return
		}

		body, info, err := bucket.Get(r.Context(), asset.Key)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				render.Status(r, rest.HttpErrorNotFound.Status)
				render.JSON(w, r, rest.ErrorResponse{
					Type:    rest.HttpErrorNotFound.Type,
					Message: rest.HttpErrorNotFound.Message,
				})
				return
			}

			log.Printf("fetch asset %q: %v", asset.Key, err)
			http.Error(w, "Unable to fetch asset", http.StatusBadGateway)
			return
		}
		defer body.Close()

		contentType := info.ContentType
		if contentType == "" {
			contentType = asset.ContentType
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// The key contains the file's hash, so its content never changes.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

		if _, err := io.Copy(w, body); err != nil {
			log.Printf("stream asset %q: %v", asset.Key, err)
		}
	}
}
