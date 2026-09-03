package async

import (
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
)

const Queue = "reservation-service.booking-events"

func Topology() rabbitmq.Topology {
	return rabbitmq.Topology{
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: contract.ReservationEventsExchange, Kind: "topic", Durable: true},
			{Name: contract.BookingEventsExchange, Kind: "topic", Durable: true},
		},
		Queues: []rabbitmq.QueueConfig{
			{Name: Queue, Durable: true},
		},
		Bindings: []rabbitmq.BindingConfig{
			{Queue: Queue, Exchange: contract.BookingEventsExchange, RoutingKey: contract.RoutingBookingConfirmedReservation},
		},
	}
}
