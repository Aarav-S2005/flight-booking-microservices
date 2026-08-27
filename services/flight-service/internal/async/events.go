package async

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Exchange   = "flight.events"
	Queue      = "flight-service.events"
	RoutingKey = "flight.seat.updated"
)

type SeatUpdatedEvent struct {
	FlightID uuid.UUID `json:"flightId"`
	NewSeat  int       `json:"newSeat"`
	Version  uint64    `json:"version"`
}

func Topology() rabbitmq.Topology {
	return rabbitmq.Topology{
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: Exchange, Kind: "topic", Durable: true},
		},
		Queues: []rabbitmq.QueueConfig{
			{Name: Queue, Durable: true},
		},
		Bindings: []rabbitmq.BindingConfig{
			{Queue: Queue, Exchange: Exchange, RoutingKey: RoutingKey},
		},
	}
}

func WrapSeatUpdatedHandler(handler func(SeatUpdatedEvent) error) rabbitmq.Handler {
	return func(ctx context.Context, msg amqp.Delivery) rabbitmq.Action {
		var event SeatUpdatedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("bad payload, discarding: %v", err)
			return rabbitmq.NackDiscard
		}
		if err := handler(event); err != nil {
			log.Printf("handler error, requeueing: %v", err)
			return rabbitmq.NackRequeue
		}
		return rabbitmq.Ack
	}
}
