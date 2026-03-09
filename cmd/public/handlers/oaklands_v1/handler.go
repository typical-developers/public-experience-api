package oaklands_v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/httprate"
	"github.com/typical-developers/goblox/opencloud"
	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

type OaklandsV1Opts struct {
	OpencloudClient *opencloud.Client
	Usecase         oaklands.OaklandsUsecase
}

type OaklandsV1Routes struct {
	opencloud *opencloud.Client
	uc        oaklands.OaklandsUsecase
}

func oaklandsRateLimit(limit int, window time.Duration, bucket string) func(http.Handler) http.Handler {
	return httprate.Limit(limit, window,
		httprate.WithKeyFuncs(
			httprate.Key("oaklands:"+bucket),
			httprate.KeyByIP,
			httprate.KeyByEndpoint,
		),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "HttpErrorTooManyRequest",
				Message: "You are being rate limited.",
			}, http.StatusTooManyRequests)
		}),
	)
}

// NewOaklandsV1 will create new v1 routes for Oaklands related endpoints.
func NewOaklandsV1(r *chi.Mux, opts *OaklandsV1Opts) {
	o := &OaklandsV1Routes{
		opencloud: opts.OpencloudClient,
		uc:        opts.Usecase,
	}

	r.Route("/v1/oaklands", func(r chi.Router) {
		r.Use(oaklandsRateLimit(120, 1*time.Minute, "per-minute"))
		r.Use(oaklandsRateLimit(6, 1*time.Second, "per-second"))

		r.Route("/economy", func(r chi.Router) {
			r.Route("/stock-market", func(r chi.Router) {
				r.Get("/trees", o.GetStockMarketTrees)
				r.Get("/rocks", o.GetStockMarketRocks)
				r.Get("/ores", o.GetStockMarketOres)
			})
		})
	})
}
