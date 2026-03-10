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

type ContentSyncData struct {
	Changelogs  map[string]ChangelogVersion `json:"Changelogs"`
	Newsletters struct {
		Latest string                `json:"Latest"`
		Pages  map[string]Newsletter `json:"Pages"`
	} `json:"Newsletters"`
	StockMarket map[string][]StockMarketMaterial `json:"StockMarket"`
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
