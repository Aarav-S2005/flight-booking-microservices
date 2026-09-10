package database

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	PaymentID          uuid.UUID  `db:"payment_id"`
	UserID             uuid.UUID  `db:"user_id"`
	BookingID          uuid.UUID  `db:"booking_id"`
	Amount             int        `db:"amount"`
	PaymentCompletedAt *time.Time `db:"payment_completed_at"`
}
