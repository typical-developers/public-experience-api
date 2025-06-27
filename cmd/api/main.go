package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/typical-developers/public-experience-api/api"
)

func main() {
	app := fiber.New()
	api.Register(app)

	_ = app.Listen(":3000")
}
