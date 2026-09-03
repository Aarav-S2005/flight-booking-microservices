package contract

const PaymentEventsExchange = "payment.events"

const (
	RoutingPaymentCompletedForBooking = "payment.completed.booking"
)

type PaymentCompletedEvent struct {
	UserID    string `json:"user_id"`
	BookingID string `json:"booking_id"`
	PaymentID string `json:"payment_id"`
	TotalFare int    `json:"total_fare"`
}
