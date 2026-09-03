package contract

const ReservationEventsExchange = "reservation.events"

const (
	RoutingReservationConfirmedBooking = "reservation.confirmed.booking"
)

type ReservationEventsForBooking struct {
	UserID        string `json:"user_id"`
	BookingID     string `json:"booking_id"`
	ReservationID string `json:"reservation_id"`
}
