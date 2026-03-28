package oaklands_v1

import (
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

//	@Router			/v1/oaklands/changelogs/{version} [GET]
//	@Description	Fetch a changelog version.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			version			path		string	true	"The version of changelog. For quick access to the latest changelog, use 'latest' as the value." default(latest)
//	@Param			useId			query		string	false	"Use the changelog's id instead. You do not have to provide a value to this, you can just add it as `?useId`."
//
//	@Success		200				{object}	object{data=Changelog}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetChangelog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, useId := r.URL.Query()["useId"]
	version := chi.URLParam(r, "version")
	changelog, err := o.uc.GetChangelogVersion(ctx, version, useId)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	releaseDate, _ := time.Parse(time.RFC3339, changelog.DateToISO8601())
	response := rest.Response[Changelog]{
		Data: Changelog{
			ChangelogVersion: ChangelogVersion{
				ID:      changelog.ID,
				Version: changelog.Version,
				Date:    releaseDate,
			},

			Changed: changelog.Changed,
			Added:   changelog.Added,
			Fixed:   changelog.Fixed,
		},
	}

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (o *OaklandsV1Routes) sortChangelogs(values []oaklands.Changelogs, orderBy string) []oaklands.Changelogs {
	descending := orderBy == "desc"

	sort.Slice(values, func(i, j int) bool {
		if descending {
			return values[i].Date.After(values[j].Date)
		}

		return values[i].Date.Before(values[j].Date)
	})

	return values
}

//	@Router			/v1/oaklands/changelogs [GET]
//	@Description	Fetch a list of changelog versions with their information.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			order_by		query		string	false	"The direction to order by. This will use the changelog's date to order."	default(desc)	enums(desc, asc)
//
//	@Success		200				{object}	object{data=[]ChangelogVersion}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) ListChangelogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderBy := httpx.QueryGet(r, "order_by", "desc")

	changelogs, err := o.uc.GetChangelogs(ctx)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	changelogs = o.sortChangelogs(changelogs, orderBy)
	response := rest.Response[[]ChangelogVersion]{
		Data: make([]ChangelogVersion, len(changelogs)),
	}

	for i, version := range changelogs {
		response.Data[i] = ChangelogVersion{
			ID:      version.ID,
			Version: version.Version,
			Date:    version.Date,
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
