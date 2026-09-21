package endpoint

type ReserveSeatsRequestDTO struct {
	BookingID string        `json:"booking_id"`
	Seats     SeatSelection `json:"seats"`
}

type SeatSelection struct {
	Flights []SelectPerFlight `json:"flights"`
}

type SelectPerFlight struct {
	FlightID      string `json:"flight_id"`
	SeatsSelected []Seat `json:"seats_selected"`
}

type Seat struct {
	Column     string `json:"column"`
	SeatNumber string `json:"number"`
}

type ValidateBookingForReservationRequestDTO struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
}

type ValidateBookingForReservationResponseDTO struct {
	Passengers []string `json:"passengers"`
	FlightIDs  []string `json:"flight_ids"`
	Status     string   `json:"status"`
}

type GetFlightSeatsFromReservationRequest struct {
	SeatsLeft int `json:"seats_left"`
}
