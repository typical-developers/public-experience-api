package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/typical-developers/public-experience-api/cmd/public/config"
	_ "github.com/typical-developers/public-experience-api/cmd/public/docs"
)

// serveStatic will serve static files on the root.
func serveStatic(r chi.Router) {
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/*", http.StripPrefix("/", fs))
}

// @Title           Typical Developers - Public Experience API
// @Description     This is a publicly accessible API to get data in Typical Developers experiences.
//
// @Tag.Name Oaklands
// @Tag.Description Oaklands related endpoints.
func main() {
	r := chi.NewMux()
	r.Get("/docs/*", httpSwagger.Handler())
	serveStatic(r)

	port := fmt.Sprintf(":%s", config.C.Port)
	panic(http.ListenAndServe(port, r))
}
