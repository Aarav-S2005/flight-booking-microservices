package endpoint

import (
	"errors"
	"net/http"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/store"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
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
	r.Get("/flight/validate-multiple", h.validateFlights)
	r.Get("/flight/validate/{flightID}", h.validateFlightID)
	r.Get("/search", h.searchFlights)
	r.Post("/flight/validate-fare", h.validateFare)
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

func (h *Handler) validateFlights(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query()["flightID"]
	flightIDs := make([]uuid.UUID, 0, len(ids))

	if flightIDs == nil {
		http.Error(w, "flight_id is required", http.StatusBadRequest)
		return
	}

	for _, stringId := range ids {
		id, err := uuid.Parse(stringId)
		if err != nil {
			http.Error(w, "invalid flight_id", http.StatusBadRequest)
			return
		}
		flightIDs = append(flightIDs, id)
	}
	resBody, err := h.service.checkAllFlightIDs(r.Context(), flightIDs)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	utility.ConvertStructToJSON(w, 200, resBody)
}

func (h *Handler) validateFlightID(w http.ResponseWriter, r *http.Request) {
	flightID := chi.URLParam(r, "flight-id")
	if flightID == "" {
		app_error.HandleError(w, app_error.BadRequest("flight-id missing in URL", errors.New("flight-id required")))
		return
	}
	flightIDUUID, err := uuid.Parse(flightID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	resBody, err := h.service.checkFlightID(r.Context(), flightIDUUID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	utility.ConvertStructToJSON(w, 200, resBody)
}

func (h *Handler) validateFare(w http.ResponseWriter, r *http.Request) {
	var reqBody ValidateFareRequestDTO
	err := utility.ConvertJSONToStruct(r, &reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse request body", err))
		return
	}
	err = h.service.validateFare(r.Context(), reqBody)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
