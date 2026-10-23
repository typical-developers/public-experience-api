package assets

import (
	"context"
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
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

func ServeAssets(client *s3.Client, store *data.OaklandsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		filename := chi.URLParam(r, "filename")

		if id == "" || filename == "" ||
			id == "." || id == ".." ||
			strings.ContainsAny(id+filename, "/\\") {
			_ = httpx.WriteJSON(w, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			}, rest.HttpErrorNotFound.Status)
			return
		}

		extension := path.Ext(filename)
		form := strings.TrimSuffix(filename, extension)
		if form == "" {
			_ = httpx.WriteJSON(w, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			}, rest.HttpErrorNotFound.Status)
			return
		}

		switch strings.ToLower(extension) {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".avif", ".svg":
		default:
			_ = httpx.WriteJSON(w, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			}, rest.HttpErrorNotFound.Status)
			return
		}

		unlockTime, err := store.ItemUnlockTimes.Get(r.Context(), id)
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("get unlock time for %q: %v", id, err)
			http.Error(w, "Unable to check asset visibility", http.StatusInternalServerError)
			return
		}
		if err == nil && time.Now().Before(unlockTime) {
			_ = httpx.WriteJSON(w, rest.ErrorResponse{
				Type:    rest.HttpErrorNotFound.Type,
				Message: rest.HttpErrorNotFound.Message,
			}, rest.HttpErrorNotFound.Status)
			return
		}

		item, err := store.Items.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				_ = httpx.WriteJSON(w, rest.ErrorResponse{
					Type:    rest.HttpErrorNotFound.Type,
					Message: rest.HttpErrorNotFound.Message,
				}, rest.HttpErrorNotFound.Status)
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
					_ = httpx.WriteJSON(w, rest.ErrorResponse{
						Type:    rest.HttpErrorNotFound.Type,
						Message: rest.HttpErrorNotFound.Message,
					}, rest.HttpErrorNotFound.Status)
					return
				}
			}
		}

		key := "items/" + id + "/" + filename

		object, err := client.GetObject(r.Context(), &s3.GetObjectInput{
			Bucket: aws.String("public-api"),
			Key:    aws.String(key),
		})
		if err != nil {
			var noSuchKey *types.NoSuchKey
			var notFound *types.NotFound
			if errors.As(err, &noSuchKey) || errors.As(err, &notFound) {
				_ = httpx.WriteJSON(w, rest.ErrorResponse{
					Type:    rest.HttpErrorNotFound.Type,
					Message: rest.HttpErrorNotFound.Message,
				}, rest.HttpErrorNotFound.Status)
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

func OaklandsItemAssetExists(client *s3.Client, id, form string) *string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prefix := "items/" + id + "/" + form + "."

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String("public-api"),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("find image %q: %v", prefix, err)
			return nil
		}

		for _, object := range page.Contents {
			key := aws.ToString(object.Key)
			extension := strings.ToLower(strings.TrimPrefix(key, prefix))

			switch extension {
			case "png", "jpg", "jpeg", "webp", "gif", "avif", "svg":
				return aws.String("/assets/" + key)
			}
		}
	}

	return nil
}
