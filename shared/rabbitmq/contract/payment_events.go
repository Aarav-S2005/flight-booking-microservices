package contract

const PaymentEventsExchange = "payment.events"

const (
	RoutingPaymentCompletedForBooking      = "payment.completed.booking"
	RoutingPaymentCompletedForNotification = "payment.completed.notification"
)

type PaymentCompletedEvent struct {
	BookingID string `json:"booking_id"`
	PaymentID string `json:"payment_id"`
	TotalFare int    `json:"total_fare"`
}
