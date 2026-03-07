package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/typical-developers/public-experience-api/cmd/public/config"
	_ "github.com/typical-developers/public-experience-api/cmd/public/docs"
)

// @title           Typical Developers - Public Experience API
// @version         1.0
// @description     This is a publicly accessible API to get data in Typical Developers experiences.
//
// @host      localhost:8080
// @BasePath  /v1/
func main() {
	r := chi.NewMux()
	r.Get("/docs/*", httpSwagger.Handler())

	port := fmt.Sprintf(":%s", config.C.Port)
	panic(http.ListenAndServe(port, r))
}
