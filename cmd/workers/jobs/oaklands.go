package jobs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/config"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/internal/scripts"
	"go.uber.org/zap"
)

type OaklandsCronJobs struct {
	opencloud *opencloud.Client
	store     *data.OaklandsStore
}

type OaklandsCronJobsOpts struct {
	Handler         *Handler
	OpencloudClient *opencloud.Client
	Redis           *redis.Client
}

func NewOaklandsCronJobs(opts *OaklandsCronJobsOpts) {
	c := &OaklandsCronJobs{
		opencloud: opts.OpencloudClient,
		store:     data.NewOaklandsStore(opts.Redis),
	}

	jobs := []Job{
		{
			Spec: "*/5 * * * *",
			Config: JobConfig{
				Enabled: true,
				Retry: &RetryConfig{
					RetryAttempts: 3,
					RetryDelay:    time.Duration(5 * time.Second),
				},
			},
			JobFunc: c.GetConfig,
		},
		{
			Spec: "*/1 * * * *",
			Config: JobConfig{
				Enabled: true,
				Retry: &RetryConfig{
					RetryAttempts: 3,
					RetryDelay:    time.Duration(5 * time.Second),
				},
			},
			JobFunc: c.CheckForUpdates,
		},
		{
			Spec: "0 4,16 * * *",
			Config: JobConfig{
				Enabled: true,
				Retry: &RetryConfig{
					RetryAttempts: 3,
					RetryDelay:    time.Duration(5 * time.Second),
				},
			},
			JobFunc: c.RefreshClassicShop,
		},
	}

	opts.Handler.AddJobs(jobs...)
}

// GetConfig will fetch for update config values from Oaklands.
// This happenes every 5 minutes.
func (c *OaklandsCronJobs) GetConfig(ctx context.Context) {
	config, err := oaklands.GetConfig(ctx, c.opencloud)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	if err := c.store.Config.Set(ctx, data.OaklandsConfig{
		NewsletterOverride: config.NewsletterOverride,
		HiddenNewsletters:  config.HiddenNewsletters,
	}); err != nil {
		zap.L().Error(err.Error())
		return
	}

	zap.L().Info("success",
		zap.String("job", "GetConfig"),
	)
}

func (c *OaklandsCronJobs) getRecentPublishedVersion(ctx context.Context) (*opencloud.AssetVersion, error) {
	opts := opencloud.Options{
		MaxPageSize: new(50),
	}

	var version *opencloud.AssetVersion
	for {
		versions, resp, err := c.opencloud.Assets.GetAssetVersions(ctx, oaklands.LuauExecutionPlaceID, &opts)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("http status returned %s", resp.Status)
		}

		for _, assetVersion := range versions.AssetVersions {
			if !config.C.Oaklands.AlwaysExecuteInLatest && !assetVersion.Published {
				continue
			}

			version = &assetVersion
			break
		}
		if version != nil {
			break
		}

		opts.PageToken = &versions.NextPageToken
	}

	return version, nil
}

// CheckForUpdates will run an API check to see if Oaklands has been updated.
// If it has been, it will fetch for new data and update ephemeral storage.
// This happenes every 5 minutes.
func (c *OaklandsCronJobs) CheckForUpdates(ctx context.Context) {
	now := time.Now()

	versionDetails, err := c.getRecentPublishedVersion(ctx)
	if err != nil {
		panic(err)
	}

	version := versionDetails.Path[strings.LastIndex(versionDetails.Path, "/")+1:]
	lastVersion, err := c.store.Version.Get(ctx)
	if err != nil && err != redis.Nil {
		panic(err)
	}

	if dead, err := c.store.DeadVersions.Get(ctx, fmt.Sprintf("%s:%s", oaklands.LuauExecutionPlaceID, version)); err == nil && dead {
		zap.L().Debug("Latest published version is flagged dead, will not fetch.",
			zap.String("version", version),
		)
		return
	}

	if version == lastVersion.Version {
		zap.L().Debug("No new published version, will not refetch.")

		if err := c.store.Version.Set(ctx, data.OaklandsVersion{
			Version:    lastVersion.Version,
			LastUpdate: lastVersion.LastUpdate,
			LastCheck:  now,
			NextCheck:  now.Truncate(time.Minute).Add(time.Minute),
		}); err != nil {
			panic(err)
		}

		return
	}

	oaklands.SetPlaceVersionOverride(version)
	content, err := oaklands.GetContentSync(ctx, c.opencloud)
	if err != nil {
		if errors.Is(err, scripts.ErrModuleErrored) {
			zap.L().Debug("The most recently fetched version is dead.",
				zap.String("version", version),
			)

			// set to last alive version
			if lastVersion.Version != "" {
				oaklands.SetPlaceVersionOverride(lastVersion.Version)
			}

			_ = c.store.DeadVersions.Set(ctx, fmt.Sprintf("%s:%s", oaklands.LuauExecutionPlaceID, version), true)

			return
		}

		panic(err)
	}

	place, _, err := c.opencloud.UniverseAndPlaces.GetPlace(ctx, oaklands.UniverseID, oaklands.LuauExecutionPlaceID)
	if err != nil {
		panic(err)
	}

	updateTime, err := time.Parse(time.RFC3339, place.UpdateTime)
	if err != nil {
		panic(err)
	}

	if err := c.store.Version.Set(ctx, data.OaklandsVersion{
		Version:    version,
		LastUpdate: updateTime,
		LastCheck:  now,
		NextCheck:  now.Truncate(time.Minute).Add(time.Minute),
	}); err != nil {
		panic(err)
	}

	translations := make(map[string]map[string]string, len(content.Translations))
	for locale, values := range content.Translations {
		translations[strings.ToLower(locale)] = values
	}
	if err := c.store.Translations.SetMany(ctx, translations); err != nil {
		panic(err)
	}

	items := map[string]data.OaklandsItem{}
	for id, item := range content.ItemDetails {
		forms := []data.OaklandsItemForm{}

		for _, form := range item.Forms {
			forms = append(forms, data.OaklandsItemForm{
				FormType:    form.FormType,
				ConvertType: form.ConvertType,
				Data:        form.Data,
			})
		}

		items[id] = data.OaklandsItem{
			Name:        item.Name,
			DisplayName: item.DisplayName,
			Description: item.Description,
			Forms:       forms,
		}
	}
	if err := c.store.Items.SetMany(ctx, items); err != nil {
		panic(err)
	}

	if err := c.store.StoreStock.SetMany(ctx, content.StoreItems); err != nil {
		panic(err)
	}

	changelogs := map[string]data.OaklandsChangelog{}
	for version, changelog := range content.Changelogs {
		date, err := time.Parse(time.RFC3339, changelog.DateToISO8601())
		if err != nil {
			zap.L().Error(err.Error())
			continue
		}

		changelogs[version] = data.OaklandsChangelog{
			ID:      changelog.ID,
			Version: changelog.Version,
			Date:    date,
			Changed: changelog.Changed,
			Added:   changelog.Added,
			Fixed:   changelog.Fixed,
		}
	}
	if err := c.store.Changelog.SetMany(ctx, changelogs); err != nil {
		panic(err)
	}

	newsletters := map[string]data.OaklandsNewsletter{}
	for id, newsletter := range content.Newsletters.Pages {
		date, err := time.Parse(time.RFC3339, newsletter.DateToISO8601())
		if err != nil {
			panic(err)
		}

		sections := make([]data.OaklandsNewsletterSection, len(newsletter.Sections))
		for i, section := range newsletter.Sections {
			sections[i] = data.OaklandsNewsletterSection{
				Header:  section.Header,
				Content: section.Content,
			}
		}

		newsletters[id] = data.OaklandsNewsletter{
			Date:          date,
			Header:        newsletter.Header,
			Subheader:     newsletter.Subheader,
			BannerImageId: newsletter.BannerImageId,
			Section:       sections,
		}
	}
	if err := c.store.Newsletter.SetMany(ctx, newsletters); err != nil {
		panic(err)
	}
	if content.Newsletters.Latest != "" {
		if err := c.store.NewsletterLatest.Set(ctx, content.Newsletters.Latest); err != nil {
			panic(err)
		}
	}

	minimap := []data.OaklandsMinimap{}
	for _, layer := range content.Minimap {
		minimap = append(minimap, data.OaklandsMinimap{
			LayerID: layer.LayerID,
			X: data.OaklandsMinimapPosition{
				Min: layer.X.Min,
				Max: layer.X.Max,
			},
			Z: data.OaklandsMinimapPosition{
				Min: layer.Z.Min,
				Max: layer.Z.Max,
			},
			Images: layer.ImageIDs,
		})
	}
	c.store.Minimap.Set(ctx, minimap)
}

// RefreshClassicShop will fetch and update the classic shop.
// This happens every 12 hours.
func (c *OaklandsCronJobs) RefreshClassicShop(ctx context.Context) {
	if oaklands.PlaceVersionOverride == nil {
		return
	}

	items, err := oaklands.GetClassicShop(context.Background(), c.opencloud)
	if err != nil {
		panic(err)
	}

	if err := c.store.StoreStock.Set(ctx, "ClassicStore", items); err != nil {
		panic(err)
	}
}
