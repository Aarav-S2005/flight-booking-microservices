package endpoint

type ReserveSeatsRequestDTO struct {
	BookingID string `json:"booking_id"`
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
