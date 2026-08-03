package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/auth-service/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (repo *Repository) addUserIfAbsent(ctx context.Context, email string, passwordHash string) (uuid.UUID, error) {
	var id uuid.UUID

	err := repo.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
		RETURNING id
	`, email, passwordHash, time.Now()).Scan(&id)

	switch {
	case err == nil:
		return id, nil
	case errors.Is(err, pgx.ErrNoRows):
		return uuid.Nil, ErrUserAlreadyExists
	default:
		return uuid.Nil, err
	}
}

func (repo *Repository) findUserByEmail(ctx context.Context, email string) (*db.User, error) {
	row := repo.db.QueryRow(ctx, "select * from users where email = $1", email)
	if row == nil {
		return nil, ErrUserNotFound
	}
	var user db.User
	err := row.Scan(&user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
