package oaklands_v1

import (
	"errors"
	"net/http"
	"sort"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

func (o *OaklandsV1Routes) sortStoreItems(items []oaklands.StoreItem, sortBy, orderBy string) []oaklands.StoreItem {
	descending := orderBy == "desc"

	sort.Slice(items, func(i, j int) bool {
		if sortBy == "name" {
			if descending {
				return items[i].Name > items[j].Name
			}
			return items[i].Name < items[j].Name
		}

		if descending {
			return items[i].Price > items[j].Price
		}

		return items[i].Price < items[j].Price
	})

	return items
}

//	@Router			/v1/oaklands/stores/{store_name} [GET]
//	@Description	Fetch the items inside of a specific store.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			store_name		path		string	true	"The name of the store that you want to fetch.<br>You can fetch a list of stores by requesting [`/v1/oaklands/stores`](/#tag/oaklands/GET/v1/oaklands/stores)."
//	@Param			sort_by			query		string	false	"The value to sort by."			default(price)	enums(price, name)
//	@Param			order_by		query		string	false	"The direction to order by."	default(desc)	enums(desc, asc)
//
//	@Success		200				{object}	object{data=[]StoreItem}
//	@Failure		500				{object}	models.ErrorResponse
//	@Failure		503				{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetStore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	storeName := chi.URLParam(r, "store_name")
	sort := httpx.QueryGet(r, "sort_by", "price")
	order := httpx.QueryGet(r, "order_by", "desc")

	items, err := o.uc.GetStoreItems(ctx, storeName)
	if err != nil {
		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "InternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		return
	}

	items = o.sortStoreItems(items, sort, order)

	response := models.Response[[]StoreItem]{
		Data: make([]StoreItem, len(items)),
	}

	for i, item := range items {
		response.Data[i] = StoreItem{
			Name:        item.Name,
			DisplayName: item.DisplayName,

			// TODO: Images will be added in the future.
			// Need to figure out the best way to streamline creating them.
			Image: nil,

			Description: item.Description,
			Currency:    item.Currency,
			Price:       item.Price,
		}
	}

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/v1/oaklands/stores [GET]
//	@Description	List the available stores.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//
//	@Success		200				{object}	object{data=[]string}
//	@Failure		500				{object}	models.ErrorResponse
//	@Failure		503				{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) ListStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stores, err := o.uc.ListStores(ctx)
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

	response := models.Response[[]string]{
		Data: stores,
	}

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
