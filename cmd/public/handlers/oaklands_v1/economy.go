package oaklands_v1

import (
	"context"
	"net/http"
	"sort"

	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
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

func newStockMarketResponse(materials []oaklands.StockMarketMaterial) models.Response[[]StockMarket] {
	response := models.Response[[]StockMarket]{
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
	getter func(ctx context.Context) ([]oaklands.StockMarketMaterial, error),
) {
	ctx := r.Context()

	sort := httpx.QueryGet(r, "sort_by", "current_multiplier")
	order := httpx.QueryGet(r, "order_by", "desc")

	materials, err := getter(ctx)
	if err != nil {
		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "HtppInternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	materials = o.sortStockMarketMaterials(materials, sort, order)
	response := newStockMarketResponse(materials)

	if err := httpx.WriteJSON(w, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/v1/oaklands/economy/stock-market/trees [GET]
//	@Description	Fetch the current tree stock market.
//
//	@Tags			Oaklands
//
//	@Param			sort_by		query		string	false	"The field to sort by."			default(current_multiplier)	enums(name, current_multiplier, last_multiplier, values.current_value, values.base_value)
//	@Param			order_by	query		string	false	"The direction to order by."	default(desc)				enums(desc, asc)
//
//	@Success		200			{object}	object{data=[]StockMarket}
//	@Failure		429			{object}	models.ErrorResponse
//	@Failure		500			{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetStockMarketTrees(w http.ResponseWriter, r *http.Request) {
	o.writeStockMarket(w, r, o.uc.GetStockMarketTrees)
}

//	@Router			/v1/oaklands/economy/stock-market/rocks [GET]
//	@Description	Fetch the current rock stock market.
//
//	@Tags			Oaklands
//
//	@Param			sort_by		query		string	false	"The field to sort by."			default(current_multiplier)	enums(name, current_multiplier, last_multiplier, values.current_value, values.base_value)
//	@Param			order_by	query		string	false	"The direction to order by."	default(desc)				enums(desc, asc)
//
//	@Success		200			{object}	object{data=[]StockMarket}
//	@Failure		429			{object}	models.ErrorResponse
//	@Failure		500			{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetStockMarketRocks(w http.ResponseWriter, r *http.Request) {
	o.writeStockMarket(w, r, o.uc.GetStockMarketRocks)
}

//	@Router			/v1/oaklands/economy/stock-market/ores [GET]
//	@Description	Fetch the current ore stock market.
//
//	@Tags			Oaklands
//
//	@Param			sort_by		query		string	false	"The field to sort by."			default(current_multiplier)	enums(name, current_multiplier, last_multiplier, values.current_value, values.base_value)
//	@Param			order_by	query		string	false	"The direction to order by."	default(desc)				enums(desc, asc)
//
//	@Success		200			{object}	object{data=[]StockMarket}
//	@Failure		429			{object}	models.ErrorResponse
//	@Failure		500			{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetStockMarketOres(w http.ResponseWriter, r *http.Request) {
	o.writeStockMarket(w, r, o.uc.GetStockMarketOres)
}
