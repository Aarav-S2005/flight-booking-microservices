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
	Price                  int       `json:"price"`
}

type CreateFlightDTO struct {
	FlightNumber           string    `json:"flight_number"`
	AirlineName            string    `json:"airline_name"`
	AircraftType           string    `json:"aircraft_type"`
	SourceAirportCode      string    `json:"source_airport_code"`
	DestinationAirportCode string    `json:"destination_airport_code"`
	DepartureTime          time.Time `json:"departure_time"`
	ArrivalTime            time.Time `json:"arrival_time"`
	Price                  int       `json:"price"`
}

type GetFlightSeatsFromBookingResponse struct {
	SeatsLeft int `json:"seats_left"`
}
