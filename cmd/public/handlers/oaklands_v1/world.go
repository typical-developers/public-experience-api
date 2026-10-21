package oaklands_v1

import (
	"net/http"

	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

//	@Router			/oaklands/v1/world/minimap [GET]
//	@Description	Get information for the world's minimap.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//
//	@Success		200				{object}	object{data=Minimap}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetMinimap(w http.ResponseWriter, r *http.Request) {
	layers, err := o.store.Minimap.Get(r.Context())
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[Minimap]{
		Data: Minimap{
			Layers: make([]MinimapLayer, len(layers)),
		},
	}

	for i, layer := range layers {
		response.Data.Layers[i] = MinimapLayer{
			LayerID: layer.LayerID,
			X: MinimapPosition{
				Min: layer.X.Min,
				Max: layer.X.Max,
			},
			Z: MinimapPosition{
				Min: layer.Z.Min,
				Max: layer.Z.Max,
			},
			ImageIDs: layer.Images,
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
