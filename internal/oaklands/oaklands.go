package oaklands

import (
	"context"
	_ "embed"

	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/scripts"
	"go.uber.org/zap"
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
)

var (
	PlaceVersionOverride *string
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

// SetPlaceVersionOverride will set a place version to use instead of defaulting to the most recent version.
func SetPlaceVersionOverride(version string) {
	PlaceVersionOverride = &version
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
	StockMarket map[string][]StockMarketMaterial `json:"StockMarket"`
	ItemDetails map[string]ItemDetails           `json:"ItemDetails"`
	StoreItems  map[string][]string              `json:"StoreItems"`
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
func GetContentSync(ctx context.Context, oc *opencloud.Client) (*ContentSyncData, error) {
	script := scripts.NewScript(oc, ContentSyncScript)

	opts := scripts.ExecuteOptions{
		UniverseID:         UniverseID,
		PlaceID:            LuauExecutionPlaceID,
		EnableBinaryOutput: new(true),
	}
	if PlaceVersionOverride != nil {
		zap.L().Debug("version",
			zap.String("message", "PlaceVersionOverride is not nil, running script in specified version."),
			zap.String("version", *PlaceVersionOverride),
		)

		opts.Version = PlaceVersionOverride
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
func GetStockMarket(ctx context.Context, oc *opencloud.Client) (*StockMarketData, error) {
	script := scripts.NewScript(oc, StockMarketScript)

	opts := scripts.ExecuteOptions{
		UniverseID:         UniverseID,
		PlaceID:            LuauExecutionPlaceID,
		EnableBinaryOutput: new(true),
	}
	if PlaceVersionOverride != nil {
		zap.L().Debug("version",
			zap.String("message", "PlaceVersionOverride is not nil, running script in specified version."),
			zap.String("version", *PlaceVersionOverride),
		)

		opts.Version = PlaceVersionOverride
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
func GetClassicShop(ctx context.Context, oc *opencloud.Client) ([]string, error) {
	script := scripts.NewScript(oc, ClassicShopScript)

	opts := scripts.ExecuteOptions{
		UniverseID: UniverseID,
		PlaceID:    LuauExecutionPlaceID,
	}
	if PlaceVersionOverride != nil {
		zap.L().Debug("version",
			zap.String("message", "PlaceVersionOverride is not nil, running script in specified version."),
			zap.String("version", *PlaceVersionOverride),
		)

		opts.Version = PlaceVersionOverride
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
