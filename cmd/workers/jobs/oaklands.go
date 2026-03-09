package jobs

import (
	"context"
	"time"

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
	h := &OaklandsCronJobs{
		opencloud: opts.OpencloudClient,
		r:         opts.Repository,
	}

	h.CheckForUpdates()

	if _, err := opts.Cron.AddFunc("@every 5m", h.CheckForUpdates); err != nil {
		panic(err)
	}
}

// CheckForUpdates will run an API check to see if Oaklands has been updated.
// If it has been, it will fetch for new data and update ephemeral storage.
func (c *OaklandsCronJobs) CheckForUpdates() {
	ctx := context.Background()

	lastSync, err := c.r.GetLastSyncTime(ctx)
	if err != nil {
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

	if lastSync == nil || lastSync.Before(updateTime) {
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
