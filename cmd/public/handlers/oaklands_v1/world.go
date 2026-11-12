package oaklands_v1

import (
	"net/http"
	"slices"

	"github.com/go-chi/chi"
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
//	@Success		200				{object}	object{data=typicalapi.OaklandsMinimap}
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
		locations := make([]MinimapLocation, len(layer.Locations))
		for j, location := range layer.Locations {
			markerID := ""
			if location.MarkerImageID != nil {
				markerID = *location.MarkerImageID
			}

			locations[j] = MinimapLocation{
				Name: location.Name,
				Position: MinimapLocationPosition{
					X: location.Position.X,
					Y: location.Position.Y,
					Z: location.Position.Z,
				},
				Marker:      newMedia(markerID, nil),
				Color:       location.Color,
				PropertyID:  location.PropertyID,
				Conditional: location.Conditional,
			}
		}

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
			ImageIDs:  layer.Images,
			Media:     newMediaList(layer.Images),
			Locations: locations,
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/oaklands/v1/world/regions [GET]
//	@Description	List the IDs of every region.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//
//	@Success		200				{object}	object{data=[]string}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) ListRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := o.store.Regions.Keys(r.Context())
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}
	slices.Sort(regions)

	response := rest.Response[[]string]{Data: regions}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/oaklands/v1/world/regions/{id} [GET]
//	@Description	Get information for a region.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			id				path		string	true	"The ID of the region (e.g. `MagmaCave`)."
//
//	@Success		200				{object}	object{data=typicalapi.OaklandsRegion}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetRegion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	region, err := o.store.Regions.Get(ctx, id)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	// Resources and their ores reference index entries, which are keyed by `<category>:<id>`.
	// The first len(region.Resources) keys line up with the resources.
	keys := make([]string, len(region.Resources))
	for i, resource := range region.Resources {
		keys[i] = resource.Category + ":" + resource.ID
		for _, ore := range resource.Ores {
			keys = append(keys, "ores:"+ore.ID)
		}
	}

	entries, err := o.store.Index.GetMany(ctx, keys...)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[Region]{Data: Region{
		ID:              region.ID,
		DisplayName:     region.DisplayName,
		Description:     region.Description,
		Tier:            region.Tier,
		Layer:           region.Layer,
		DiscoveryReward: region.DiscoveryReward,
		Bounds:          make([]RegionBound, len(region.Bounds)),
		Resources:       make([]RegionResource, 0, len(region.Resources)),
	}}

	if region.Island != nil {
		response.Data.Island = &RegionIsland{
			ID:          region.Island.ID,
			DisplayName: region.Island.DisplayName,
			Tier:        region.Island.Tier,
		}
	}

	for i, bound := range region.Bounds {
		response.Data.Bounds[i] = RegionBound{
			Position: RegionVector{X: bound.Position.X, Y: bound.Position.Y, Z: bound.Position.Z},
			Size:     RegionVector{X: bound.Size.X, Y: bound.Size.Y, Z: bound.Size.Z},
			Rotation: RegionVector{X: bound.Rotation.X, Y: bound.Rotation.Y, Z: bound.Rotation.Z},
		}
	}

	for i, resource := range region.Resources {
		entry, ok := entries[keys[i]]
		if !ok {
			continue
		}

		result := RegionResource{
			OaklandsIndexEntryInfo: IndexEntryInfo{
				Category: resource.Category,
				ID:       resource.ID,
				Media:    newMedia(entry.Image, nil),
			},
			DisplayName: entry.DisplayName,
			SpawnChance: resource.SpawnChance,
		}

		for _, ore := range resource.Ores {
			oreEntry, ok := entries["ores:"+ore.ID]
			if !ok {
				continue
			}

			result.Ores = append(result.Ores, RegionOre{
				ID:          ore.ID,
				DisplayName: oreEntry.DisplayName,
				Chance:      ore.Chance,
			})
		}

		response.Data.Resources = append(response.Data.Resources, result)
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
