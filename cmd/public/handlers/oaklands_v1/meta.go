package oaklands_v1

import (
	"net/http"

	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

// @Router			/v1/oaklands/sync [GET]
// @Description	Fetch information on when data was last synced from the game.<br>
// @Description	It should ne noted that a majority of data is refetched from Oaklands every 5 minutes.
//
// @Tags			Oaklands
//
// @Success		200	{object}	object{data=Sync}
// @Failure		429	{object}	models.ErrorResponse
// @Failure		500	{object}	models.ErrorResponse
func (o OaklandsV1Routes) GetSyncTimes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sync, err := o.uc.GetSyncTimes(ctx)
	if err != nil {
		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "HtppInternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		return
	}

	response := models.Response[Sync]{
		Data: Sync{
			LastContentSync: sync.LastContentSync,
			StockMarket: SyncdInfoWithReset{
				LastSync: sync.StockMarket.LastSync,
				NextSync: *sync.StockMarket.NextSync,
			},
		},
	}

	if err := httpx.WriteJSON(w, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
