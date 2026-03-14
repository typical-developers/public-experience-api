package oaklands

import (
	"context"
	_ "embed"

	"github.com/typical-developers/goblox/opencloud"
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
)

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

	result, err := script.Execute(ctx, scripts.ExecuteOptions{
		UniverseID:         UniverseID,
		PlaceID:            StagingPlaceID,
		EnableBinaryOutput: new(true),
	})
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

	result, err := script.Execute(ctx, scripts.ExecuteOptions{
		UniverseID:         UniverseID,
		PlaceID:            StagingPlaceID,
		EnableBinaryOutput: new(true),
	})

	if err != nil {
		return nil, err
	}

	data := new(StockMarketData)
	if err := result.DecodeBinaryOutput(data); err != nil {
		return nil, err
	}

	return data, nil
}
