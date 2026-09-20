package endpoint

import (
	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/endpoint/auth"
	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/endpoint/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(db *pgxpool.Pool, tokenAuth *jwtauth.JWTAuth) chi.Router {
	authHandler := auth.NewHandler(db, tokenAuth)
	usersHandler := users.NewHandler(db)

	authRouter := authHandler.InitRoutes()
	usersRouter := usersHandler.InitRoutes()

	r := chi.NewRouter()
	r.Mount("/auth", authRouter)
	r.Mount("/users", usersRouter)

	return r
}
