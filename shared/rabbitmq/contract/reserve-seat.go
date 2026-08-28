package contract

import "github.com/google/uuid"

const (
	CreateReservationEventsExchange   = "reservation.events"
	CreateReservationEventsRoutingKey = "reservation.seat.created"
)

const (
	CreatePaymentEventsExchange   = "payment.events"
	CreatePaymentEventsRoutingKey = "payment.seat.created"
)

type CreateReservationEvent struct {
	BookingID      uuid.UUID `json:"booking_id"`
	PassengerCount int       `json:"passenger_count"`
}
