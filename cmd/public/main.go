package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/luckfire/goblox/opencloud"
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/cmd/public/handlers/assets"
	"github.com/typical-developers/public-experience-api/cmd/public/handlers/health"
	"github.com/typical-developers/public-experience-api/cmd/public/handlers/oaklands_v1"
	"github.com/typical-developers/public-experience-api/cmd/public/handlers/static"
	"github.com/typical-developers/public-experience-api/internal/config"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/internal/logger"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
)

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
//	@Description		Abuse detection will result in a indefinite ban from accessing the API. Please reach out in our [Discord Server](https://discord.gg/typical)
//	@Description		if you have ran into this issue.
//
//	@Tag.Name			Oaklands
//	@Tag.Description	All of the available Oaklands endpoints.
//
// swagger:ignore
func main() {
	logger.Init(logger.Options{
		Environment: config.C.Environment,
		LogLevel:    config.C.LogLevel,
	})

	if strings.ToLower(config.C.Environment) == "development" {
		oaklands.SetStaging()
	}

	oc := opencloud.NewClient().WithAPIKey(config.C.OpencloudKey)
	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.C.Redis.Host, config.C.Redis.Port),
		Password: config.C.Redis.Password,
		DB:       config.C.Redis.DB,
	})

	s3Config, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(config.C.S3.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.C.S3.AccessKey,
			config.C.S3.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		panic(err)
	}
	s3Client := s3.NewFromConfig(
		s3Config,
		func(o *s3.Options) {
			o.BaseEndpoint = &config.C.S3.Endpoint
			o.UsePathStyle = true
		})

	// Index of which item images exist in S3, so handlers don't call S3 per request.
	assetIndex := assets.NewIndex(s3Client, config.C.S3.Bucket, "items/")
	if err := assetIndex.Start(context.Background(), 5*time.Minute); err != nil {
		panic(err)
	}

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

	store := data.NewOaklandsStore(redis)

	health.NewHealth(r)
	oaklands_v1.NewOaklandsV1(r, &oaklands_v1.OaklandsV1Opts{
		OpencloudClient: oc,
		Store:           store,
		Assets:          assetIndex,
	})

	r.Get("/assets/items/{id}/{filename}", assets.ServeAssets(s3Client, store))
	r.NotFound(static.ServeStatic(store))

	panic(http.ListenAndServe(":8080", r))
}
