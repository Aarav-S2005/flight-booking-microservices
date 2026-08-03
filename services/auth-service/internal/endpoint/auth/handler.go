package auth

import (
	"net/http"

	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	auth_middlewares "github.com/Aarav-S2005/flight-booking-microservices/shared/middlewares/auth-middlewares"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/utility"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool, tokenAuth *jwtauth.JWTAuth) *Handler {
	return &Handler{service: NewService(db, tokenAuth)}
}

func (h *Handler) InitRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/signup", h.signup)
	r.Post("/login", h.login)

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(h.service.tokenAuth))
		r.Use(auth_middlewares.Authenticator(h.service.tokenAuth))

		r.Post("/logout", h.logout)
	})
	return r
}

func (h *Handler) signup(w http.ResponseWriter, r *http.Request) {
	var reqBody LoginRequest
	err := utility.ConvertJSONToStruct(r, &reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse json", err))
		return
	}
	token, err := h.service.signup(r.Context(), reqBody)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	SetCookie(w, "jwt", token)
	utility.ConvertStructToJSON(w, 201, LoginResponse{Email: reqBody.Email})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var reqBody LoginRequest
	err := utility.ConvertJSONToStruct(r, &reqBody)
	if err != nil {
		app_error.HandleError(w, app_error.BadRequest("could not parse json", err))
		return
	}
	token, err := h.service.login(r.Context(), reqBody)
	if err != nil {
		app_error.HandleError(w, err)
		return
	}
	SetCookie(w, "jwt", token)
	utility.ConvertStructToJSON(w, 201, LoginResponse{Email: reqBody.Email})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	RemoveCookie(w, "jwt")
	w.WriteHeader(http.StatusOK)
}
