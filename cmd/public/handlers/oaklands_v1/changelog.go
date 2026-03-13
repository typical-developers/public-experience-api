package oaklands_v1

import (
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

//	@Router			/v1/oaklands/changelogs/{version} [GET]
//	@Description	Fetch a changelog version.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			version			path		string	true	"The version of changelog. For quick access to the latest changelog, use 'latest' as the value."
//
//	@Success		200				{object}	object{data=Changelog}
//	@Failure		503				{object}	models.ErrorResponse
//	@Failure		500				{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetChangelogVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	version := chi.URLParam(r, "version")
	changelog, err := o.uc.GetChangelogVersion(ctx, version)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "NotFound",
				Message: "This changelog does not exist.",
			}, http.StatusNotFound)

			return
		}

		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "InternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		return
	}

	releaseDate, _ := time.Parse(time.RFC3339, changelog.DateToISO8601())
	response := models.Response[Changelog]{
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
//	@Failure		503				{object}	models.ErrorResponse
//	@Failure		500				{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetChangelog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderBy := httpx.QueryGet(r, "order_by", "desc")

	changelogs, err := o.uc.GetChangelogs(ctx)
	if err != nil {
		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "InternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		return
	}

	changelogs = o.sortChangelogs(changelogs, orderBy)
	response := models.Response[[]ChangelogVersion]{
		Data: make([]ChangelogVersion, len(changelogs)),
	}

	if len(changelogs) <= 0 {
		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "ResourceNotCached",
			Message: "The requested resource is not cached. Try again in a bit.",
		}, http.StatusServiceUnavailable)

		return
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
