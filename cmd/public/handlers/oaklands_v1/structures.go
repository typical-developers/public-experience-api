package oaklands_v1

import "time"

// @name	Oaklands.V1.MaterialValues
type MaterialValues struct {
	// The type of material.
	Type string `json:"type"`
	// The base value for the material.
	BaseValue float32 `json:"base_value"`
	// The current value of the material.
	CurrentValue float32 `json:"current_value"`
} //	@name	Oaklands.V1.MaterialValues

type StockMarket struct {
	// The identifier for the material.
	Name string `json:"name"`
	// The display name of the material.
	DisplayName string `json:"display_name"`
	// The currency of the material.
	CurrencyType string `json:"currency_type"`
	// The current multiplier.
	CurrentMultiplier float32 `json:"current_multiplier"`
	// The multipler from the last time the stock refreshed.
	LastMultiplier float32 `json:"last_multiplier"`
	// The values of the different types for the materials
	Values []MaterialValues `json:"values"`
} //	@name	Oaklands.V1.StockMarket

type SyncInfo struct {
	//	The last time that the resource was updated.
	LastSync time.Time `json:"last_sync"`
} //	@name	Oaklands.V1.UpdatedInfo

type SyncdInfoWithReset struct {
	//	The last time that the resource was updated.
	LastSync time.Time `json:"last_sync"`
	// The next time that the resource will be updated
	NextSync time.Time `json:"next_sync"`
} //	@name	Oaklands.V1.UpdatedInfoWithReset

type Sync struct {
	// The last time that data from Oaklands that can be updated from experience changes was synced.
	LastContentSync time.Time `json:"last_content_sync"`
	// Sync information for the stock market.
	StockMarket SyncdInfoWithReset `json:"stock_market"`
} //	@name	Oaklands.V1.Sync
