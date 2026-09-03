package database

import (
	"time"

	"github.com/google/uuid"
)

type Reservation struct {
	ReservationID  uuid.UUID `db:"reservation_id" json:"reservation_id"`
	BookingID      uuid.UUID `db:"booking_id" json:"booking_id"`
	PassengerCount int       `db:"passenger_count" json:"passenger_count"`
}

type ReservationFlight struct {
	ReservationID       uuid.UUID `db:"reservation_id" json:"reservation_id"`
	FlightID            uuid.UUID `db:"flight_id" json:"flight_id"`
	AircraftType        string    `db:"aircraft_type" json:"aircraft_type"`
	FlightDepartureTime time.Time `db:"flight_departure_time" json:"flight_departure_time"`
	SegmentNumber       int       `db:"segment_number" json:"segment_number"`
}

type SeatAllocation struct {
	ReservationID uuid.UUID `db:"reservation_id" json:"reservation_id"`
	FlightID      uuid.UUID `db:"flight_id" json:"flight_id"`
	SeatAllocated string    `db:"seat_allocated" json:"seat_allocated"`
	PassengerID   uuid.UUID `db:"passenger_id" json:"passenger_id"`
}
