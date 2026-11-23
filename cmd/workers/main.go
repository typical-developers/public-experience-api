package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/luckfire/goblox/auth"
	"github.com/luckfire/goblox/legacy"
	"github.com/luckfire/goblox/opencloud"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/typical-developers/public-experience-api/cmd/workers/jobs"
	"github.com/typical-developers/public-experience-api/internal/config"
	"github.com/typical-developers/public-experience-api/internal/logger"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/internal/storage"

	"go.uber.org/zap"
)

type CustomCronLogger struct {
	*zap.SugaredLogger
}

func (l CustomCronLogger) Info(msg string, keysAndValues ...any) {
	l.Infow(msg, keysAndValues...)
}

func (l CustomCronLogger) Error(err error, msg string, keysAndValues ...any) {
	args := append([]any{"error", err}, keysAndValues...)
	l.Errorw(msg, args...)
}

func main() {
	logger.Init(logger.Options{
		Environment: config.C.Environment,
		LogLevel:    config.C.LogLevel,
	})

	if strings.ToLower(config.C.Environment) == "development" {
		oaklands.SetStaging()
	}

	l := CustomCronLogger{zap.L().Sugar()}
	c := cron.New(
		cron.WithLogger(l),
		cron.WithLocation(time.UTC),
		cron.WithChain(
			cron.SkipIfStillRunning(l),
			cron.Recover(l),
		),
	)

	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.C.Redis.Host, config.C.Redis.Port),
		Password: config.C.Redis.Password,
		DB:       config.C.Redis.DB,
	})

	s3, err := storage.NewS3(context.Background(), config.C.S3)
	if err != nil {
		panic(err)
	}
	bucket := storage.NewBucket(s3, config.C.S3.Bucket, storage.WithConcurrency(config.C.Assets.Concurrency))
	if err := bucket.Ping(context.Background()); err != nil {
		panic(err)
	}

	oc := opencloud.NewClient().WithCredential(auth.IntrospectedAPIKey(config.C.OpencloudKey, 10*time.Minute))

	jobHandler := jobs.NewHandler(jobs.HandlerOpts{
		Cron:   c,
		Logger: l,
	})

	jobs.NewOaklandsCronJobs(&jobs.OaklandsCronJobsOpts{
		Handler:         jobHandler,
		OpencloudClient: oc,
		Redis:           redis,
	})

	jobs.NewAssetCronJobs(&jobs.AssetCronJobsOpts{
		Handler:      jobHandler,
		LegacyClient: legacy.NewClient(),
		Redis:        redis,
		Bucket:       bucket,
	})

	c.Start()
	runtime.Goexit()
}
