package static

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
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

// ServeStatic will return an http handler for serving static files.
func ServeStatic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)
		path = filepath.ToSlash(path)

		if path == "." || path == "/" {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "HttpErrorNotFound",
				Message: "This page could not be found.",
			}, http.StatusNotFound)

			return
		}
		path = strings.TrimPrefix(path, "/")

		file, err := fs.Stat(root, path)
		if err != nil {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "HttpErrorNotFound",
				Message: "This page could not be found.",
			}, http.StatusNotFound)

			return
		}

		if file.IsDir() {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "HttpErrorNotFound",
				Message: "This page could not be found.",
			}, http.StatusNotFound)

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
