package async

import (
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
)

func Topology() rabbitmq.Topology {
	return rabbitmq.Topology{
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: contract.FlightUpdateEventsExchange, Kind: "topic", Durable: true},
		},
		Bindings: []rabbitmq.BindingConfig{
			{Exchange: contract.FlightUpdateEventsExchange, RoutingKey: contract.FlightSeatUpdateRoutingKey},
		},
	}
}
