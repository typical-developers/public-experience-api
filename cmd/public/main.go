package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/cmd/public/config"
	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
	"github.com/typical-developers/public-experience-api/cmd/public/handlers/oaklands_v1"
	"github.com/typical-developers/public-experience-api/internal/apperror"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

var (
	HttpErrorNotFound = apperror.NewAppError("HttpErrorNotFound", "This page could not be found.", http.StatusNotFound, nil)
)

// serveStatic will serve static files on the root.
func serveStatic(r chi.Router) {
	root := "static"
	fs := http.FileServer(http.Dir(root))

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(root, r.URL.Path)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    HttpErrorNotFound.Type,
				Message: HttpErrorNotFound.Message,
			}, HttpErrorNotFound.Status)
			return
		}

		if info.IsDir() {
			if _, err := os.Stat(filepath.Join(path, "index.html")); os.IsNotExist(err) {
				_ = httpx.WriteJSON(w, models.ErrorResponse{
					Type:    HttpErrorNotFound.Type,
					Message: HttpErrorNotFound.Message,
				}, HttpErrorNotFound.Status)
				return
			}
		}

		http.StripPrefix("/", fs).ServeHTTP(w, r)
	}))
}

//	@Title				Typical Developers - Public Experience API
//
//	@Description		This is the official, publicly accessible, API to get data in Typical Developers' experiences.
//	@Description		# Notes
//	@Description		- All dates and timestamps are returned in [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) format and are set as a UTC timezone.<br>
//	@Description		---
//	@Description		# Caching
//	@Description		All responses will return an `ETag` header, which is an MD5 hash of the response. You are able to use this hash to verify
//	@Description		if there's updated content by supplying a `If-None-Match` in the requests header for the next request you make.<br>
//	@Description		- Status `200 - OK` will be returned if the resource has been updated.
//	@Description		- Status `304 - Not Modified` will be returned if the resource has not been updated.
//	@Description		---
//	@Description		# Ratelimits
//	@Description		At the moment, the API does not have any rate limits. Please use the API responsibly by following good practices.
//	@Description		Abuse detection will result in a indefinite ban from accessing the API. Please reach out in our [Discord Server](https.discord.gg/typical)
//	@Description		if you have ran into this issue.
//
//	@Tag.Name			Oaklands
//	@Tag.Description	All of the available Oaklands endpoints.
//
// swagger:ignore
func main() {
	oc := opencloud.NewClient().WithAPIKey(config.C.OpencloudKey)
	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.C.Redis.Host, config.C.Redis.Port),
		Password: config.C.Redis.Password,
		DB:       config.C.Redis.DB,
	})

	r := chi.NewMux()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "If-None-Match"},
		ExposedHeaders:   []string{"Last-Modified", "ETag", "Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	oaklandsRepository := oaklands.NewOaklandsRepository(&oaklands.OaklandsRepositoryOpts{RedisClient: redis})
	oaklandsUsecase := oaklands.NewOaklandsUsecase(oaklandsRepository)

	oaklands_v1.NewOaklandsV1(r, &oaklands_v1.OaklandsV1Opts{
		OpencloudClient: oc,
		Usecase:         oaklandsUsecase,
	})

	serveStatic(r)

	port := fmt.Sprintf(":%s", config.C.Port)
	panic(http.ListenAndServe(port, r))
}
