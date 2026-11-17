package oaklands_v1

import (
	"strings"

	"github.com/go-chi/chi"
	"github.com/luckfire/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/cmd/public/handlers/assets"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/config"
	"github.com/typical-developers/public-experience-api/internal/data"
)

type OaklandsV1Opts struct {
	OpencloudClient *opencloud.Client
	Store           *data.OaklandsStore
	Assets          *assets.Index
}

type OaklandsV1Routes struct {
	opencloud *opencloud.Client
	store     *data.OaklandsStore
	assets    *assets.Index
}

// newMedia creates media from a Roblox asset id and/or the path to an asset hosted by the API.
// Returns nil when neither exist.
func newMedia(assetID string, path *string) *Media {
	if i := strings.Index(assetID, "://"); i != -1 {
		assetID = assetID[i+3:]
	}

	media := &Media{}
	if assetID != "" && assetID != "0" {
		media.AssetID = &assetID
	}
	if path != nil {
		media.URL = new(strings.TrimSuffix(config.C.AssetBaseURL, "/") + *path)
	}

	if media.AssetID == nil && media.URL == nil {
		return nil
	}

	return media
}

// newMediaList creates media for each Roblox asset id, skipping any that are empty.
func newMediaList(assetIDs []string) []Media {
	list := make([]Media, 0, len(assetIDs))
	for _, id := range assetIDs {
		if media := newMedia(id, nil); media != nil {
			list = append(list, *media)
		}
	}

	return list
}

// NewOaklandsV1 will create new v1 routes for Oaklands related endpoints.
func NewOaklandsV1(r *chi.Mux, opts *OaklandsV1Opts) {
	o := &OaklandsV1Routes{
		opencloud: opts.OpencloudClient,
		store:     opts.Store,
		assets:    opts.Assets,
	}

	r.Route("/oaklands/v1", func(r chi.Router) {
		r.Get("/version-info", o.GetVersionInfo)

		r.Group(func(r chi.Router) {
			r.Use(rest.ETag)

			r.Route("/economy", func(r chi.Router) {
				r.Get("/daily-material-buffs", o.GetDailyMaterialBuffs)
			})

			r.Route("/world", func(r chi.Router) {
				r.Get("/minimap", o.GetMinimap)
				r.Get("/regions", o.ListRegions)
				r.Get("/regions/{id}", o.GetRegion)
			})

			r.Route("/changelogs", func(r chi.Router) {
				r.Get("/", o.ListChangelogs)
				r.Get("/{version}", o.GetChangelog)
			})

			r.Route("/newsletters", func(r chi.Router) {
				r.Get("/", o.ListNewsletters)
				r.Get("/{id}", o.GetNewsletter)
			})

			r.Route("/index", func(r chi.Router) {
				r.Get("/", o.ListIndexCategories)
				r.Get("/{category}", o.ListIndex)
				r.Get("/{category}/{id}", o.GetIndexEntry)
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
	})
}
