package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
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

	if _, err := opts.Cron.AddFunc("@every 5m", c.GetConfig); err != nil {
		panic(err)
	}

	if _, err := opts.Cron.AddFunc("@every 5m", c.CheckForUpdates); err != nil {
		panic(err)
	}

	if _, err := opts.Cron.AddFunc("0 4,10,16,22 * * *", c.RefreshStockMarkets); err != nil {
		panic(err)
	}
}

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
}
