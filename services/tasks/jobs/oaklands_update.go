package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/cache"
	"github.com/typical-developers/public-experience-api/internal/experiences"
	"github.com/typical-developers/public-experience-api/internal/luau"
)

func getCachedNewsletterKeys(ctx context.Context) map[string]bool {
	keys := make(map[string]bool)

	existingKeys := cache.Client.Keys(ctx, "oaklands:newsletters:*")
	for _, key := range existingKeys.Val() {
		key := strings.Split(key, ":")[2]
		keys[key] = true
	}

	return keys
}

func getCachedChangelogKeys(ctx context.Context) map[string]bool {
	keys := make(map[string]bool)

	existingKeys := cache.Client.Keys(ctx, "oaklands:changelogs:*")
	for _, key := range existingKeys.Val() {
		key := strings.Split(key, ":")[2]
		keys[key] = true
	}

	return keys
}

func inputUpload(ctx context.Context) (*string, error) {
	var buff bytes.Buffer
	data := OaklandsUpdateBinaryInput{
		CachedNewsletters: getCachedNewsletterKeys(ctx),
		CachedChangelogs:  getCachedChangelogKeys(ctx),
	}

	err := json.NewEncoder(&buff).Encode(data)
	if err != nil {
		return nil, err
	}
	input := buff.Bytes()

	binaryInput, _, err := experiences.Opencloud.LuauExecution.CreateLuauExecutionSessionTaskBinaryInput(ctx, "3666294218", opencloud.LuauExecutionSessionTaskBinaryInputCreate{
		Size: opencloud.Pointer(len(input)),
	})
	if err != nil {
		return nil, err
	}

	_, err = experiences.Opencloud.LuauExecution.UploadLuauExecutionSessionTaskBinaryInput(ctx, binaryInput.UploadURI, input)
	if err != nil {
		return nil, err
	}

	return &binaryInput.Path, nil
}

func CheckOaklandsUpdates() {
	ctx := context.Background()
	lastUpdatedEpoch := cache.Client.Get(ctx, "oaklands:last_updated").Val()
	parsedLastUpdatedEpoch, err := strconv.ParseInt(lastUpdatedEpoch, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to parse last updated epoch")
		return
	}

	details, _, err := experiences.Opencloud.UniverseAndPlaces.GetPlace(ctx, "3666294218", "9938675423")
	if err != nil {
		log.WithError(err).Error("Failed to fetch place details")
		return
	}
	lastUpdatedTime, err := time.Parse(time.RFC3339, details.UpdateTime)
	if err != nil {
		log.WithError(err).Error("Failed to parse last updated time")
		return
	}
	if lastUpdatedTime.Unix() == parsedLastUpdatedEpoch {
		log.Info("No recent Oaklands update detected.")
		return
	}

	binaryInput, err := inputUpload(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to upload binary input")
		return
	}

	_, binary, err := luau.Run[any, luau.OaklandsUpdateData](ctx, "3666294218", "9938675423", luau.OaklandsUpdatesScriptPath, &luau.RunOptions{
		EnableBinaryOutput: opencloud.Pointer(true),
		BinaryInput:        binaryInput,
	})
	if err != nil {
		log.WithError(err).Error("Failed to run Oaklands updates")
		return
	}

	// Cache Translations
	translations := make(map[string]map[string]any)
	for lang, keys := range binary.Translations {
		if lang == "ALTKEYS" {
			continue
		}

		translations[lang] = make(map[string]any)
		maps.Copy(translations[lang], keys)
	}

	cache.SetCached(ctx, "oaklands:translations", "$", translations, nil)
	// ---

	// Cache Newsletters
	if binary.Newsletters.Pages != nil {
		for page, pageData := range *binary.Newsletters.Pages {
			cache.SetCached(
				ctx,
				fmt.Sprintf("oaklands:newsletters:%s", page), "$",
				pageData, nil,
			)
		}
	}
	// ---

	// Cache Changelogs
	if binary.Changelogs != nil {
		for version, versionData := range *binary.Changelogs {
			cache.SetCached(
				ctx,
				fmt.Sprintf("oaklands:changelogs:%s", version), "$",
				versionData, nil,
			)
		}
	}
	// ---

	// Cache Item Details
	for item, itemData := range binary.ItemDetails {
		cache.SetCached(
			ctx,
			fmt.Sprintf("oaklands:items:%s", item), "$",
			itemData, nil,
		)
	}
	// ---

	// Cache Stock Market
	cache.SetCached(ctx, "oaklands:stock_market", "$", binary.StockMarket, nil)
	// ---

	// Cache Store Items
	for store, items := range binary.StoreItems {
		cache.SetCached(
			ctx,
			fmt.Sprintf("oaklands:store_items:%s", store), "$",
			items, nil,
		)
	}
	// ---

	// Cache Ore Rarity (V1)
	cache.SetCached(ctx, "oaklands:ore_rarity_v1", "$", binary.OreRarityV1, nil)
	// ---

	cache.Client.Set(ctx, "oaklands:last_updated", lastUpdatedTime.Unix(), 0)
}
