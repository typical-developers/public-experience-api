package api

import (
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	v1 := app.Group("/v1")

	oaklands := v1.Group("/oaklands")
	{
		oaklands.Get("/translations/keys", OaklandsTranslationKeysV1)
		oaklands.Get("/translations", OaklandsTranslationsV1)
	}
}
