package main

import (
	"github.com/gofiber/fiber/v2"
	_ "github.com/typical-developers/public-experience-api/internal/docs"
	"github.com/typical-developers/public-experience-api/services/api/config"
	"github.com/typical-developers/public-experience-api/services/api/endpoints"
)

func main() {
	config.Load()

	app := fiber.New()
	endpoints.Register(app)

	_ = app.Listen(":3000")
}
