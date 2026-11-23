package oaklands_v1

import (
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/data"
)

//	@Router			/oaklands/v1/items [GET]
//	@Description	List the available items, without the item details.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//
//	@Success		200				{object}	object{data=[]typicalapi.OaklandsItemInfo}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o OaklandsV1Routes) ListItems(w http.ResponseWriter, r *http.Request) {
	ids, err := o.store.Items.Keys(r.Context())
	if err != nil {
		rest.WriteRESTError(w, r, err)
		return
	}
	slices.Sort(ids)

	unlockTimes, err := o.store.ItemUnlockTimes.GetMany(r.Context(), ids...)
	if err != nil {
		rest.WriteRESTError(w, r, err)
		return
	}

	items, err := o.store.Items.GetMany(r.Context(), ids...)
	if err != nil {
		rest.WriteRESTError(w, r, err)
		return
	}

	response := rest.Response[[]ItemInfo]{
		Data: make([]ItemInfo, 0, len(ids)),
	}
	now := time.Now()
	for _, id := range ids {
		unlockTime, ok := unlockTimes[id]
		if ok && now.Before(unlockTime) {
			continue
		}

		item, ok := items[id]
		if !ok {
			continue
		}

		response.Data = append(response.Data, ItemInfo{
			ID:          id,
			DisplayName: item.DisplayName,
		})
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=60, stale-while-revalidate=60")
	render.JSON(w, r, response)
}

//	@Router			/oaklands/v1/items/{id} [GET]
//	@Description	Get information for an item.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			id				path		string	true	"The id of the newsletter. For quick access to the latest newsletter, use 'latest' as the value."
//
//	@Success		200				{object}	object{data=typicalapi.OaklandsItem}
//	@Failure		404				{object}	rest.ErrorResponse
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o OaklandsV1Routes) GetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := o.store.Items.Get(r.Context(), id)
	if err != nil {
		rest.WriteLookupError(w, r, err)
		return
	}
	if unlockTime, err := o.store.ItemUnlockTimes.Get(r.Context(), id); err == nil {
		if time.Now().Before(unlockTime) {
			// Items that aren't unlocked yet are hidden from the client.
			rest.WriteRESTError(w, r, rest.ErrNotFound)
			return
		}
	}
	hasUnopenedGift := false
	imageForms := []string{}
	for _, form := range item.Forms {
		if gift, ok := form.Data.(data.OaklandsItemFormGift); ok {
			imageForms = append(imageForms, "gift")
			if !gift.IsOpened() {
				hasUnopenedGift = true
			}
		}
	}
	imageForms = append(imageForms, "item", "box", "store")
	for _, form := range item.Forms {
		imageForms = append(imageForms, form.FormType)
	}
	var image *string
	for _, form := range imageForms {
		if hasUnopenedGift && form != "store" && form != "box" && form != "gift" {
			continue
		}
		image = o.assets.URL(item.Name, form)
		if image != nil {
			break
		}
	}

	response := rest.Response[Item]{Data: Item{
		ID:          item.Name,
		DisplayName: item.DisplayName,
		Description: item.Description,
		Media:       o.newMedia("", image),
		Forms:       make([]ItemForm, 0, len(item.Forms)),
	}}

	for _, form := range item.Forms {
		if hasUnopenedGift && form.FormType != "store" && form.FormType != "box" && form.FormType != "gift" {
			continue
		}
		formImage := o.assets.URL(item.Name, form.FormType)
		result := ItemForm{
			Type:        form.FormType,
			Media:       o.newMedia("", formImage),
			ConvertType: form.ConvertType,
		}

		switch info := form.Data.(type) {
		case data.OaklandsItemFormStore:
			response.Data.Currency = &info.Currency
			response.Data.Price = &info.Price
		case data.OaklandsItemFormGift:
			response.Data.UnboxTime = new(time.Unix(info.UnboxTime, 0))
			if !info.IsOpened() {
				response.Data.Forms = append(response.Data.Forms, result)
				continue
			}

			resultID := item.Name
			if info.ConvertID != nil && *info.ConvertID != "" {
				resultID = *info.ConvertID
				result.ConvertID = info.ConvertID
			}

			for _, unboxType := range info.UnboxTypes {
				unboxImage := o.assets.URL(resultID, unboxType.FormType)
				result.UnboxTypes = append(result.UnboxTypes, ItemFormUnboxType{
					Type:        unboxType.FormType,
					Media:       o.newMedia("", unboxImage),
					ConvertType: unboxType.ConvertType,
				})
			}

			response.Data.Forms = append(response.Data.Forms, result)
			continue
		case data.OaklandsItemFormItem:
			if info.Stats != nil {
				response.Data.Stats = &info.Stats
			}
		}

		response.Data.Forms = append(response.Data.Forms, result)
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=60, stale-while-revalidate=60")
	render.JSON(w, r, response)
}
