package oaklands_v1

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
)

type OaklandsV1Opts struct {
	OpencloudClient *opencloud.Client
	Usecase         oaklands.OaklandsUsecase
}

type OaklandsV1Routes struct {
	opencloud *opencloud.Client
	uc        oaklands.OaklandsUsecase
}

func cacheControl() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
			next.ServeHTTP(w, r)
		})
	}
}

// NewOaklandsV1 will create new v1 routes for Oaklands related endpoints.
func NewOaklandsV1(r *chi.Mux, opts *OaklandsV1Opts) {
	o := &OaklandsV1Routes{
		opencloud: opts.OpencloudClient,
		uc:        opts.Usecase,
	}

	r.Route("/v1/oaklands", func(r chi.Router) {
		r.Get("/sync", o.GetSyncTimes)

		r.Route("/economy", func(r chi.Router) {
			r.Use(cacheControl())

			r.Route("/stock-market", func(r chi.Router) {
				r.Get("/trees", o.GetStockMarketTrees)
				r.Get("/rocks", o.GetStockMarketRocks)
				r.Get("/ores", o.GetStockMarketOres)
			})
		})
	})
}
