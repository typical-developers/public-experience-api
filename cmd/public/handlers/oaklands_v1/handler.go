package oaklands_v1

import (
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

// NewOaklandsV1 will create new v1 routes for Oaklands related endpoints.
func NewOaklandsV1(r *chi.Mux, opts *OaklandsV1Opts) {
	o := &OaklandsV1Routes{
		opencloud: opts.OpencloudClient,
		uc:        opts.Usecase,
	}

	r.Route("/v1/oaklands", func(r chi.Router) {
		r.Get("/sync", o.GetSyncTimes)

		r.Route("/economy", func(r chi.Router) {
			r.Route("/stock-market", func(r chi.Router) {
				r.Get("/trees", o.GetStockMarketTrees)
				r.Get("/rocks", o.GetStockMarketRocks)
				r.Get("/ores", o.GetStockMarketOres)
			})
		})

		r.Route("/changelogs", func(r chi.Router) {
			r.Get("/", o.GetChangelog)
			r.Get("/{version}", o.GetChangelogVersion)
		})

		r.Route("/newsletters", func(r chi.Router) {
			r.Get("/", o.GetNewsletters)
			r.Get("/{id}", o.GetNewsletter)
		})

		r.Route("/stores", func(r chi.Router) {
			r.Get("/", o.ListStores)
			r.Get("/{store_name}", o.GetStore)
		})
	})
}
