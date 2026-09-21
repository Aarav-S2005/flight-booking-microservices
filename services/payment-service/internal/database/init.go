package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const paymentsSchema = `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;
	CREATE TABLE IF NOT EXISTS payments (
		payment_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id              UUID NOT NULL,
		booking_id           UUID NOT NULL,
		amount               BIGINT NOT NULL,
		payment_completed_at TIMESTAMPTZ,
		UNIQUE (user_id, booking_id)
	);
`

func InitSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, paymentsSchema)
	return err
}
