package contract

const BookingEventsExchange = "booking.events"

const (
	RoutingFlightSeatUpdated            = "flight.seat.updated"
	RoutingBookingConfirmedPayment      = "booking.confirmed.payment"
	RoutingBookingConfirmedReservation  = "booking.confirmed.reservation"
	RoutingBookingConfirmedNotification = "booking.confirmed.notification"
)

type FlightSeatUpdatedEvent struct {
	FlightID string `json:"flightId"`
	NewSeat  int    `json:"newSeat"`
	Version  uint64 `json:"version"`
}

type BookingConfirmedForReservationEvent struct {
	BookingID      string   `json:"booking_id"`
	PassengerIDs   []string `json:"passenger_ids"`
	FlightSegments []string `json:"flight_segments"`
}

type BookingConfirmedForNotificationEvent struct {
	UserID    string `json:"user_id"`
	BookingID string `json:"booking_id"`
}

type BookingConfirmedForPaymentEvent struct {
	UserID    string `json:"user_id"`
	BookingID string `json:"booking_id"`
	TotalFare int    `json:"total_fare"`
}
