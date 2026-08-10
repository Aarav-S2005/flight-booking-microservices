package endpoint

import (
	"time"

	"github.com/google/uuid"
)

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
	FlightID           uuid.UUID `json:"flight_id"`
	FlightNumber       string    `json:"flight_number"`
	AirlineName        string    `json:"airline_name"`
	SourceAirport      string    `json:"source_airport"`
	DestinationAirport string    `json:"destination_airport"`
	DepartureTime      time.Time `json:"departure_time"`
	ArrivalTime        time.Time `json:"arrival_time"`
	Duration           int       `json:"duration"`
}

type GetFlightResponse struct {
	FlightID               uuid.UUID `json:"flight_id"`
	FlightNumber           string    `db:"flight_number"`
	AirlineName            string    `db:"airline_name"`
	AircraftType           string    `db:"aircraft_type"`
	SourceAirportCode      string    `db:"source_airport_code"`
	SourceAirportName      string    `db:"source_airport_name"`
	DestinationAirportCode string    `db:"destination_airport_code"`
	DestinationAirportName string    `db:"destination_airport_name"`
	DepartureTime          time.Time `db:"departure_time"`
	ArrivalTime            time.Time `db:"arrival_time"`
	DurationInMins         int       `db:"duration"`
	Price                  int       `db:"price"`
}
