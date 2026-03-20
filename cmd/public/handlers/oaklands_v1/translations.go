package oaklands_v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

func (o *OaklandsV1Routes) prefixTranslations(translations *map[string]string, prefixes ...string) *map[string]string {
	filteredTranslations := make(map[string]string, 0)
	for _, prefix := range prefixes {
		for key, value := range *translations {
			if !strings.HasPrefix(key, prefix) {
				continue
			}

			filteredTranslations[key] = value
		}
	}

	return &filteredTranslations
}

//	@Router			/v1/oaklands/translations [GET]
//	@Description	List the available translations.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			prefix			query		string	false	"Return a subset of keys that have a prefix (i.e. using `ItemName` would return values for the keys `ItemName` and `ItemName_description`).<br>It is suggested to only return the keys you are needing."
//
//	@Success		200				{object}	object{data=map[string]string}
//	@Failure		500				{object}	models.ErrorResponse
//	@Failure		503				{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetTranslations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	prefix := httpx.QueryGetAll(r, "prefix")

	translations, err := o.uc.GetTranslations(ctx, "en_us")
	if err != nil {
		if errors.Is(err, redis.Nil) {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "ResourceNotCached",
				Message: "The requested resource is not cached. Try again in a bit.",
			}, http.StatusServiceUnavailable)

			return
		}

		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "InternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		return
	}

	if len(prefix) > 0 {
		translations = o.prefixTranslations(translations, prefix...)
	}

	response := models.Response[map[string]string]{
		Data: *translations,
	}

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
