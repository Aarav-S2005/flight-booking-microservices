package database

import (
	"context"
	"errors"
	"time"

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

func (repo *Repository) AddUserIfAbsent(ctx context.Context, email string, passwordHash string) (uuid.UUID, error) {
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

func (repo *Repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	row := repo.db.QueryRow(ctx, "select * from users where email = $1", email)
	var user User
	err := row.Scan(&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (repo *Repository) FindUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := repo.db.QueryRow(ctx, "select * from users where id = $1", id).Scan(&user.ID, user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}
