package users

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

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{service: NewService(db)}
}

func (h *Handler) InitRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/emails/{id}", h.getEmailByUserID)
	return r
}

func (h *Handler) getEmailByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	email, err := h.service.getEmailByUserID(r.Context(), userID)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	res := GetEmailResponse{Email: email}
	utility.ConvertStructToJSON(w, 200, &res)
}
