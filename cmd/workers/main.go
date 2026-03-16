package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/cmd/workers/config"
	"github.com/typical-developers/public-experience-api/cmd/workers/jobs"
	_ "github.com/typical-developers/public-experience-api/internal/logger"
	"github.com/typical-developers/public-experience-api/internal/oaklands"

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
	l := CustomCronLogger{zap.L().Sugar()}

	c := cron.New(
		cron.WithLogger(l),
		cron.WithLocation(time.UTC),
		cron.WithChain(
			cron.Recover(l),
		),
	)

	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.C.Redis.Host, config.C.Redis.Port),
		Password: config.C.Redis.Password,
		DB:       config.C.Redis.DB,
	})

	oc := opencloud.NewClient().WithAPIKey(config.C.OpencloudKey)
	oaklandsRepository := oaklands.NewOaklandsRepository(&oaklands.OaklandsRepositoryOpts{RedisClient: redis})

	jobs.NewOaklandsCronJobs(&jobs.OaklandsCronJobsOpts{
		Cron:            c,
		OpencloudClient: oc,
		Repository:      oaklandsRepository,
	})

	c.Start()
	runtime.Goexit()
}
