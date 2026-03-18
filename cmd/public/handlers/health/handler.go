package health

import (
	"net/http"

	"github.com/go-chi/chi"
)

type HealthRoutes struct{}

func NewHealth(r *chi.Mux) {
	h := HealthRoutes{}
	r.Get("/", h.GetHealth)
}

func (h *HealthRoutes) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
