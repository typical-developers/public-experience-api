package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/luckfire/goblox/legacy"
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/internal/config"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	// How long an asset can go unreferenced before its file is deleted.
	unreferencedAssetTTL = 7 * 24 * time.Hour
	// The thumbnails API allows 100 IDs per request, but more than 20 Image assets
	// in one request fails the whole request.
	thumbnailBatchSize = 20
	// The longest an asset waits between failed attempts.
	maxAssetBackoff = 24 * time.Hour
	// Hosted keys contain the file's hash, so a key's content never changes.
	hostedAssetCacheControl = "public, max-age=31536000, immutable"
)

type AssetCronJobs struct {
	legacy *legacy.Client
	// Downloads files from the CDN URLs that Roblox returns.
	http *http.Client

	redis    *redis.Client
	bucket   *storage.Bucket
	store    *data.AssetStore
	oaklands *data.OaklandsStore
}

type AssetCronJobsOpts struct {
	Handler      *Handler
	LegacyClient *legacy.Client
	Redis        *redis.Client
	Bucket       *storage.Bucket
}

func NewAssetCronJobs(opts *AssetCronJobsOpts) {
	c := &AssetCronJobs{
		legacy: opts.LegacyClient,
		http:   &http.Client{Timeout: 2 * time.Minute},

		redis:    opts.Redis,
		bucket:   opts.Bucket,
		store:    data.NewAssetStore(opts.Redis),
		oaklands: data.NewOaklandsStore(opts.Redis),
	}

	jobs := []Job{
		{
			Spec: "*/1 * * * *",
			Config: JobConfig{
				Enabled: true,
				Retry: &RetryConfig{
					RetryAttempts: 3,
					RetryDelay:    time.Duration(5 * time.Second),
				},
			},
			JobFunc: c.SyncRobloxAssets,
		},
	}

	opts.Handler.AddJobs(jobs...)
}

// SyncRobloxAssets will fetch the Roblox images referenced by synced data and host them in the bucket.
// Images are fetched in batches through the public thumbnails API, which needs no API key. Videos are
// not hosted, since downloading them needs an Open Cloud scope that can manage assets.
// Nothing is fetched until Oaklands has stored a version, which only happens once all of its data is stored.
// This happens every minute, but only does work for new assets and assets that are due for another attempt.
func (c *AssetCronJobs) SyncRobloxAssets(ctx context.Context) {
	now := time.Now()

	if _, err := c.oaklands.Version.Get(ctx); err != nil {
		if errors.Is(err, redis.Nil) {
			zap.L().Debug("Oaklands has not been synced yet, assets will not be fetched.")
			return
		}

		panic(err)
	}

	// Every referenced image ID, along with the thumbnail size that it is fetched at. Thumbnails are fitted into
	// the requested size, so newsletter banners use the largest wide size, newsletter content uses a 16:9 size,
	// and everything else uses the largest square size.
	referenced := map[string]legacy.ThumbnailSize{}
	reference := func(value string, size legacy.ThumbnailSize) {
		if i := strings.Index(value, "://"); i != -1 {
			value = value[i+3:]
		}
		if _, err := strconv.ParseUint(value, 10, 64); err != nil || value == "0" {
			return
		}

		// An image used in several places keeps the banner size first, then the newsletter content size.
		switch current := referenced[value]; {
		case current == "", current == legacy.ThumbnailSize700x700, size == legacy.ThumbnailSize1440x456:
			referenced[value] = size
		}
	}

	oaklandsConfig, err := c.oaklands.Config.Get(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		panic(err)
	}

	newsletterIDs, err := c.oaklands.Newsletter.Keys(ctx)
	if err != nil {
		panic(err)
	}
	newsletters, err := c.oaklands.Newsletter.GetMany(ctx, newsletterIDs...)
	if err != nil {
		panic(err)
	}
	for id, newsletter := range newsletters {
		// Hidden newsletters may not be released yet, so their assets are not hosted.
		if slices.Contains(oaklandsConfig.HiddenNewsletters, id) {
			continue
		}

		reference(newsletter.BannerImageId, legacy.ThumbnailSize1440x456)
		for _, section := range newsletter.Section {
			for _, content := range section.Content {
				switch v := content.(type) {
				case oaklands.NewsletterImage:
					reference(v.ImageId, legacy.ThumbnailSize768x432)
				case oaklands.NewsletterImageCarousel:
					for _, imageID := range v.ImageIds {
						reference(imageID, legacy.ThumbnailSize768x432)
					}
				}
			}
		}
	}

	minimap, err := c.oaklands.Minimap.Get(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		panic(err)
	}
	for _, layer := range minimap {
		for _, imageID := range layer.Images {
			reference(imageID, legacy.ThumbnailSize700x700)
		}
		for _, location := range layer.Locations {
			if location.MarkerImageID != nil {
				reference(*location.MarkerImageID, legacy.ThumbnailSize700x700)
			}
		}
	}

	indexKeys, err := c.oaklands.Index.Keys(ctx)
	if err != nil {
		panic(err)
	}
	index, err := c.oaklands.Index.GetMany(ctx, indexKeys...)
	if err != nil {
		panic(err)
	}
	for _, entry := range index {
		reference(entry.Image, legacy.ThumbnailSize700x700)
	}

	assetIDs, err := c.store.Roblox.Keys(ctx)
	if err != nil {
		panic(err)
	}
	manifest, err := c.store.Roblox.GetMany(ctx, assetIDs...)
	if err != nil {
		panic(err)
	}

	// Only the changed assets are written back to the manifest.
	updates := map[string]data.RobloxAsset{}
	due := []string{}
	for id := range referenced {
		asset, exists := manifest[id]
		if !exists {
			asset = data.RobloxAsset{AssetID: id, Kind: data.RobloxAssetKindImage, State: data.RobloxAssetStatePending}
		}

		// The reference time is only refreshed hourly, so every run doesn't rewrite the manifest.
		if now.Sub(asset.LastReferencedAt) > time.Hour {
			asset.LastReferencedAt = now
			updates[id] = asset
		}
		manifest[id] = asset

		if asset.State != data.RobloxAssetStateHosted && !now.Before(asset.NextAttemptAt) {
			due = append(due, id)
		}
	}

	// Assets that have not been referenced for a while are removed. This is skipped when nothing
	// is referenced, so an empty read can never remove every asset.
	deletions := []string{}
	removed := []string{}
	if len(referenced) > 0 {
		for id, asset := range manifest {
			if _, ok := referenced[id]; ok || now.Sub(asset.LastReferencedAt) < unreferencedAssetTTL {
				continue
			}

			if asset.Key != "" {
				deletions = append(deletions, asset.Key)
			}
			removed = append(removed, id)
		}
	}

	// Files that are already in the bucket, such as from a run that stopped after uploading but before storing,
	// are stored from the bucket without being downloaded or uploaded again.
	if len(due) > 0 {
		for object, err := range c.bucket.List(ctx, "roblox/") {
			if err != nil {
				panic(err)
			}

			// Keys are `roblox/<asset id>/<hash>.<extension>`.
			id, filename, ok := strings.Cut(strings.TrimPrefix(object.Key, "roblox/"), "/")
			if _, isReferenced := referenced[id]; !ok || !isReferenced || manifest[id].State == data.RobloxAssetStateHosted {
				continue
			}
			hash, extension, _ := strings.Cut(filename, ".")

			asset := manifest[id]
			asset.Kind = data.RobloxAssetKindImage
			asset.State = data.RobloxAssetStateHosted
			asset.Source = data.RobloxAssetSourceThumbnail
			asset.Key = object.Key
			asset.ContentType = "image/" + extension
			// Only the start of the hash is in the key.
			asset.Hash = hash
			asset.Size = object.Size
			asset.Attempts = 0
			asset.LastError = ""
			asset.NextAttemptAt = time.Time{}
			asset.FetchedAt = object.LastModified

			manifest[id] = asset
			updates[id] = asset

			zap.L().Debug("Asset is already in the bucket, it will not be fetched again.",
				zap.String("asset_id", id),
				zap.String("key", object.Key),
			)
		}

		due = slices.DeleteFunc(due, func(id string) bool {
			return manifest[id].State == data.RobloxAssetStateHosted
		})
	}

	fail := func(id, state string, err error, retryIn time.Duration) {
		asset := manifest[id]

		asset.State = state
		asset.Attempts++
		asset.LastError = err.Error()
		if retryIn == 0 {
			retryIn = min(time.Minute<<min(asset.Attempts, 11), maxAssetBackoff)
		}
		asset.NextAttemptAt = now.Add(retryIn)

		manifest[id] = asset
		updates[id] = asset

		zap.L().Debug("Asset could not be fetched, it will be retried.",
			zap.String("asset_id", id),
			zap.String("state", state),
			zap.Int("attempts", asset.Attempts),
			zap.Duration("retry_in", retryIn),
			zap.Error(err),
		)
	}

	// Images are batched by the thumbnail size that they are fetched at.
	images := map[legacy.ThumbnailSize][]string{}
	for _, id := range due {
		images[referenced[id]] = append(images[referenced[id]], id)
	}

	if len(due) > 0 {
		zap.L().Debug("Fetching Roblox assets.",
			zap.Int("images", len(due)),
		)
	}

	// Where each file is downloaded from.
	locations := map[string]string{}

	// Thumbnails that are still being generated are kept here, and are requested again once everything else is fetched.
	pending := map[legacy.ThumbnailSize][]string{}
	for size, ids := range images {
		for chunk := range slices.Chunk(ids, thumbnailBatchSize) {
			time.Sleep(config.C.Assets.Cooldown)

			assetIDs := make([]int64, len(chunk))
			for i, id := range chunk {
				// The IDs are already known to be numeric.
				assetIDs[i], _ = strconv.ParseInt(id, 10, 64)
			}

			thumbnails, _, err := c.legacy.Thumbnails.GetAssetThumbnails(ctx, &legacy.GetAssetThumbnailsOptions{
				AssetIDs:   assetIDs,
				Size:       size,
				Format:     legacy.ThumbnailFormatPng,
				IsCircular: new(false),
			})
			if err != nil {
				for _, id := range chunk {
					fail(id, data.RobloxAssetStateFailed, err, 0)
				}
				continue
			}

			zap.L().Debug("Fetched a batch of thumbnails.",
				zap.String("size", string(size)),
				zap.Strings("asset_ids", chunk),
				zap.Int("returned", len(thumbnails)),
			)

			returned := map[string]bool{}
			for _, thumbnail := range thumbnails {
				id := strconv.FormatInt(thumbnail.TargetID, 10)
				returned[id] = true
				err := fmt.Errorf("thumbnail is %s", thumbnail.State)

				switch {
				case thumbnail.State == legacy.ThumbnailStateCompleted && thumbnail.ImageURL != nil:
					locations[id] = *thumbnail.ImageURL
				case thumbnail.State == legacy.ThumbnailStatePending,
					thumbnail.State == legacy.ThumbnailStateInReview,
					thumbnail.State == legacy.ThumbnailStateTemporarilyUnavailable:
					pending[size] = append(pending[size], id)
				case thumbnail.State == legacy.ThumbnailStateBlocked:
					fail(id, data.RobloxAssetStateBlocked, err, maxAssetBackoff)
				default:
					// Videos also end up here, since they are not hosted.
					fail(id, data.RobloxAssetStateFailed, err, 0)
				}
			}

			for _, id := range chunk {
				if !returned[id] {
					fail(id, data.RobloxAssetStateFailed, errors.New("asset has no thumbnail"), 0)
				}
			}
		}
	}

	// The pending thumbnails are requested again, now that everything else has been fetched.
	// A thumbnail that still isn't ready is not a failure, so the asset is checked again on a later run.
	for size, ids := range pending {
		for chunk := range slices.Chunk(ids, thumbnailBatchSize) {
			time.Sleep(config.C.Assets.Cooldown)

			assetIDs := make([]int64, len(chunk))
			for i, id := range chunk {
				assetIDs[i], _ = strconv.ParseInt(id, 10, 64)
			}

			thumbnails, _, err := c.legacy.Thumbnails.GetAssetThumbnails(ctx, &legacy.GetAssetThumbnailsOptions{
				AssetIDs:   assetIDs,
				Size:       size,
				Format:     legacy.ThumbnailFormatPng,
				IsCircular: new(false),
			})
			if err != nil {
				zap.L().Debug("Could not request the pending thumbnails again.",
					zap.Strings("asset_ids", chunk),
					zap.Error(err),
				)
			}

			ready := map[string]string{}
			for _, thumbnail := range thumbnails {
				if thumbnail.State == legacy.ThumbnailStateCompleted && thumbnail.ImageURL != nil {
					ready[strconv.FormatInt(thumbnail.TargetID, 10)] = *thumbnail.ImageURL
				}
			}

			for _, id := range chunk {
				if location, ok := ready[id]; ok {
					locations[id] = location
					continue
				}

				asset := manifest[id]
				asset.State = data.RobloxAssetStatePending
				asset.NextAttemptAt = now.Add(2 * time.Minute)
				manifest[id] = asset
				updates[id] = asset

				zap.L().Debug("Thumbnail is still not ready, it will be checked again on a later run.",
					zap.String("asset_id", id),
				)
			}
		}
	}

	// The files are downloaded concurrently. Each file is uploaded as soon as it's downloaded, and stored in the
	// manifest as soon as it's uploaded, so a restart mid-run doesn't lose the assets that were already hosted.
	// S3 can't upload from a URL, so the file has to pass through the worker.
	var (
		mu        sync.Mutex
		uploads   = map[string]data.RobloxAsset{}
		downloads errgroup.Group

		failedUploads int
		uploadErr     error
	)
	downloads.SetLimit(max(config.C.Assets.Concurrency, 1))
	for id, location := range locations {
		downloads.Go(func() (err error) {
			defer func() {
				if err != nil {
					mu.Lock()
					defer mu.Unlock()
					fail(id, data.RobloxAssetStateFailed, err, 0)
				}
			}()

			// Each concurrent download waits, so downloads are spread out instead of all starting at once.
			time.Sleep(config.C.Assets.Cooldown)

			zap.L().Debug("Downloading asset.",
				zap.String("asset_id", id),
			)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
			if err != nil {
				return err
			}

			// Gzipped responses are decompressed by the transport.
			resp, err := c.http.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("download responded with %s", resp.Status)
			}

			body, err := io.ReadAll(io.LimitReader(resp.Body, config.C.Assets.MaxBytes+1))
			if err != nil {
				return err
			}
			if int64(len(body)) > config.C.Assets.MaxBytes {
				return errors.New("asset is larger than the max size")
			}

			contentType := http.DetectContentType(body)
			kind, extension, _ := strings.Cut(contentType, "/")
			if kind != data.RobloxAssetKindImage {
				return fmt.Errorf("asset is an unsupported type (%s)", contentType)
			}

			sum := sha256.Sum256(body)
			hash := hex.EncodeToString(sum[:])
			key := fmt.Sprintf("roblox/%s/%s.%s", id, hash[:12], extension)

			zap.L().Debug("Downloaded asset.",
				zap.String("asset_id", id),
				zap.String("content_type", contentType),
				zap.Int("bytes", len(body)),
			)

			if err := c.bucket.Put(ctx, storage.Object{
				Key:          key,
				Body:         body,
				ContentType:  contentType,
				CacheControl: hostedAssetCacheControl,
			}); err != nil {
				mu.Lock()
				failedUploads++
				uploadErr = err
				mu.Unlock()

				return err
			}

			zap.L().Debug("Uploaded asset.",
				zap.String("asset_id", id),
				zap.String("key", key),
			)

			mu.Lock()
			asset := manifest[id]
			mu.Unlock()

			asset.Kind = kind
			asset.State = data.RobloxAssetStateHosted
			asset.Source = data.RobloxAssetSourceThumbnail
			asset.Key = key
			asset.ContentType = contentType
			asset.Hash = hash
			asset.Size = int64(len(body))
			asset.Attempts = 0
			asset.LastError = ""
			asset.NextAttemptAt = time.Time{}
			asset.FetchedAt = now

			if err := c.store.Roblox.Set(ctx, id, asset); err != nil {
				return err
			}

			zap.L().Debug("Stored asset.",
				zap.String("asset_id", id),
				zap.String("key", asset.Key),
				zap.Int64("bytes", asset.Size),
			)

			mu.Lock()
			defer mu.Unlock()

			// The stored asset replaces any older copy that is waiting to be written at the end of the run.
			manifest[id] = asset
			updates[id] = asset
			uploads[id] = asset

			return nil
		})
	}
	_ = downloads.Wait()

	// When every upload fails, the bucket is most likely unreachable.
	if failedUploads > 0 && len(uploads) == 0 {
		panic(uploadErr)
	}

	// The hosted assets are already stored, this writes the rest of the changes.
	if err := c.store.Roblox.SetMany(ctx, updates); err != nil {
		panic(err)
	}

	// Removed assets keep their manifest entry until their file is deleted, so the deletion is retried.
	deleteErr := c.bucket.DeleteMany(ctx, deletions)
	var batchErr *storage.BatchError
	errors.As(deleteErr, &batchErr)
	if deleteErr != nil {
		zap.L().Warn("Some hosted assets could not be deleted.", zap.Error(deleteErr))
	}
	for _, id := range removed {
		if batchErr != nil && batchErr.Failed(manifest[id].Key) != nil {
			continue
		}

		if err := c.store.Roblox.Delete(ctx, id); err != nil {
			panic(err)
		}

		zap.L().Debug("Removed an asset that is no longer referenced.",
			zap.String("asset_id", id),
			zap.String("key", manifest[id].Key),
		)
	}

	if len(updates) > 0 || len(removed) > 0 {
		if err := c.redis.Publish(ctx, data.RobloxAssetsUpdatedChannel, "").Err(); err != nil {
			zap.L().Error(err.Error())
		}
	}

	if len(due) > 0 || len(removed) > 0 {
		zap.L().Info("success",
			zap.String("job", "SyncRobloxAssets"),
			zap.Int("due", len(due)),
			zap.Int("hosted", len(uploads)),
			zap.Int("removed", len(removed)),
		)
	}
}
