package endpoint

import (
	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/endpoint/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(db *pgxpool.Pool, tokenAuth *jwtauth.JWTAuth) chi.Router {
	authHandler := auth.NewHandler(db, tokenAuth)

	authRouter := authHandler.InitRoutes()

	//r := chi.NewRouter()
	//r.Mount("/auth", authRouter)

	return authRouter
}
