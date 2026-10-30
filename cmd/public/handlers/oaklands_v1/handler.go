package oaklands_v1

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
)

type OaklandsV1Opts struct {
	OpencloudClient *opencloud.Client
	Usecase         oaklands.OaklandsUsecase
	Store           *data.OaklandsStore
	S3              *s3.Client
}

type OaklandsV1Routes struct {
	opencloud *opencloud.Client
	uc        oaklands.OaklandsUsecase
	store     *data.OaklandsStore
	s3        *s3.Client
}

// NewOaklandsV1 will create new v1 routes for Oaklands related endpoints.
func NewOaklandsV1(r *chi.Mux, opts *OaklandsV1Opts) {
	o := &OaklandsV1Routes{
		opencloud: opts.OpencloudClient,
		uc:        opts.Usecase,
		store:     opts.Store,
		s3:        opts.S3,
	}

	r.Route("/oaklands/v1", func(r chi.Router) {
		r.Get("/version-info", o.GetVersionInfo)

		r.Route("/economy", func(r chi.Router) {
			r.Get("/daily-material-buffs", o.GetDailyMaterialBuffs)
		})

		r.Route("/world", func(r chi.Router) {
			r.Get("/minimap", o.GetMinimap)
		})

		r.Route("/changelogs", func(r chi.Router) {
			r.Get("/", o.ListChangelogs)
			r.Get("/{version}", o.GetChangelog)
		})

		r.Route("/newsletters", func(r chi.Router) {
			r.Get("/", o.ListNewsletters)
			r.Get("/{id}", o.GetNewsletter)
		})

		r.Route("/items", func(r chi.Router) {
			r.Get("/", o.ListItems)
			r.Get("/{id}", o.GetItem)
		})

		r.Route("/stores", func(r chi.Router) {
			r.Get("/", o.ListStores)
			r.Get("/{store_name}", o.GetStore)
		})

		r.Get("/translations", o.GetTranslations)
	})
}
