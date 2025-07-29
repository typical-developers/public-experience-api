package endpoints

import "github.com/typical-developers/public-experience-api/internal/experiences"

type APIResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type ErrorAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type OaklandsTranslations APIResponse[map[string]string]

type OaklandsTranslationKeys APIResponse[[]string]

type OaklandsOreRarityV1 APIResponse[experiences.OaklandsOreRarity_v1]

type OaklandStockMarketResetInfo struct {
	ResetTime   string `json:"reset_time"`
	UpdatedTime string `json:"updated_time"`
}

type OaklandsStockMarketV1Info struct {
	OaklandStockMarketResetInfo
	experiences.OaklandsStockMarketV1
}

type OaklandsStockMarket APIResponse[OaklandsStockMarketV1Info]

type OaklandsStockMarketMaterial APIResponse[experiences.OaklandsStockMarketEntry]

// ---

func Ptr[T any](v T) *T {
	return &v
}
