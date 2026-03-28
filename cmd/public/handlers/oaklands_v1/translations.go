package oaklands_v1

import (
	"net/http"
	"strings"

	"github.com/typical-developers/public-experience-api/cmd/public/rest"
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
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetTranslations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	prefix := httpx.QueryGetAll(r, "prefix")

	translations, err := o.uc.GetTranslations(ctx, "en_us")
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	if len(prefix) > 0 {
		translations = o.prefixTranslations(translations, prefix...)
	}

	response := rest.Response[map[string]string]{
		Data: *translations,
	}

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
