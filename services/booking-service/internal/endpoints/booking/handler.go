package booking

import (
	"net/http"

	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	auth_middlewares "github.com/Aarav-S2005/flight-booking-microservices/shared/middlewares/auth-middlewares"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/utility"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, flightServiceURL, reservationServiceURL string, rdb *redis.Client, publisher *rabbitmq.Publisher) *Handler {
	return &Handler{service: NewService(NewRepository(db), flightServiceURL, reservationServiceURL, rdb, publisher)}
}

func (h *Handler) InitRoutes(tokenAuth *jwtauth.JWTAuth) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(auth_middlewares.Verifier(tokenAuth))
		r.Use(auth_middlewares.Authenticator(tokenAuth))
	})
	r.Post("/book", h.bookTicket)
	r.Get("/booking", h.getAllBookings)
	r.Post("/validate-booking", h.validateBookingForPayment)
	r.Post("/validate-payment", h.validatePayment)
	//r.Get("/booking/{bookingId}", h.getBooking)
	return r
}

func (h *Handler) bookTicket(w http.ResponseWriter, r *http.Request) {
	var reqBody BookTicketDTO
	err := utility.ConvertJSONToStruct(r, &reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse json", err))
		return
	}
	if err := ValidateBookTicketRequestDTO(reqBody); err != nil {
		app_error.HandleError(w, app_error.BadRequest("invalid request", err))
		return
	}
	bookingUserID, err := utility.UserIDFromContext(r.Context())
	if err != nil {
		app_error.HandleError(w, app_error.Unauthorized("failed to parse jwt claim", err))
		return
	}
	resBody, err := h.service.bookTicket(r.Context(), reqBody, bookingUserID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	utility.ConvertStructToJSON(w, 200, resBody)
}

func (h *Handler) getAllBookings(w http.ResponseWriter, r *http.Request) {
	bookingUserID, err := utility.UserIDFromContext(r.Context())
	if err != nil {
		app_error.HandleError(w, app_error.Unauthorized("failed to parse jwt claim", err))
		return
	}
	resBody, err := h.service.getAllBookings(r.Context(), bookingUserID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	utility.ConvertStructToJSON(w, 200, resBody)
}

func (h *Handler) validateBookingForPayment(w http.ResponseWriter, r *http.Request) {
	var reqBody ValidateBookingRequestDTO
	err := utility.ConvertJSONToStruct(r, &reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse json", err))
		return
	}
	userID, err := uuid.Parse(reqBody.UserID)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("invalid user id", err))
		return
	}
	bookingUserID, err := uuid.Parse(reqBody.BookingID)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("invalid booking id", err))
	}
	totalFare, err := h.service.validateBooking(r.Context(), userID, bookingUserID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	utility.ConvertStructToJSON(w, 200, ValidateBookingResponseDTO{totalFare})
}

func (h *Handler) validatePayment(w http.ResponseWriter, r *http.Request) {
	var reqBody ValidatePaymentRequestDTO
	err := utility.ConvertJSONToStruct(r, &reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse json", err))
		return
	}
	valid, err := h.service.validatePayment(r.Context(), reqBody)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	if !valid {
		utility.ConvertStructToJSON(w, 200, ValidatePaymentResponseDTO{false})
		return
	}
	utility.ConvertStructToJSON(w, 200, ValidatePaymentResponseDTO{true})
}
