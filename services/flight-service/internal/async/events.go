package async

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Queue = "flight-service.booking-events"
)

func Topology() rabbitmq.Topology {
	return rabbitmq.Topology{
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: contract.BookingEventsExchange, Kind: "topic", Durable: true},
		},
		Queues: []rabbitmq.QueueConfig{
			{Name: Queue, Durable: true},
		},
		Bindings: []rabbitmq.BindingConfig{
			{Queue: Queue, Exchange: contract.BookingEventsExchange, RoutingKey: contract.RoutingFlightSeatUpdated},
		},
	}
}

func WrapSeatUpdatedHandler(handler func(contract.FlightSeatUpdatedEvent) error) rabbitmq.Handler {
	return func(ctx context.Context, msg amqp.Delivery) rabbitmq.Action {
		var event contract.FlightSeatUpdatedEvent
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
