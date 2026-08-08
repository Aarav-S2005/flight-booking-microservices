package endpoint

import (
	"net/http"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/store"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/utility"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, snapshot *store.Registry) *Handler {
	return &Handler{service: NewService(snapshot)}
}

func (h *Handler) InitRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/search", h.searchFlights)
	return r
}

func (h *Handler) searchFlights(w http.ResponseWriter, r *http.Request) {
	query, err := ParseQuery(r)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	flights := h.service.searchFlights(r.Context(), query)
	utility.ConvertStructToJSON(w, 200, flights)
}
