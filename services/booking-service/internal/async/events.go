package async

import (
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/google/uuid"
)

const (
	FlightUpdateExchange   = "flight.events"
	FlightUpdateRoutingKey = "flight.seat.updated"
)

type SeatUpdatedEvent struct {
	FlightID uuid.UUID `json:"flightId"`
	NewSeat  int       `json:"newSeat"`
	Version  uint64    `json:"version"`
}

func Topology() rabbitmq.Topology {
	return rabbitmq.Topology{
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: FlightUpdateExchange, Kind: "topic", Durable: true},
		},
		Bindings: []rabbitmq.BindingConfig{
			{Exchange: FlightUpdateExchange, RoutingKey: FlightUpdateRoutingKey},
		},
	}
}
