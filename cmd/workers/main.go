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
	"github.com/typical-developers/public-experience-api/internal/oaklands"
)

func main() {
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.DefaultLogger),
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
