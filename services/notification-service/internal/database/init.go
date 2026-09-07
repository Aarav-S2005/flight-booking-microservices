package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	notificationSchema = `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS notifications (
			notification_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			recipient_email  TEXT NOT NULL,
			subject          TEXT NOT NULL,
			body             TEXT NOT NULL,
			sent_at          TIMESTAMPTZ NOT NULL,
			success 		 BOOL NOT NULL
		);
	`
)

func InitSchema(ctx context.Context, db *pgxpool.Pool) error {
	if _, err := db.Exec(ctx, notificationSchema); err != nil {
		return err
	}
	return nil
}
