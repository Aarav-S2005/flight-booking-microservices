package schema

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingPaymentPending    BookingStatus = "PAYMENT_PENDING"
	BookingConfirmed         BookingStatus = "CONFIRMED"
	BookingFailed            BookingStatus = "FAILED"
	BookingCancelled         BookingStatus = "CANCELLED"
	BookingPartiallyCanceled BookingStatus = "PARTIALLY_CANCELLED"
)

type Contact struct {
}

type Booking struct {
	ID            uuid.UUID     `db:"_id,omitempty" json:"id"`
	BookingUserID string        `db:"user_id" json:"user_id"`
	FlightID      int64         `db:"flight_id" json:"flight_id"`
	ReservationID string        `db:"reservation_id" json:"reservation_id"`
	PaymentID     string        `db:"payment_id,omitempty" json:"payment_id,omitempty"`
	Email         *string       `db:"email" json:"email,omitempty"`
	Phone         *string       `db:"phone" json:"phone,omitempty"`
	TotalFare     float64       `db:"total_fare" json:"total_fare"`
	Status        BookingStatus `db:"status" json:"status"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at" json:"updated_at"`
}
