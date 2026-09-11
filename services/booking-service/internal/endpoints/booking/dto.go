package booking

import (
	"time"
)

type BookTicketDTO struct {
	Email            string              `json:"email"`
	Phone            string              `json:"phone"`
	TotalFare        int                 `json:"total_fare"`
	PassengerDetails []PassengerDetails  `json:"passenger_details"`
	FlightSegments   []FlightSegmentsDTO `json:"flight_segments"`
}

type PassengerDetails struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Age            int    `json:"age"`
	Gender         string `json:"gender"`
	PassportNumber string `json:"passport_number"`
}

type FlightSegmentsDTO struct {
	FlightID     string `json:"flight_id"`
	SegmentOrder int    `json:"segment_order"`
}

type BookTicketResponseDTO struct {
	BookingID string `json:"booking_id"`
}

type GetAllBookingsDTO struct {
	Bookings []GetBookingDTO `json:"bookings"`
}

type GetBookingDTO struct {
	BookingID        string             `json:"booking_id"`
	PassengerDetails []PassengerDetails `json:"passenger_details"`
	FlightDetails    []FlightDetailsDTO `json:"flight_details"`
	TotalFare        int                `json:"total_fare"`
}

type FlightDetailsDTO struct {
	FlightNumber           string    `json:"flight_number"`
	AirlineName            string    `json:"airline_name"`
	AircraftType           string    `json:"aircraft_type"`
	SourceAirportCode      string    `json:"source_airport_code"`
	SourceAirportName      string    `json:"source_airport_name"`
	DestinationAirportCode string    `json:"destination_airport_code"`
	DestinationAirportName string    `json:"destination_airport_name"`
	DepartureTime          time.Time `json:"departure_time"`
	ArrivalTime            time.Time `json:"arrival_time"`
	DurationInMins         int       `json:"duration"`
	SegmentOrder           int       `json:"segment_order"`
}

type ValidateFareRequestDTO struct {
	FlightIDs []string `json:"flight_ids"`
	TotalFare int      `json:"total_fare"`
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
