package oaklands_v1

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
)

//	@Router			/oaklands/v1/version-info [GET]
//	@Description	Fetch version information for the game.
//	@Description	It should be noted that data is fetched whenever the published version changes.
//	@Description	Internally, updates are checked for every **1 minute**.
//
//	@Tags			Oaklands
//
//	@Success		200	{object}	object{data=typicalapi.OaklandsVersionInfo}
//	@Failure		500	{object}	rest.ErrorResponse
//	@Failure		503	{object}	rest.ErrorResponse
//
// swagger:ignore
func (o OaklandsV1Routes) GetVersionInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	info, err := o.store.Version.Get(ctx)
	if err != nil {
		rest.WriteRESTError(w, r, err)
		return
	}

	response := rest.Response[VersionInfo]{
		Data: VersionInfo{
			Version:    info.Version,
			LastUpdate: info.LastUpdate,
			NextCheck:  info.NextCheck,
		},
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=15, stale-while-revalidate=5")
	render.JSON(w, r, response)
}
