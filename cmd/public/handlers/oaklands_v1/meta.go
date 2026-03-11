package oaklands_v1

import (
	"errors"
	"net/http"

	"github.com/redis/go-redis/v9"
	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

//	@Router			/v1/oaklands/sync [GET]
//	@Description	Fetch information on when data was last synced from the game.
//	@Description	It should be noted that a majority of data is fetched when checking if the experience was updated,
//	@Description	such as newsletters and changelogs. Some content with resync at certain intervals by itself
//	@Description	alongside update sync checks.<br><br>
//	@Description	Internally, updates are checked for every **5 minutes**.
//
//	@Tags			Oaklands
//
//	@Success		200	{object}	object{data=Sync}
//	@Failure		429	{object}	models.ErrorResponse
//	@Failure		500	{object}	models.ErrorResponse
//
// swagger:ignore
func (o OaklandsV1Routes) GetSyncTimes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sync, err := o.uc.GetSyncTimes(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "ResourceNotCached",
				Message: "The requested resource is not cached. Try again in a bit.",
			}, http.StatusServiceUnavailable)
		}

		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "InternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		return
	}

	response := models.Response[Sync]{
		Data: Sync{
			Meta: SyncMeta{
				LastUpdate: sync.LastContentSync,
				NextCheck:  sync.NextSyncCheck,
			},
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
