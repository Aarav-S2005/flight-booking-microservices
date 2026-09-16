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

type ValidatePaymentToBookingRequestDTO struct {
	PaymentTime time.Time `json:"payment_time"`
	UserID      string    `json:"user_id"`
	BookingID   string    `json:"booking_id"`
}

type ValidatePaymentToBookingResponseDTO struct {
	Valid bool `json:"valid"`
}

type ValidatePaymentRequestDTO struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
}
