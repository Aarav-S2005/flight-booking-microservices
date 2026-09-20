package contract

const BookingEventsExchange = "booking.events"

const (
	RoutingFlightSeatUpdated           = "flight.seat.updated"
	RoutingBookingConfirmedPayment     = "booking.confirmed.payment"
	RoutingBookingConfirmedReservation = "booking.confirmed.reservation"

	RoutingNotifyBookingConfirmed     = "booking.notify.booking_confirmed"
	RoutingNotifyReservationConfirmed = "booking.notify.reservation_confirmed"
	RoutingNotifyPaymentCompleted     = "booking.notify.payment_completed"
	RoutingNotifyPaymentFailed        = "booking.notify.payment_failed"
)

type FlightSeatUpdatedEvent struct {
	FlightID string `json:"flight_id"`
	NewSeat  int    `json:"new_seat"`
	Version  uint64 `json:"version"`
}

type BookingConfirmedForReservationEvent struct {
	BookingID      string   `json:"booking_id"`
	UserID         string   `json:"user_id"`
	PassengerIDs   []string `json:"passenger_ids"`
	FlightSegments []string `json:"flight_segments"`
}

type BookingConfirmedForPaymentEvent struct {
	UserID    string `json:"user_id"`
	BookingID string `json:"booking_id"`
	TotalFare int    `json:"total_fare"`
}

type NotifyBookingConfirmed struct {
	UserID    string `json:"user_id"`
	BookingID string `json:"booking_id"`
}

type NotifyReservationConfirmed struct {
	UserID        string `json:"user_id"`
	BookingID     string `json:"booking_id"`
	ReservationID string `json:"reservation_id"`
}

type NotifyPaymentCompleted struct {
	UserID    string `json:"user_id"`
	BookingID string `json:"booking_id"`
	PaymentID string `json:"payment_id"`
	TotalFare int    `json:"total_fare"`
}

type NotifyPaymentFailed struct {
	UserID    string `json:"user_id"`
	BookingID string `json:"booking_id"`
	PaymentID string `json:"payment_id"`
}
