package oaklands_v1

import "time"

type MaterialValues struct {
	// The type of material.
	Type string `json:"type"`
	// The base value for the material.
	BaseValue float32 `json:"base_value"`
	// The current value of the material.
	CurrentValue float32 `json:"current_value"`
}

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
}

type UpdatedInfo struct {
	//	The last time that the resource was updated.
	LastUpdated time.Time `json:"last_updated"`
}

type UpdatedInfoWithReset struct {
	//	The last time that the resource was updated.
	LastUpdated time.Time `json:"last_updated"`
	// The next time that the resource will be updated
	NextUpdate time.Time `json:"next_update"`
}

type UpdateInfo struct {
	Translations UpdatedInfo `json:"translations"`
}
