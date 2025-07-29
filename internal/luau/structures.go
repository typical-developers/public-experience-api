package luau

import (
	"github.com/typical-developers/public-experience-api/internal/experiences"
)

// --

type OaklandsUpdateData struct {
	Translations map[string]map[string]any `json:"translations"`

	Newsletters experiences.OaklandsNewsletters           `json:"newsletters"`
	Changelogs  *map[string]experiences.OaklandsChangelog `json:"changelogs"`

	ItemDetails map[string]experiences.OaklandsItemInfo `json:"item_details"`
	StoreItems  map[string][]string                     `json:"store_items"`

	StockMarket experiences.OaklandsStockMarketV1      `json:"stock_market"`
	OreRarityV1 map[string]map[string]map[string]int64 `json:"ore_rarity_v1"`
}
