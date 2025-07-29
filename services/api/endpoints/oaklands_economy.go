package endpoints

import (
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/typical-developers/public-experience-api/internal/cache"
	"github.com/typical-developers/public-experience-api/internal/experiences"
)

//	@Router		/v1/oaklands/economy/ore-rarity [GET]
//
//	@Tags		Oaklands, Oaklands - Economy
//
//	@Success	200	{object}	OaklandsOreRarityV1
//
// nolint:staticcheck
func OaklandsV1EconomyOreRarity(c *fiber.Ctx) error {
	var rarity *experiences.OaklandsOreRarity_v1
	if rarity = cache.GetCached[experiences.OaklandsOreRarity_v1](c.Context(), "oaklands:ore_rarity_v1", "$"); rarity == nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(ErrorAPIResponse{
			Success: false,
			Message: "Unable to fetch ore rarity.",
		})
	}

	return c.JSON(OaklandsOreRarityV1{
		Success: true,
		Data:    *rarity,
	})
}

//	@Router		/v1/oaklands/economy/stock-market [GET]
//
//	@Tags		Oaklands, Oaklands - Economy
//
//	@Param		currencyTypes	query		[]string	false	"The results with specific currencies that you want to include in the response."
//
//	@Success	200				{object}	OaklandsStockMarket
//
// nolint:staticcheck
func OaklandsV1EconomyStockMarket(c *fiber.Ctx) error {
	currencyTypes := c.Query("currencyTypes")

	var stockMarket *experiences.OaklandsStockMarketV1
	if stockMarket = cache.GetCached[experiences.OaklandsStockMarketV1](c.Context(), "oaklands:stock_market", "$"); stockMarket == nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(ErrorAPIResponse{
			Success: false,
			Message: "Unable to fetch stock market.",
		})
	}

	updatedTime := cache.Client.Get(c.Context(), "oaklands:stock_market:updated").Val()
	if updatedTime == "" {
		updatedTime = time.Now().UTC().Format(time.RFC3339)
	}

	// TODO: Sort based on CurrentDifference - LastDifference.
	// Maps work a lot differently in Go than they do in Javascript.
	// They will always be returned in a random order, so it's impossible to sort them properly.
	// if orderDifference != "" {
	// }

	if currencyTypes != "" {
		types := strings.Split(currencyTypes, ",")

		for k, v := range stockMarket.Trees {
			if slices.Contains(types, v.CurrencyType) {
				continue
			}

			delete(stockMarket.Trees, k)
		}

		for k, v := range stockMarket.Rocks {
			if slices.Contains(types, v.CurrencyType) {
				continue
			}

			delete(stockMarket.Rocks, k)
		}

		for k, v := range stockMarket.Ores {
			if slices.Contains(types, v.CurrencyType) {
				continue
			}

			delete(stockMarket.Ores, k)
		}
	}

	return c.JSON(OaklandsStockMarket{
		Success: true,
		Data: OaklandsStockMarketV1Info{
			OaklandStockMarketResetInfo: OaklandStockMarketResetInfo{
				ResetTime:   stockMarket.NextReset(),
				UpdatedTime: updatedTime,
			},
			OaklandsStockMarketV1: *stockMarket,
		},
	})
}

//	@Router		/v1/oaklands/economy/stock-market/{materialType} [GET]
//	@Tags		Oaklands, Oaklands - Economy
//
//	@Param		materialType	path		string	true	"The material type to fetch the stock market information for."
//
//	@Success	200				{object}	OaklandsStockMarketMaterial
//
// nolint:staticcheck
func OaklandsV1EconomyStockMarketMaterial(c *fiber.Ctx) error {
	materialType := strings.ToLower(c.Params("materialType"))

	var stockMarket *experiences.OaklandsStockMarketV1
	if stockMarket = cache.GetCached[experiences.OaklandsStockMarketV1](c.Context(), "oaklands:stock_market", "$"); stockMarket == nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(ErrorAPIResponse{
			Success: false,
			Message: "Unable to fetch stock market.",
		})
	}

	materials := make(map[string]*experiences.OaklandsStockMarketEntry)
	for k, v := range stockMarket.Trees {
		materials[k] = &v
	}
	for k, v := range stockMarket.Rocks {
		materials[k] = &v
	}
	for k, v := range stockMarket.Ores {
		materials[k] = &v
	}

	material := materials[materialType]
	if material == nil {
		c.Status(fiber.StatusNotFound)
		return c.JSON(ErrorAPIResponse{
			Success: false,
			Message: "No stock market information found for the specified material type.",
		})
	}

	return c.JSON(OaklandsStockMarketMaterial{
		Success: true,
		Data:    *material,
	})
}
