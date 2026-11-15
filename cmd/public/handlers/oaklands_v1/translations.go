package oaklands_v1

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
)

//	@Router			/oaklands/v1/translations [GET]
//	@Description	List the available translations.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//
//	@Success		200				{object}	object{data=map[string]string}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetTranslations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	translations, err := o.store.Translations.Get(ctx, "en_us")
	if err != nil {
		rest.WriteRESTError(w, r, err)
		return
	}

	// Hide the name and description (i.e. `ItemName` and `ItemName_description`)
	// of items that aren't viewable yet, e.g. what an unopened gift unboxes into.
	now := time.Now()
	for id, unlockTime := range o.store.ItemUnlockTimes.Entries(ctx) {
		if !now.Before(unlockTime) {
			continue
		}

		for key := range translations {
			if key == id || strings.HasPrefix(key, id+"_") {
				delete(translations, key)
			}
		}
	}

	response := rest.Response[map[string]string]{
		Data: translations,
	}

	render.JSON(w, r, response)
}
