package oaklands

import (
	"context"

	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/scripts"
)

var (
	UniverseID        = "3666294218"
	ProductionPlaceID = "9938675423"
	StagingPlaceID    = "13353432458"
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
	script, err := scripts.NewScriptFromFile(oc, "../../internal/oaklands/scripts/ContentSync.luau")
	if err != nil {
		return nil, err
	}

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
