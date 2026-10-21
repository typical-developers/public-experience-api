package oaklands_v1

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

func maxValueBase(values []oaklands.StockMarketValue) float32 {
	if len(values) == 0 {
		return 0
	}

	max := values[0].BaseValue
	for i := 1; i < len(values); i++ {
		if values[i].BaseValue > max {
			max = values[i].BaseValue
		}
	}

	return max
}

func maxValueCurrent(values []oaklands.StockMarketValue) float32 {
	if len(values) == 0 {
		return 0
	}

	max := values[0].CurrentValue
	for i := 1; i < len(values); i++ {
		if values[i].CurrentValue > max {
			max = values[i].CurrentValue
		}
	}

	return max
}

func (o *OaklandsV1Routes) sortStockMarketMaterials(values []oaklands.StockMarketMaterial, sortBy, orderBy string) []oaklands.StockMarketMaterial {
	descending := orderBy == "desc"

	sort.Slice(values, func(i, j int) bool {
		if sortBy == "name" {
			if descending {
				return values[i].Name > values[j].Name
			}
			return values[i].Name < values[j].Name
		}

		var a, b float32

		switch sortBy {
		case "last_multiplier":
			a = values[i].LastMultiplier
			b = values[j].LastMultiplier
		case "current_multiplier":
			a = values[i].CurrentMultiplier
			b = values[j].CurrentMultiplier
		case "values.current_value":
			a = maxValueCurrent(values[i].Values)
			b = maxValueCurrent(values[j].Values)
		case "values.base_value":
			a = maxValueBase(values[i].Values)
			b = maxValueBase(values[j].Values)
		default:
			return false
		}

		if descending {
			return a > b
		}

		return a < b
	})

	return values
}

func newStockMarketResponse(materials []oaklands.StockMarketMaterial) rest.Response[[]StockMarket] {
	response := rest.Response[[]StockMarket]{
		Data: make([]StockMarket, len(materials)),
	}

	for i, material := range materials {
		response.Data[i] = StockMarket{
			Name:         material.Name,
			DisplayName:  material.DisplayName,
			CurrencyType: material.CurrencyType,

			CurrentMultiplier: material.CurrentMultiplier,
			LastMultiplier:    material.LastMultiplier,

			Values: make([]MaterialValues, len(material.Values)),
		}

		for j, value := range material.Values {
			response.Data[i].Values[j] = MaterialValues{
				Type:         value.Type,
				BaseValue:    value.BaseValue,
				CurrentValue: value.CurrentValue,
			}
		}
	}

	return response
}

func (o *OaklandsV1Routes) writeStockMarket(
	w http.ResponseWriter,
	r *http.Request,
	getter func(ctx context.Context) (*oaklands.StockMarket, error),
) {
	ctx := r.Context()

	sort := httpx.QueryGet(r, "sort_by", "current_multiplier")
	order := httpx.QueryGet(r, "order_by", "desc")

	market, err := getter(ctx)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	market.Stock = o.sortStockMarketMaterials(market.Stock, sort, order)
	response := newStockMarketResponse(market.Stock)

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	w.Header().Set("Last-Modified", market.LastSync.Format(http.TimeFormat))
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (o *OaklandsV1Routes) GetStockMarketTrees(w http.ResponseWriter, r *http.Request) {
	o.writeStockMarket(w, r, o.uc.GetStockMarketTrees)
}

func (o *OaklandsV1Routes) GetStockMarketRocks(w http.ResponseWriter, r *http.Request) {
	o.writeStockMarket(w, r, o.uc.GetStockMarketRocks)
}

func (o *OaklandsV1Routes) GetStockMarketOres(w http.ResponseWriter, r *http.Request) {
	o.writeStockMarket(w, r, o.uc.GetStockMarketOres)
}

//	@Router			/oaklands/v1/economy/daily-material-buffs [GET]
//	@Description	Get all of the daily material buff values (tree, rocks, ores).
//
//	@Tags			Oaklands
//
//	@Success		200	{object}	object{refresh_at=time.Time,data=map[string][]MaterialBuff}
//	@Failure		503	{object}	rest.ErrorResponse
//	@Failure		500	{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetDailyMaterialBuffs(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	response := MaterialBuffResponse{
		RefreshAt: time.Date(
			now.Year(), now.Month(), now.Day()+1,
			4, 0, 0, 0, time.UTC,
		),
		Response: rest.Response[map[string][]MaterialBuff]{
			Data: map[string][]MaterialBuff{},
		},
	}

	for materialType, materials := range o.store.DailyMaterialBuffs.Entries(r.Context()) {
		mt := strings.ToLower(materialType)

		response.Data[mt] = make([]MaterialBuff, len(materials))
		for i, material := range materials {
			values := make([]MaterialBuffValue, len(material.Values))

			for i, value := range material.Values {
				values[i] = MaterialBuffValue{
					Type:        value.Type,
					BaseValue:   value.BaseValue,
					BuffedValue: value.BuffedValue,
				}
			}

			response.Data[mt][i] = MaterialBuff{
				ResourceID:   material.Name,
				DisplayName:  material.DisplayName,
				CurrencyType: material.CurrencyType,
				Multiplier:   material.Multiplier,
				Values:       values,
			}
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
