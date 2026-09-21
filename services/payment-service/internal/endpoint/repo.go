package endpoint

import (
	"context"
	"errors"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/payment-service/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) findRecordByUserIDAndBookingID(ctx context.Context, userID, bookingID uuid.UUID) (database.Payment, error) {
	var payment database.Payment
	err := r.db.QueryRow(ctx, "select * from payments where user_id = $1 and booking_id = $2", userID, bookingID).Scan(&payment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Payment{}, ErrPaymentNotFound
		}
		return payment, err
	}
	return payment, nil
}

func (r *Repository) savePayment(ctx context.Context, paymentTime time.Time, userID, bookingID uuid.UUID, amount int) error {
	_, err := r.db.Exec(ctx, "insert into payments (booking_id, user_id, amount, payment_completed_at) values ($1, $2, $3, $4) on conflict (booking_id, user_id) do update set payment_completed_at = $4", bookingID, userID, amount, paymentTime)
	return err
}
