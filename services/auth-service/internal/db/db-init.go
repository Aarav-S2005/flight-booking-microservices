package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	initSchema = `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
		created_at TIMESTAMP NOT NULL DEFAULT now(),
		updated_at TIMESTAMP NOT NULL DEFAULT now()
	); `
)

func InitSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, initSchema)
	if err != nil {
		return err
	}
	return nil
}
