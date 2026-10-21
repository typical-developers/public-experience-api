package oaklands_v1

import (
	"net/http"

	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

//	@Router			/oaklands/v1/version-info [GET]
//	@Description	Fetch version information for the game.
//	@Description	It should be noted that data is fetched whenever the published version changes.
//	@Description	Internally, updates are checked for every **1 minute**.
//
//	@Tags			Oaklands
//
//	@Success		200	{object}	object{data=VersionInfo}
//	@Failure		500	{object}	rest.ErrorResponse
//	@Failure		503	{object}	rest.ErrorResponse
//
// swagger:ignore
func (o OaklandsV1Routes) GetVersionInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	info, err := o.store.Version.Get(ctx)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[VersionInfo]{
		Data: VersionInfo{
			Version:    info.Version,
			LastUpdate: info.LastUpdate,
			NextCheck:  info.NextCheck,
		},
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=60, stale-while-revalidate=60")
	if err := httpx.WriteJSON(w, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (o OaklandsV1Routes) GetSyncTimes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sync, err := o.uc.GetSyncTimes(ctx)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[Sync]{
		Data: Sync{
			Meta: SyncMeta{
				LastUpdate: sync.LastContentSync,
				NextCheck:  sync.NextSyncCheck,
			},
			StockMarket: SyncdInfoWithReset{
				LastSync: sync.StockMarket.LastSync,
				NextSync: *sync.StockMarket.NextSync,
			},
			ClassicStore: SyncdInfoWithReset{
				LastSync: sync.ClassicStore.LastSync,
				NextSync: *sync.ClassicStore.NextSync,
			},
		},
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=60, stale-while-revalidate=60")
	if err := httpx.WriteJSON(w, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
