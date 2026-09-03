package database

import (
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	Id                     uuid.UUID `db:"id"`
	FlightNumber           string    `db:"flight_number"`
	AirlineName            string    `db:"airline_name"`
	AircraftType           string    `db:"aircraft_type"`
	SeatsLeft              int       `db:"seats_left"`
	SourceAirportCode      string    `db:"source_airport_code"`
	DestinationAirportCode string    `db:"destination_airport_code"`
	DepartureTime          time.Time `db:"departure_time"`
	ArrivalTime            time.Time `db:"arrival_time"`
	DurationInMins         int       `db:"duration"`
	Price                  int       `db:"price"`
	CreatedAt              time.Time `db:"created_at"`
}
