package oaklands

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

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

	// The sections that could not be used and the reason why, keyed by section name.
	// A skipped section is left empty and should not replace what is already stored.
	Skipped map[string]string `json:"-"`
}

// ErrNoSections is returned when a content sync finished but none of its sections could be used.
var ErrNoSections = errors.New("content sync returned no usable sections")

// supportedSectionVersions is the newest version of each section that can be decoded.
// A section that reports a newer version is skipped until support for it is added here.
var supportedSectionVersions = map[string]int{
	"Translations":       1,
	"Changelogs":         1,
	"Newsletters":        1,
	"StockMarket":        1,
	"DailyMaterialBuffs": 1,
	"ItemDetails":        1,
	"StoreItems":         1,
	"Minimap":            1,
	"Index":              1,
	"Regions":            1,
}

// contentSyncExport is the output of the ContentSync script.
type contentSyncExport struct {
	// The data of each section that loaded, keyed by section name.
	Sections map[string]json.RawMessage `json:"Sections"`
	// The version of each section that loaded, keyed by section name.
	Versions map[string]int `json:"Versions"`
	// The reason a section did not load, keyed by section name.
	Errors map[string]string `json:"Errors"`
}

// decodeSection will decode a single section of the export into dst.
// If the section is missing, has an unsupported version, or does not decode,
// dst is left untouched and the reason is added to skipped.
func decodeSection[T any](export *contentSyncExport, skipped map[string]string, name string, dst *T) {
	raw, ok := export.Sections[name]
	if !ok {
		if reason, ok := export.Errors[name]; ok {
			skipped[name] = reason
		} else {
			skipped[name] = "not exported by this place version"
		}

		return
	}

	if export.Versions[name] > supportedSectionVersions[name] {
		skipped[name] = fmt.Sprintf("section version %d is not supported", export.Versions[name])
		return
	}

	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		skipped[name] = err.Error()
		return
	}

	*dst = value
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
//
// Each section is decoded on its own. A section that is missing or can not be decoded
// is listed in ContentSyncData.Skipped instead of failing the whole sync.
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

	export := new(contentSyncExport)
	if err := result.DecodeBinaryOutput(export); err != nil {
		return nil, err
	}

	data := &ContentSyncData{
		Skipped: map[string]string{},
	}
	decodeSection(export, data.Skipped, "Translations", &data.Translations)
	decodeSection(export, data.Skipped, "Changelogs", &data.Changelogs)
	decodeSection(export, data.Skipped, "Newsletters", &data.Newsletters)
	decodeSection(export, data.Skipped, "StockMarket", &data.StockMarket)
	decodeSection(export, data.Skipped, "DailyMaterialBuffs", &data.DailyMaterialBuffs)
	decodeSection(export, data.Skipped, "ItemDetails", &data.ItemDetails)
	decodeSection(export, data.Skipped, "StoreItems", &data.StoreItems)
	decodeSection(export, data.Skipped, "Minimap", &data.Minimap)
	decodeSection(export, data.Skipped, "Index", &data.Index)
	decodeSection(export, data.Skipped, "Regions", &data.Regions)

	if len(data.Skipped) == len(supportedSectionVersions) {
		return nil, ErrNoSections
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
