package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Save(ctx context.Context, notification Notification) error
}

type Repo struct {
	db *pgxpool.Pool
}

func (r Repo) Save(ctx context.Context, notification Notification) error {
	_, err := r.db.Exec(ctx, "insert into notifications(recipient_email, subject, body, sent_at, success) values($1, $2, $3, $4, $5)", notification.RecipientEmail, notification.Subject, notification.Body, notification.SentAt, notification.Success)
	return err
}

func NewRepository(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}
