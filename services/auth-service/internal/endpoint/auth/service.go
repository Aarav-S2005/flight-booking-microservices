package auth

import (
	"context"
	"errors"

	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/jwt"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *Repository
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) signup(ctx context.Context, reqBody LoginRequest) (string, error) {
	hashedPassword, err := HashPassword(reqBody.Password)
	if err != nil {
		return "", err
	}
	id, err := s.repo.addUserIfAbsent(ctx, reqBody.Email, hashedPassword)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserAlreadyExists):
			return "", app_error.Conflict("user already exists", err)
		default:
			return "", err
		}
	}
	token, err := jwt.SignJwt(id)
	return token, nil
}

func (s *Service) login(ctx context.Context, reqBody LoginRequest) (string, error) {
	hashedPassword, err := HashPassword(reqBody.Password)
	if err != nil {
		return "", err
	}
	user, err := s.repo.findUserByEmail(ctx, reqBody.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", app_error.Unauthorized("invalid email or password", err)
		}
		return "", err
	}
	if user.PasswordHash != hashedPassword {
		return "", app_error.Unauthorized("invalid email or password", errors.New("invalid email or password"))
	}
	return "", nil
}
