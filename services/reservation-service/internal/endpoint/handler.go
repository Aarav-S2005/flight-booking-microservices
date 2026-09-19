package endpoint

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/database"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/middlewares"
	auth_middlewares "github.com/Aarav-S2005/flight-booking-microservices/shared/middlewares/auth-middlewares"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/utility"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, publisher *rabbitmq.Publisher, flightServiceURL string, bookingServiceURL string) *Handler {
	return &Handler{
		service: NewService(database.NewRepository(db), publisher, flightServiceURL, bookingServiceURL),
	}
}

func (h *Handler) InitRoutes(tokenAuth *jwtauth.JWTAuth, logger *slog.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.Logger(logger))
	r.Use(auth_middlewares.Verifier(tokenAuth))
	r.Use(auth_middlewares.Authenticator(tokenAuth))
	r.Post("/reserve", h.reserveSeats)
	return r
}

func (h *Handler) reserveSeats(w http.ResponseWriter, r *http.Request) {
	userID, err := utility.UserIDFromContext(r.Context())
	if err != nil {
		app_error.HandleError(w, app_error.Unauthorized("failed to parse jwt claim", err))
		return
	}
	var reqBody ReserveSeatsRequestDTO
	err = utility.ConvertJSONToStruct(r, &reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse json", errors.New("could not parse json")))
		return
	}
	err = h.service.reserveSeats(r.Context(), reqBody, userID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
