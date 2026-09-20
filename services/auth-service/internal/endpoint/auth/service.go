package auth

import (
	"context"
	"errors"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/database"
	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/jwt"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo      *database.Repository
	tokenAuth *jwtauth.JWTAuth
	dummyHash string
}

func NewService(db *pgxpool.Pool, tokenAuth *jwtauth.JWTAuth) *Service {
	dh, err := generateDummyHash()
	if err != nil {
		log.Fatal(err)
	}
	return &Service{repo: database.NewRepository(db), tokenAuth: tokenAuth, dummyHash: dh}
}

func (s *Service) signup(ctx context.Context, reqBody LoginRequest) (string, error) {
	hashedPassword, err := HashPassword(reqBody.Password)
	if err != nil {
		return "", err
	}
	id, err := s.repo.AddUserIfAbsent(ctx, reqBody.Email, hashedPassword)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrUserAlreadyExists):
			return "", app_error.Conflict("user already exists", err)
		default:
			return "", err
		}
	}
	token, err := jwt.SignJwt(s.tokenAuth, id)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) login(ctx context.Context, reqBody LoginRequest) (string, error) {
	user, err := s.repo.FindUserByEmail(ctx, reqBody.Email)
	passwordHash := s.dummyHash
	if err != nil {
		log.Print("wrong email")
		if !errors.Is(err, database.ErrUserNotFound) {
			return "", err
		}
	}
	passwordHash = user.PasswordHash
	if !VerifyPassword(reqBody.Password, passwordHash) {
		log.Print("wrong password")
		return "", app_error.Unauthorized("invalid email or password", errors.New("invalid email or password"))
	}
	token, err := jwt.SignJwt(s.tokenAuth, user.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
