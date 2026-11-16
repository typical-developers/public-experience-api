package oaklands

import (
	"context"
	_ "embed"

	"github.com/luckfire/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/scripts"
)

var (
	UniverseID        = "3666294218"
	ProductionPlaceID = "9938675423"
	StagingPlaceID    = "13353432458"
)

var (
	//go:embed scripts/ContentSync.luau
	ContentSyncScript string
	//go:embed scripts/StockMarket.luau
	StockMarketScript string
	//go:embed scripts/ClassicShop.luau
	ClassicShopScript string
	//go:embed scripts/DailyMaterialBuffs.luau
	DailyMaterialBuffs string
)

var (
	LuauExecutionPlaceID = ProductionPlaceID
)

// SetProduction will change the `LuauExecutionPlaceID` to `ProductionPlaceID`.
// By default, ProductionPlaceID will be used for Luau scripts.
// Calling this method will set it to Production.
func SetProduction() {
	LuauExecutionPlaceID = ProductionPlaceID
}

// SetStaging will change the `LuauExecutionPlaceID` to `StagingPlaceID`.
// By default, ProductionPlaceID will be used for Luau scripts.
// Calling this method will set it to Staging.
func SetStaging() {
	LuauExecutionPlaceID = StagingPlaceID
}

type Config struct {
	// The latest newsletter override.
	NewsletterOverride string `json:"NewsletterOverride"`
	// Newsletters to not display.
	HiddenNewsletters []string `json:"HiddenNewsletters"`
}

type ContentSyncData struct {
	Translations Translations                `json:"Translations"`
	Changelogs   map[string]ChangelogVersion `json:"Changelogs"`
	Newsletters  struct {
		Latest string                `json:"Latest"`
		Pages  map[string]Newsletter `json:"Pages"`
	} `json:"Newsletters"`
	StockMarket        map[string][]StockMarketMaterial `json:"StockMarket"`
	DailyMaterialBuffs map[string][]DailyMaterialBuff   `json:"DailyMaterialBuffs"`
	ItemDetails        map[string]ItemDetails           `json:"ItemDetails"`
	StoreItems         map[string][]string              `json:"StoreItems"`
	Minimap            []MinimapLayer                   `json:"Minimap"`
	Index              map[string]map[string]IndexEntry `json:"Index"`
	Regions            map[string]Region                `json:"Regions"`
}

// GetConfig will get needed config values from the Oaklands experience config.
func GetConfig(ctx context.Context, oc *opencloud.Client) (*Config, error) {
	config, _, err := oc.Config.GetConfigWithoutMetadata(ctx, UniverseID, "InExperienceConfig")
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if newsletterOverride, ok := config.Entries["Client_LatestNewsletterOverride"].(string); ok {
		c.NewsletterOverride = newsletterOverride
	}

	if hiddenNewsletters, ok := config.Entries["Client_HideNewsletters"].([]string); ok {
		c.HiddenNewsletters = hiddenNewsletters
	} else {
		c.HiddenNewsletters = []string{}
	}

	return c, nil
}

// GetContentSync will run the ContentSync script to get new data (i.e. when Oaklands updates).
// The script is executed in the provided place version.
func GetContentSync(ctx context.Context, oc *opencloud.Client, version string) (*ContentSyncData, error) {
	script := scripts.NewScript(oc, ContentSyncScript)

	opts := scripts.ExecuteOptions{
		UniverseID:         UniverseID,
		PlaceID:            LuauExecutionPlaceID,
		Version:            &version,
		EnableBinaryOutput: new(true),
	}
	result, err := script.Execute(ctx, opts)
	if err != nil {
		return nil, err
	}

	data := new(ContentSyncData)
	if err := result.DecodeBinaryOutput(data); err != nil {
		return nil, err
	}

	return data, nil
}

type StockMarketData map[string][]StockMarketMaterial

// GetStockMarket will run the StockMarket script to get new data.
// The script is executed in the provided place version.
func GetStockMarket(ctx context.Context, oc *opencloud.Client, version string) (*StockMarketData, error) {
	script := scripts.NewScript(oc, StockMarketScript)

	opts := scripts.ExecuteOptions{
		UniverseID:         UniverseID,
		PlaceID:            LuauExecutionPlaceID,
		Version:            &version,
		EnableBinaryOutput: new(true),
	}
	result, err := script.Execute(ctx, opts)
	if err != nil {
		return nil, err
	}

	data := new(StockMarketData)
	if err := result.DecodeBinaryOutput(data); err != nil {
		return nil, err
	}

	return data, nil
}

// GetClassicShop will get the current classic shop items.
// The script is executed in the provided place version.
func GetClassicShop(ctx context.Context, oc *opencloud.Client, version string) ([]string, error) {
	script := scripts.NewScript(oc, ClassicShopScript)

	opts := scripts.ExecuteOptions{
		UniverseID: UniverseID,
		PlaceID:    LuauExecutionPlaceID,
		Version:    &version,
	}
	result, err := script.Execute(ctx, opts)
	if err != nil {
		return nil, err
	}

	data := []string{}
	if err := result.DecodeResult(&data); err != nil {
		return nil, err
	}

	return data, nil
}

type DailyMaterialBuffsData map[string][]DailyMaterialBuff

// GetDailyMaterialBuffs will get the current daily material buffs.
// The script is executed in the provided place version.
func GetDailyMaterialBuffs(ctx context.Context, oc *opencloud.Client, version string) (*DailyMaterialBuffsData, error) {
	script := scripts.NewScript(oc, DailyMaterialBuffs)

	opts := scripts.ExecuteOptions{
		UniverseID:         UniverseID,
		PlaceID:            LuauExecutionPlaceID,
		Version:            &version,
		EnableBinaryOutput: new(true),
	}
	result, err := script.Execute(ctx, opts)
	if err != nil {
		return nil, err
	}

	data := new(DailyMaterialBuffsData)
	if err := result.DecodeBinaryOutput(data); err != nil {
		return nil, err
	}

	return data, nil
}
