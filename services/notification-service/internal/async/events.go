package async

import (
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
)

const (
	BookingQueue     = "notification-service.booking-events"
	ReservationQueue = "notification-service.reservation-events"
	PaymentQueue     = "notification-service.payment-events"
)

func Topology() rabbitmq.Topology {
	return rabbitmq.Topology{
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: contract.BookingEventsExchange, Kind: "topic", Durable: true},
			{Name: contract.ReservationEventsExchange, Kind: "topic", Durable: true},
			{Name: contract.PaymentEventsExchange, Kind: "topic", Durable: true},
		},
		Queues: []rabbitmq.QueueConfig{
			{Name: BookingQueue, Durable: true},
			{Name: ReservationQueue, Durable: true},
			{Name: PaymentQueue, Durable: true},
		},
		Bindings: []rabbitmq.BindingConfig{
			{Queue: BookingQueue, Exchange: contract.BookingEventsExchange, RoutingKey: contract.RoutingBookingConfirmedNotification},
			{Queue: ReservationQueue, Exchange: contract.ReservationEventsExchange, RoutingKey: contract.RoutingReservationConfirmedNotification},
			{Queue: PaymentQueue, Exchange: contract.PaymentEventsExchange, RoutingKey: contract.RoutingPaymentCompletedForNotification},
		},
	}
}
