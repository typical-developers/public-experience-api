package oaklands_v1

import (
	"net/http"
	"strings"
	"time"
	_ "time/tzdata" // embed zoneinfo so America/New_York always resolves

	"github.com/go-chi/render"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
)

var easternTZ, _ = time.LoadLocation("America/New_York") // always resolves: time/tzdata is embedded

//	@Router			/oaklands/v1/economy/daily-material-buffs [GET]
//	@Description	Get all of the daily material buff values (tree, rocks, ores).
//
//	@Tags			Oaklands
//
//	@Success		200	{object}	object{refresh_at=time.Time,data=map[string][]typicalapi.OaklandsMaterialBuff}
//	@Failure		503	{object}	rest.ErrorResponse
//	@Failure		500	{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetDailyMaterialBuffs(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	response := MaterialBuffResponse{
		RefreshAt: time.Date(
			now.Year(), now.Month(), now.Day()+1,
			0, 0, 0, 0, easternTZ,
		).UTC(),
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
	render.JSON(w, r, response)
}
