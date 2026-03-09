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
	"github.com/typical-developers/public-experience-api/docs"
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
//	@Description		# Ratelimits
//	@Description		> [!NOTE]
//	@Description		> If you are constantly hitting ratelimits and need help, reach out in our community development channels in our [Discord Server](https://discord.gg/typical).<br>
//	@Description		<!---->
//	@Description		| Duration | Requests |
//	@Description		|----------|----------|
//	@Description		| Daily   | Unlimited |
//	@Description		| Per Minute    | 120 |
//	@Description		| Per Second    | 6  |
//	@Description		<!---->
//	@Description		## Headers
//	@Description		`X-RateLimit-Limit`:     The total amount of requests that can be made.<br>
//	@Description		`X-RateLimit-Remaining`: The remaining amount of requests that can be made before the rate-limit is exhausted.<br>
//	@Description		`X-RateLimit-Reset`:     The remaining amount of time for when the rate-limit resets.<br>
//
//	@Tag.Name			Oaklands
//	@Tag.Description	All of the available Oaklands endpoints.
//
// swagger:ignore
func main() {
	if config.C.ReferenceConfig.PublicHost != "" {
		docs.SwaggerInfo.Host = config.C.ReferenceConfig.PublicHost
		docs.SwaggerInfo.Schemes = []string{"https"}
	}

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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	oaklandsRepository := oaklands.NewOaklandsRepository(&oaklands.OaklandsRepositoryOpts{RedisClient: redis})
	oaklandsUsecase := oaklands.NewOaklandsUsecase(oaklandsRepository)

	oaklands_v1.NewOaklandsV1(r, &oaklands_v1.OaklandsV1Opts{
		OpencloudClient: oc,
		Usecase:         oaklandsUsecase,
	})

	r.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(docs.SwaggerInfo.ReadDoc()))
	})

	serveStatic(r)

	port := fmt.Sprintf(":%s", config.C.Port)
	panic(http.ListenAndServe(port, r))
}
