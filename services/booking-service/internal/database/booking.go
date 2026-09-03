package database

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingPaymentPending BookingStatus = "PAYMENT_PENDING"
	BookingConfirmed      BookingStatus = "CONFIRMED"
	BookingFailed         BookingStatus = "FAILED"
	BookingCancelled      BookingStatus = "CANCELLED"
)

type Contact struct {
}

type Booking struct {
	BookingID     uuid.UUID     `db:"booking_id,omitempty" json:"id,omitempty"`
	BookingUserID uuid.UUID     `db:"booking_user_id" json:"user_id"`
	ReservationID *uuid.UUID    `db:"reservation_id,omitempty" json:"reservation_id,omitempty"`
	PaymentID     *uuid.UUID    `db:"payment_id,omitempty" json:"payment_id,omitempty"`
	Email         *string       `db:"email" json:"email,omitempty"`
	Phone         *string       `db:"phone" json:"phone,omitempty"`
	TotalFare     float64       `db:"total_fare" json:"total_fare,omitempty"`
	Status        BookingStatus `db:"status" json:"status,omitempty"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt     time.Time     `db:"updated_at" json:"updated_at,omitempty"`
}

type FlightSegment struct {
	BookingID    uuid.UUID `db:"booking_id" json:"booking_id"`
	FlightID     uuid.UUID `db:"flight_id" json:"flight_id"`
	SegmentOrder int       `db:"segment_order" json:"segment_order"`
}
