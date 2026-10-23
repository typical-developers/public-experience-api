package oaklands_v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/cmd/public/handlers/static"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

//	@Router			/oaklands/v1/items [GET]
//	@Description	Get a list of available items.
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
func (o OaklandsV1Routes) ListItems(w http.ResponseWriter, r *http.Request) {
	items, err := o.store.Items.Keys(r.Context())
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	unlockTimes, err := o.store.ItemUnlockTimes.GetMany(r.Context(), items...)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	visibleItems := items[:0]
	now := time.Now()
	for _, id := range items {
		unlockTime, ok := unlockTimes[id]
		if ok && now.Before(unlockTime) {
			continue
		}

		visibleItems = append(visibleItems, id)
	}
	items = visibleItems

	response := rest.Response[[]string]{Data: items}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=60, stale-while-revalidate=60")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/oaklands/v1/items/{id} [GET]
//	@Description	Get information for an item.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			id				path		string	true	"The id of the newsletter. For quick access to the latest newsletter, use 'latest' as the value."
//
//	@Success		200				{object}	object{data=Item}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o OaklandsV1Routes) GetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := o.store.Items.Get(r.Context(), id)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}
	if unlockTime, err := o.store.ItemUnlockTimes.Get(r.Context(), id); err == nil {
		if time.Now().Before(unlockTime) {
			// Too lazy to make a new error type for this, just
			// throw it off as resource not cached by sending redis.Nil
			rest.WriteRESTError(w, redis.Nil)
			return
		}
	}

	response := rest.Response[Item]{Data: Item{
		ID:          item.Name,
		DisplayName: item.DisplayName,
		Description: item.Description,

		Forms: make([]ItemForm, len(item.Forms)),
	}}

	for i, form := range item.Forms {
		if form.ConvertType != nil {
			response.Data.Image = static.OaklandsAssetExists(*form.ConvertType, item.Name)
		}

		switch info := form.Data.(type) {
		case data.OaklandsItemFormStore:
			response.Data.Currency = &info.Currency
			response.Data.Price = &info.Price

			response.Data.Forms[i] = ItemForm{
				Type:        form.FormType,
				ConvertType: form.ConvertType,
			}
		case data.OaklandsItemFormGift:
			response.Data.UnboxTime = new(time.Unix(info.UnboxTime, 0))

			response.Data.Forms[i] = ItemForm{
				Type:         form.FormType,
				ConvertType:  form.ConvertType,
				HasConvertID: info.ConvertID != nil,
				UnboxTypes:   info.UnboxTypes,
			}

			if info.IsOpened() {
				response.Data.Forms[i].ConvertID = info.ConvertID
			}
		case data.OaklandsItemFormItem:
			response.Data.Forms[i] = ItemForm{
				Type: form.FormType,
			}

			if info.Stats != nil {
				response.Data.Stats = &info.Stats
			}
		default:
			response.Data.Forms[i] = ItemForm{
				Type:        form.FormType,
				ConvertType: form.ConvertType,
			}
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=60, stale-while-revalidate=60")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
