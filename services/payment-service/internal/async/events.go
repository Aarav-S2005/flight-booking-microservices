package async

import (
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
)

const Queue = "payment-service.booking-events"

func Topology() rabbitmq.Topology {
	return rabbitmq.Topology{
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: contract.PaymentEventsExchange, Kind: "topic", Durable: true}, // owned here
			{Name: contract.BookingEventsExchange, Kind: "topic", Durable: true}, // defensive
		},
		Queues: []rabbitmq.QueueConfig{
			{Name: Queue, Durable: true},
		},
		Bindings: []rabbitmq.BindingConfig{
			{Queue: Queue, Exchange: contract.BookingEventsExchange, RoutingKey: contract.RoutingBookingConfirmedPayment},
		},
	}
}
