package oaklands_v1

import (
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi"
	"github.com/typical-developers/public-experience-api/cmd/public/handlers/assets"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/data"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

func (o *OaklandsV1Routes) storeRestockTime(name string, now time.Time) *time.Time {
	switch name {
	case "ClassicStore":
		now = now.UTC()

		next := time.Date(
			now.Year(), now.Month(), now.Day(),
			4, 0, 0, 0, time.UTC,
		)

		for !next.After(now) {
			next = next.Add(12 * time.Hour)
		}

		return &next
	default:
		return nil
	}
}

func (o *OaklandsV1Routes) sortStoreItems(items []StoreItem, sortBy, orderBy string) []StoreItem {
	descending := orderBy == "desc"

	sort.SliceStable(items, func(i, j int) bool {
		if sortBy == "name" || items[i].Price == items[j].Price {
			if descending {
				return items[i].ID > items[j].ID
			}
			return items[i].ID < items[j].ID
		}

		if descending {
			return items[i].Price > items[j].Price
		}

		return items[i].Price < items[j].Price
	})

	return items
}

//	@Router			/oaklands/v1/stores/{store_name} [GET]
//	@Description	Fetch the items inside of a specific store.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			store_name		path		string	true	"The name of the store that you want to fetch.<br>You can fetch a list of stores by requesting [`/oaklands/v1/stores`](/#tag/oaklands/GET/oaklands/v1/stores)."
//	@Param			sort_by			query		string	false	"The value to sort by."			default(price)	enums(price, name)
//	@Param			order_by		query		string	false	"The direction to order by."	default(desc)	enums(desc, asc)
//
//	@Success		200				{object}	object{restock_at=time.Time,data=[]StoreItem}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetStore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	storeName := chi.URLParam(r, "store_name")
	sortBy := httpx.QueryGet(r, "sort_by", "price")
	order := httpx.QueryGet(r, "order_by", "desc")

	stock, err := o.store.StoreStock.Get(ctx, storeName)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	unlockTimes, err := o.store.ItemUnlockTimes.GetMany(ctx, stock...)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}
	now := time.Now()

	items, err := o.store.Items.GetMany(ctx, stock...)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := StoreResponse{
		RestockAt: o.storeRestockTime(storeName, now),
		Response: rest.Response[[]StoreItem]{
			Data: make([]StoreItem, 0, len(items)),
		},
	}

	for _, name := range stock {
		if unlockAt, ok := unlockTimes[name]; ok && now.Before(unlockAt) {
			continue
		}
		item, ok := items[name]
		if !ok {
			continue
		}

		details := StoreItem{
			ID:          item.Name,
			DisplayName: item.DisplayName,
			Description: item.Description,
		}
		for _, form := range item.Forms {
			storeInfo, ok := form.Data.(data.OaklandsItemFormStore)
			if !ok {
				continue
			}

			details.Image = assets.OaklandsItemAssetExists(o.s3, item.Name, form.FormType)
			details.Currency = storeInfo.Currency
			details.Price = storeInfo.Price
			break
		}
		response.Data = append(response.Data, details)
	}
	response.Data = o.sortStoreItems(response.Data, sortBy, order)

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/oaklands/v1/stores [GET]
//	@Description	List the available stores.
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
func (o *OaklandsV1Routes) ListStores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stores, err := o.store.StoreStock.Keys(ctx)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	sort.Strings(stores)
	response := rest.Response[[]string]{Data: stores}
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
