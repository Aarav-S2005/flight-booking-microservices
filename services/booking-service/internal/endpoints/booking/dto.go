package booking

import "github.com/google/uuid"

type BookTicketDTO struct {
	Email            *string             `json:"email"`
	Phone            *string             `json:"phone"`
	PassengerDetails []PassengerDetails  `json:"passenger_details"`
	FlightSegments   []FlightSegmentsDTO `json:"flight_segments"`
}

type PassengerDetails struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Age            int     `json:"age"`
	Gender         string  `json:"gender"`
	PassportNumber *string `json:"passport_number,omitempty"`
	SeatNumber     *string `json:"seat_number,omitempty"`
}

type FlightSegmentsDTO struct {
	FlightID     uuid.UUID `json:"flight_id"`
	SegmentOrder int       `json:"segment_order"`
}
