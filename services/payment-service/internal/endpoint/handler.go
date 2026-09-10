package endpoint

import (
	"net/http"

	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/utility"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, bookingServiceURL string) *Handler {
	return &Handler{
		service: NewService(db, bookingServiceURL),
	}
}

func (h *Handler) InitRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/pay", h.Pay)
	return r
}

func (h *Handler) Pay(w http.ResponseWriter, r *http.Request) {
	userID, err := utility.UserIDFromContext(r.Context())
	var reqBody MakePaymentDTO
	err = utility.ConvertJSONToStruct(r, reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse json", err))
		return
	}
	err = h.service.Pay(r.Context(), userID, reqBody)
	if err != nil {
		app_error.HandleError(w, err)
	}
	w.WriteHeader(200)
}
