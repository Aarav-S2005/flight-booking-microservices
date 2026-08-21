package schema

import "time"

type Passenger struct {
	PassengerID    string    `db:"passenger_id" json:"passenger_id"`
	BookingID      string    `db:"booking_id" json:"booking_id"`
	FirstName      string    `db:"first_name" json:"first_name"`
	LastName       string    `db:"last_name" json:"last_name"`
	Age            int       `db:"age" json:"age"`
	Gender         string    `db:"gender" json:"gender"`
	PassportNumber *string   `db:"passport_number,omitempty" json:"passport_number,omitempty"`
	SeatNumber     *string   `db:"seat_number,omitempty" json:"seat_number,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
