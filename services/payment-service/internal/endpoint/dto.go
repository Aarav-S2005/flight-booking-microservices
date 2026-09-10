package endpoint

import "time"

type MakePaymentDTO struct {
	BookingID string `json:"booking_id"`
	Amount    int    `json:"amount"`
}

type ValidateBookingRequestDTO struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
}

type ValidateBookingResponseDTO struct {
	TotalFare int `json:"total_fare"`
}

type ValidatePaymentRequestDTO struct {
	PaymentTime time.Time `json:"payment_time"`
	UserID      string    `json:"user_id"`
	BookingID   string    `json:"booking_id"`
}

type ValidatePaymentResponseDTO struct {
	Valid bool `json:"valid"`
}
