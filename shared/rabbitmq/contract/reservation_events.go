package contract

const ReservationEventsExchange = "reservation.events"

const (
	RoutingReservationConfirmedNotification = "reservation.confirmed.notification"
	RoutingReservationConfirmedBooking      = "reservation.confirmed.booking"
)

type ReservationEventsForNotification struct {
	UserID        string `json:"user_id"`
	ReservationID string `json:"reservation_id"`
	BookingID     string `json:"booking_id"`
}

type ReservationEventsForBooking struct {
	UserID    string `json:"user_id"`
	BookingID string `json:"booking_id"`
}
