package contract

const (
	FlightUpdateEventsExchange = "flight.events"
	FlightSeatUpdateRoutingKey = "flight.seat.updated"
)

type SeatUpdatedEvent struct {
	FlightID string `json:"flightId"`
	NewSeat  int    `json:"newSeat"`
	Version  uint64 `json:"version"`
}
