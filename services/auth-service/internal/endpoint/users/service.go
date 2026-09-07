package users

import (
	"context"
	"errors"

	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/database"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *database.Repository
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{repo: database.NewRepository(db)}
}

func (s *Service) getEmailByUserID(ctx context.Context, userID string) (string, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", app_error.BadRequest("could not parse uuid", err)
	}
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			return "", app_error.NotFound("user not found", err)
		}
		return "", err
	}
	return user.Email, nil
}
