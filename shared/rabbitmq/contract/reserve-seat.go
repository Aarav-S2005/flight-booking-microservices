package contract

const (
	CreateReservationEventsExchange   = "reservation.events"
	CreateReservationEventsRoutingKey = "reservation.seat.created"
)

const (
	CreatePaymentEventsExchange   = "payment.events"
	CreatePaymentEventsRoutingKey = "payment.seat.created"
)

type CreateReservationEvent struct {
	BookingID      string `json:"booking_id"`
	PassengerCount int    `json:"passenger_count"`
}
