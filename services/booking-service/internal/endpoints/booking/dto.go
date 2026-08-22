package booking

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
	bookingID string
}

type ValidateFareRequestDTO struct {
	FlightIDs []string `json:"flight_ids"`
	TotalFare int      `json:"total_fare"`
}
