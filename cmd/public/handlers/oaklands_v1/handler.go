package oaklands_v1

import (
	"github.com/go-chi/chi"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
)

type OaklandsV1Opts struct {
	OpencloudClient *opencloud.Client
	Usecase         oaklands.OaklandsUsecase
	Store           *data.OaklandsStore
}

type OaklandsV1Routes struct {
	opencloud *opencloud.Client
	uc        oaklands.OaklandsUsecase
	store     *data.OaklandsStore
}

// NewOaklandsV1 will create new v1 routes for Oaklands related endpoints.
func NewOaklandsV1(r *chi.Mux, opts *OaklandsV1Opts) {
	o := &OaklandsV1Routes{
		opencloud: opts.OpencloudClient,
		uc:        opts.Usecase,
		store:     opts.Store,
	}

	// Temporarily route through the old /v1/oaklands routes as well.
	// Removing before production deployment is out.
	r.Route("/v1/oaklands", func(r chi.Router) {
		//
		r.Get("/sync", o.GetSyncTimes)

		r.Route("/economy", func(r chi.Router) {
			r.Route("/stock-market", func(r chi.Router) {
				r.Get("/trees", o.GetStockMarketTrees)
				r.Get("/rocks", o.GetStockMarketRocks)
				r.Get("/ores", o.GetStockMarketOres)
			})
		})
		//

		r.Get("/version-info", o.GetVersionInfo)

		r.Route("/changelogs", func(r chi.Router) {
			r.Get("/", o.ListChangelogs)
			r.Get("/{version}", o.GetChangelog)
		})

		r.Route("/newsletters", func(r chi.Router) {
			r.Get("/", o.ListNewsletters)
			r.Get("/{id}", o.GetNewsletter)
		})

		r.Route("/stores", func(r chi.Router) {
			r.Get("/", o.ListStores)
			r.Get("/{store_name}", o.GetStore)
		})

		r.Get("/translations", o.GetTranslations)
	})
	// ----

	r.Route("/oaklands/v1", func(r chi.Router) {
		//
		r.Get("/sync", o.GetSyncTimes)

		r.Route("/economy", func(r chi.Router) {
			r.Route("/stock-market", func(r chi.Router) {
				r.Get("/trees", o.GetStockMarketTrees)
				r.Get("/rocks", o.GetStockMarketRocks)
				r.Get("/ores", o.GetStockMarketOres)
			})
		})
		//

		r.Get("/version-info", o.GetVersionInfo)

		r.Route("/changelogs", func(r chi.Router) {
			r.Get("/", o.ListChangelogs)
			r.Get("/{version}", o.GetChangelog)
		})

		r.Route("/newsletters", func(r chi.Router) {
			r.Get("/", o.ListNewsletters)
			r.Get("/{id}", o.GetNewsletter)
		})

		r.Route("/stores", func(r chi.Router) {
			r.Get("/", o.ListStores)
			r.Get("/{store_name}", o.GetStore)
		})

		r.Get("/translations", o.GetTranslations)
	})
}
