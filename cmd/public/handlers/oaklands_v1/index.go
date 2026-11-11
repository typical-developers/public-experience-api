package oaklands_v1

import (
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

//	@Router			/oaklands/v1/index [GET]
//	@Description	List the material index categories.
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
func (o *OaklandsV1Routes) ListIndexCategories(w http.ResponseWriter, r *http.Request) {
	keys, err := o.store.Index.Keys(r.Context())
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	// Entries are keyed by `<category>:<id>`.
	categories := map[string]struct{}{}
	for _, key := range keys {
		category, _, _ := strings.Cut(key, ":")
		categories[category] = struct{}{}
	}

	response := rest.Response[[]string]{
		Data: slices.AppendSeq(make([]string, 0, len(categories)), maps.Keys(categories)),
	}
	slices.Sort(response.Data)

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/oaklands/v1/index/{category} [GET]
//	@Description	List the material index entries in a category, without the entry details.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			category		path		string	true	"The category of the entries (trees, rocks, ores, creatures or forageables)."
//
//	@Success		200				{object}	object{data=[]typicalapi.OaklandsIndexEntryInfo}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) ListIndex(w http.ResponseWriter, r *http.Request) {
	category := strings.ToLower(chi.URLParam(r, "category"))

	entries, err := o.store.Index.SortedValues(r.Context(), func(a, b data.OaklandsIndexEntry) int {
		return strings.Compare(a.Name, b.Name)
	})
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[[]IndexEntryInfo]{
		Data: []IndexEntryInfo{},
	}
	for _, entry := range entries {
		if entry.Category != category {
			continue
		}

		response.Data = append(response.Data, IndexEntryInfo{
			Category: entry.Category,
			ID:       entry.Name,
			Media:    newMedia(entry.Image, nil),
		})
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/oaklands/v1/index/{category}/{id} [GET]
//	@Description	Get a material index entry.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			category		path		string	true	"The category of the entry (trees, rocks, ores, creatures or forageables)."
//	@Param			id				path		string	true	"The ID of the entry (e.g. `BirchTree` or `Rose Gold`)."
//
//	@Success		200				{object}	object{data=typicalapi.OaklandsIndexEntry}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetIndexEntry(w http.ResponseWriter, r *http.Request) {
	category := strings.ToLower(chi.URLParam(r, "category"))
	id := chi.URLParam(r, "id")

	entry, err := o.store.Index.Get(r.Context(), category+":"+id)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[IndexEntry]{Data: IndexEntry{
		OaklandsIndexEntryInfo: IndexEntryInfo{
			Category: entry.Category,
			ID:       entry.Name,
			Media:    newMedia(entry.Image, nil),
		},
		DisplayName: entry.DisplayName,
		Description: entry.Description,
		Sections:    make([]IndexSection, len(entry.Sections)),
		Enemy:       entry.Enemy,
		Tier:        entry.Tier,
		Stats:       entry.Stats,
		Variants:    make(map[string]IndexVariant, len(entry.Variants)),
		Drops:       make([]IndexDrop, len(entry.Drops)),
	}}

	for i, section := range entry.Sections {
		response.Data.Sections[i] = IndexSection{
			ID:          section.ID,
			DisplayName: section.DisplayName,
		}
	}

	for variantType, variant := range entry.Variants {
		response.Data.Variants[variantType] = IndexVariant{
			DisplayName: variant.DisplayName,
			Value:       variant.Value,
		}
	}

	for i, drop := range entry.Drops {
		response.Data.Drops[i] = IndexDrop{
			Name:        drop.Name,
			DisplayName: drop.DisplayName,
			Chance:      drop.Chance,
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
