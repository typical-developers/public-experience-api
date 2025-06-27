package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/typical-developers/public-experience-api/api"
	_ "github.com/typical-developers/public-experience-api/internal/docs"
)

func main() {
	app := fiber.New()
	api.Register(app)

	_ = app.Listen(":3000")
}
