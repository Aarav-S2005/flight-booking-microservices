package async

import (
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
)

const (
	ReservationQueue = "booking-service.reservation-events"
	PaymentQueue     = "booking-service.payment-events"
)

func Topology() rabbitmq.Topology {
	return rabbitmq.Topology{
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: contract.BookingEventsExchange, Kind: "topic", Durable: true},
			{Name: contract.PaymentEventsExchange, Kind: "topic", Durable: true},
			{Name: contract.ReservationEventsExchange, Kind: "topic", Durable: true},
		},
		Queues: []rabbitmq.QueueConfig{
			{Name: ReservationQueue, Durable: true},
			{Name: PaymentQueue, Durable: true},
		},
		Bindings: []rabbitmq.BindingConfig{
			{Queue: ReservationQueue, Exchange: contract.ReservationEventsExchange, RoutingKey: contract.RoutingReservationConfirmedBooking},
			{Queue: PaymentQueue, Exchange: contract.PaymentEventsExchange, RoutingKey: contract.RoutingPaymentCompletedForBooking},
		},
	}
}
