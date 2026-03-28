package static

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/apperror"
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
func ServeStatic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := checkForFile(r.URL.Path); err != nil {
			_ = httpx.WriteJSON(w, rest.ErrorResponse{
				Type:    err.Type,
				Message: err.Message,
			}, err.Status)

			return
		}

		http.StripPrefix("/", fsHandler).ServeHTTP(w, r)
	}
}

// OaklandsAssetExists returns the asset path when a matching file exists.
func OaklandsAssetExists(assetType, assetName string) *string {
	assetDir := filepath.ToSlash(filepath.Join(
		"assets",
		strings.Trim(assetType, "/"),
	))

	files, err := fs.ReadDir(root, assetDir)
	if err != nil {
		return nil
	}

	lookupName := strings.Trim(assetName, "/")
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
		if nameWithoutExt != lookupName {
			continue
		}

		path := "/" + filepath.ToSlash(filepath.Join(assetDir, filename))
		return &path
	}

	return nil
}
