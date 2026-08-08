package endpoint

import "time"

type SearchResponse struct {
	Flights []Route `json:"flights"`
}

type Route struct {
	SourceAirport      string    `json:"source_airport"`
	DestinationAirport string    `json:"destination_airport"`
	Price              int       `json:"price"`
	TotalDuration      int       `json:"total_duration"`
	Segments           []Segment `json:"segments"`
}

type Segment struct {
	FlightNumber       string    `json:"flight_number"`
	AirlineName        string    `json:"airline_name"`
	SourceAirport      string    `json:"source_airport"`
	DestinationAirport string    `json:"destination_airport"`
	DepartureTime      time.Time `json:"departure_time"`
	ArrivalTime        time.Time `json:"arrival_time"`
	Duration           int       `json:"duration"`
}
