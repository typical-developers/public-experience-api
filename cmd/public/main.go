package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/typical-developers/public-experience-api/cmd/public/config"
	_ "github.com/typical-developers/public-experience-api/cmd/public/docs"
)

// serveStatic will serve static files on the root.
func serveStatic(r chi.Router) {
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/*", http.StripPrefix("/", fs))
}

//	@Title				Typical Developers - Public Experience API
//
//	@Description		This is the official, publicly accessible, API to get data in Typical Developers' experiences.
//	@Description		# Notes
//	@Description		- All dates and timestamps are returned in [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) format and are set as a UTC timezone.<br>
//	@Description		---
//	@Description		# Ratelimits
//	@Description		If you are constantly hitting ratelimits and need help, reach out in our community development channels in our [Discord Server](https://discord.gg/typical).<br>
//	@Description		<!---->
//	@Description		| Duration | Requests |
//	@Description		|----------|----------|
//	@Description		| Daily   | Unlimited |
//	@Description		| Per Minute    | 120 |
//	@Description		| Per Second    | 6  |
//	@Description		<!---->
//	@Description		## Headers
//	@Description		`X-RateLimit-Limit`:     The total amount of requests that can be made.<br>
//	@Description		`X-RateLimit-Remaining`: The remaining amount of requests that can be made before the rate-limit is exhausted.<br>
//	@Description		`X-RateLimit-Reset`:     The remaining amount of time for when the rate-limit resets.<br>
//
//	@Tag.Name			Oaklands
//	@Tag.Description	All of the available Oaklands endpoints.
//
// swagger:ignore
func main() {
	r := chi.NewMux()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	serveStatic(r)

	port := fmt.Sprintf(":%s", config.C.Port)
	panic(http.ListenAndServe(port, r))
}
