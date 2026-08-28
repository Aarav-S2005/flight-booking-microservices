package contract

import "github.com/google/uuid"

const (
	FlightUpdateEventsExchange = "flight.events"
	FlightSeatUpdateRoutingKey = "flight.seat.updated"
)

type SeatUpdatedEvent struct {
	FlightID uuid.UUID `json:"flightId"`
	NewSeat  int       `json:"newSeat"`
	Version  uint64    `json:"version"`
}
