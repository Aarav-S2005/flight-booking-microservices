package database

import "github.com/google/uuid"

type Flight struct {
	FlightID     uuid.UUID `db:"flight_id"`
	AircraftType string    `db:"aircraft_type"`
	SeatsLeft    int       `db:"seats_left"`
	TotalSeats   int       `db:"total_seats"`
	Version      uint64    `db:"version"`
}
