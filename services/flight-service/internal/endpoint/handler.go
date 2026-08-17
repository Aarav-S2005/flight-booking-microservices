package endpoint

import (
	"errors"
	"net/http"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/store"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	auth_middlewares "github.com/Aarav-S2005/flight-booking-microservices/shared/middlewares/auth-middlewares"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/utility"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, registry *store.Registry, reservationURL string) *Handler {
	return &Handler{service: NewService(registry, NewRepository(db), reservationURL)}
}

func (h *Handler) InitRoutes(tokenAuth *jwtauth.JWTAuth) chi.Router {
	r := chi.NewRouter()
	r.Get("/flight/{flightID}", h.getFlight)
	r.Post("/admin/flight", h.createFlight)
	r.Group(func(r chi.Router) {
		r.Use(auth_middlewares.Verifier(tokenAuth))
		r.Use(auth_middlewares.Authenticator(tokenAuth))
		r.Get("/search", h.searchFlights)
	})
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

func (h *Handler) getFlight(w http.ResponseWriter, r *http.Request) {
	flightID := chi.URLParam(r, "flightID")
	if flightID == "" {
		app_error.HandleError(w, app_error.BadRequest("flightID missing in URL", errors.New("flightID missing in URL")))
		return
	}
	flightIDUUID, err := uuid.Parse(flightID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	res, err := h.service.getFlight(r.Context(), flightIDUUID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	utility.ConvertStructToJSON(w, 200, res)
}

func (h *Handler) createFlight(w http.ResponseWriter, r *http.Request) {
	var reqBody CreateFlightDTO
	err := utility.ConvertJSONToStruct(r, &reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse request body", err))
		return
	}
	err = h.service.createFlight(r.Context(), reqBody)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
