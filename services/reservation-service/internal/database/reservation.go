package database

import (
	"time"

	"github.com/google/uuid"
)

type Reservation struct {
	ReservationID  uuid.UUID `db:"reservation_id" json:"reservation_id"`
	UserID         uuid.UUID `db:"user_id" json:"user_id"`
	BookingID      uuid.UUID `db:"booking_id" json:"booking_id"`
	PassengerCount int       `db:"passenger_count" json:"passenger_count"`
}

type ReservationFlight struct {
	ReservationID       uuid.UUID  `db:"reservation_id" json:"reservation_id"`
	FlightID            uuid.UUID  `db:"flight_id" json:"flight_id"`
	AircraftType        *string    `db:"aircraft_type" json:"aircraft_type"`
	FlightDepartureTime *time.Time `db:"flight_departure_time" json:"flight_departure_time"`
	SegmentNumber       int        `db:"segment_number" json:"segment_number"`
}

type SeatAllocation struct {
	ReservationID   uuid.UUID `db:"reservation_id" json:"reservation_id"`
	FlightID        uuid.UUID `db:"flight_id" json:"flight_id"`
	ColumnAllocated *string   `db:"column_allocated" json:"column_allocated"`
	SeatNumber      *int      `db:"seat_number" json:"seat_number"`
	PassengerID     uuid.UUID `db:"passenger_id" json:"passenger_id"`
}
