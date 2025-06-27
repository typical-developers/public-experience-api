package api

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
		oaklands.Get("/translations/keys", OaklandsTranslationKeysV1)
		oaklands.Get("/translations", OaklandsTranslationsV1)
	}
}
