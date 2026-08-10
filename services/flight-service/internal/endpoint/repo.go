package endpoint

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (repo *Repository) getAirportByCode(ctx context.Context, airportCode string) (string, error) {
	var name string
	err := repo.db.QueryRow(ctx, "select name from airports where airport_code = $1", airportCode).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}
