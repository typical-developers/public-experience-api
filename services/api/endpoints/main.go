package endpoints

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

//	@title				Public Experience API
//	@version			1.0
//	@description		The public experience API for fetching data from Typical Developers' experiences.
//
//	@tag.name			Oaklands
//	@tag.description	All endpoints related to Oaklands.
//
//	@tag.name			Oaklands - Translations
//	@tag.description	All endpoints related to Oaklands translations.
//
//	@tag.name			Oaklands - Economy
//	@tag.description	All endpoints related to Oaklands economy.
//
// nolint:staticcheck
func Register(app *fiber.App) {
	// Registers swagger documentation.
	app.Get("/docs/*", swagger.New(swagger.Config{
		DeepLinking:  false,
		DocExpansion: "list",
	}))

	v1 := app.Group("/v1")

	oaklands := v1.Group("/oaklands")
	{
		oaklands.Get("/translations/keys", OaklandsV1TranslationKeys)
		oaklands.Get("/translations", OaklandsV1Translations)

		oaklands.Get("/economy/ore-rarity", OaklandsV1EconomyOreRarity)
		oaklands.Get("/economy/stock-market", OaklandsV1EconomyStockMarket)
		oaklands.Get("/economy/stock-market/:materialType", OaklandsV1EconomyStockMarketMaterial)
	}
}
