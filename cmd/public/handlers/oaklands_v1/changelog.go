package oaklands_v1

import (
	"cmp"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

//	@Router			/oaklands/v1/changelogs/{version} [GET]
//	@Description	Fetch a changelog version.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			version			path		string	true	"The version of changelog. For quick access to the latest changelog, use 'latest' as the value."	default(latest)
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
	var changelog data.OaklandsChangelog
	var err error
	latest := strings.EqualFold(version, "latest")
	if latest || useId {
		var id int64
		if !latest {
			id, err = strconv.ParseInt(version, 10, 32)
			if err != nil {
				rest.WriteRESTError(w, redis.Nil)
				return
			}
		}

		changelogs, fetchErr := o.store.Changelog.AllSortFunc(ctx, func(a, b data.OaklandsChangelog) int {
			if a.ID == b.ID {
				return strings.Compare(b.Version, a.Version)
			}
			return cmp.Compare(b.ID, a.ID)
		})
		if fetchErr != nil {
			rest.WriteRESTError(w, fetchErr)
			return
		}

		err = redis.Nil
		for _, entry := range changelogs {
			if latest || entry.ID == int32(id) {
				changelog = entry
				err = nil
				break
			}
		}
	} else {
		changelog, err = o.store.Changelog.Get(ctx, version)
	}
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[Changelog]{
		Data: Changelog{
			ChangelogVersion: ChangelogVersion{
				ID:      changelog.ID,
				Version: changelog.Version,
				Date:    changelog.Date,
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

//	@Router			/oaklands/v1/changelogs [GET]
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

	changelogs, err := o.store.Changelog.AllSortFunc(ctx, func(a, b data.OaklandsChangelog) int {
		comparison := a.Date.Compare(b.Date)
		if comparison == 0 {
			comparison = strings.Compare(a.Version, b.Version)
		}
		if orderBy == "desc" {
			return -comparison
		}
		return comparison
	})
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

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
