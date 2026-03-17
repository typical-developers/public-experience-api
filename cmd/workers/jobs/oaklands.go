package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"go.uber.org/zap"
)

type OaklandsCronJobs struct {
	opencloud *opencloud.Client
	r         oaklands.OaklandsRepository
}

type OaklandsCronJobsOpts struct {
	Cron            *cron.Cron
	OpencloudClient *opencloud.Client
	Repository      oaklands.OaklandsRepository
}

func NewOaklandsCronJobs(opts *OaklandsCronJobsOpts) {
	c := &OaklandsCronJobs{
		opencloud: opts.OpencloudClient,
		r:         opts.Repository,
	}

	// if _, err := opts.Cron.AddFunc("@every 10s", c.TestJob); err != nil {
	// 	println(err.Error())
	// }

	if _, err := opts.Cron.AddFunc("*/5 * * * *", c.GetConfig); err != nil {
		panic(err)
	}

	if _, err := opts.Cron.AddFunc("*/5 * * * *", c.CheckForUpdates); err != nil {
		panic(err)
	}

	if _, err := opts.Cron.AddFunc("0 4,10,16,22 * * *", c.RefreshStockMarkets); err != nil {
		panic(err)
	}

	if _, err := opts.Cron.AddFunc("0 4,16 * * *", c.RefreshClassicShop); err != nil {
		panic(err)
	}
}

// func (c *OaklandsCronJobs) TestJob() {
// 	panicRate := float64(0.75)

// 	time.Sleep(5 * time.Second)

// 	if rand.Float64() < panicRate {
// 		panic(fmt.Sprintf("test job panicked with rate %.2f", panicRate))
// 	}

// 	zap.L().Info("test job completed without panic",
// 		zap.String("job", "TestJob"),
// 		zap.Float64("panic_rate", panicRate),
// 	)
// }

// GetConfig will fetch for update config values from Oaklands.
// This happenes every 5 minutes.
func (c *OaklandsCronJobs) GetConfig() {
	ctx := context.Background()

	config, err := oaklands.GetConfig(ctx, c.opencloud)
	if err != nil {
		panic(err)
	}

	if err := c.r.UpdateConfig(ctx, *config); err != nil {
		panic(err)
	}

	zap.L().Info("successfully updated config",
		zap.String("job", "GetConfig"),
	)
}

// CheckForUpdates will run an API check to see if Oaklands has been updated.
// If it has been, it will fetch for new data and update ephemeral storage.
// This happenes every 5 minutes.
func (c *OaklandsCronJobs) CheckForUpdates() {
	ctx := context.Background()

	lastSync, err := c.r.GetSyncTimes(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		panic(err)
	}

	details, _, err := c.opencloud.UniverseAndPlaces.GetPlace(ctx, oaklands.UniverseID, oaklands.StagingPlaceID)
	if err != nil {
		panic(err)
	}

	updateTime, err := time.Parse(time.RFC3339, details.UpdateTime)
	if err != nil {
		panic(err)
	}

	if lastSync == nil || lastSync.LastContentSync.Before(updateTime) {
		content, err := oaklands.GetContentSync(ctx, c.opencloud)
		if err != nil {
			panic(err)
		}

		if content == nil {
			return
		}

		if err := c.r.ContentSync(ctx, *content); err != nil {
			panic(err)
		}
	}

	zap.L().Info("successfully synced game content",
		zap.String("job", "CheckForUpdates"),
	)
}

// RefreshStockMarkets will fetch and update the stock market values for each relating stock market.
// This happens every 6 hours.
func (c *OaklandsCronJobs) RefreshStockMarkets() {
	ctx := context.Background()

	stock, err := oaklands.GetStockMarket(ctx, c.opencloud)
	if err != nil {
		panic(err)
	}

	if stock == nil {
		return
	}

	if err := c.r.SetStockMarket(ctx, *stock); err != nil {
		panic(err)
	}

	zap.L().Info("successfully synced stock markets",
		zap.String("job", "RefreshStockMarkets"),
	)
}

// RefreshClassicShop will fetch and update the classic shop.
// This happens every 12 hours.
func (c *OaklandsCronJobs) RefreshClassicShop() {
	ctx := context.Background()

	items, err := oaklands.GetClassicShop(context.Background(), c.opencloud)
	if err != nil {
		panic(err)
	}

	if len(items) > 0 {
		if err := c.r.SetClassicStoreItems(ctx, items); err != nil {
			panic(err)
		}
	}

	zap.L().Info("successfully synced classic shop",
		zap.String("job", "RefreshClassicShop"),
	)
}
