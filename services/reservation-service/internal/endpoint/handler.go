package endpoint

import (
	"log/slog"

	"github.com/Aarav-S2005/flight-booking-microservices/shared/middlewares"
	auth_middlewares "github.com/Aarav-S2005/flight-booking-microservices/shared/middlewares/auth-middlewares"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, publisher *rabbitmq.Publisher, flightServiceURL string, bookingServiceURL string) *Handler {
	return &Handler{
		service: NewService(NewRepository(db), publisher, flightServiceURL, bookingServiceURL),
	}
}

func (h *Handler) InitRoutes(tokenAuth *jwtauth.JWTAuth, logger *slog.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.Logger(logger))
	r.Use(auth_middlewares.Verifier(tokenAuth))
	r.Use(auth_middlewares.Authenticator(tokenAuth))
	return r
}
