package booking

import (
	"net/http"

	auth_middlewares "github.com/Aarav-S2005/flight-booking-microservices/shared/middlewares/auth-middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{service: NewService(NewRepository(db))}
}

func (h *Handler) InitRoutes(tokenAuth *jwtauth.JWTAuth) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(auth_middlewares.Verifier(tokenAuth))
		r.Use(auth_middlewares.Authenticator(tokenAuth))
	})
	r.Post("/book", h.bookTicket)
	r.Get("/booking", h.getAllBookings)
	r.Get("/booking/{bookingId}", h.getBooking)
	return r
}

func (h *Handler) bookTicket(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) getAllBookings(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) getBooking(w http.ResponseWriter, r *http.Request) {}
