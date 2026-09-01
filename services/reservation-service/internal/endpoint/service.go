package endpoint

import (
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/go-resty/resty/v2"
)

type Service struct {
	repo              *Repository
	publisher         *rabbitmq.Publisher
	flightServiceURL  string
	bookingServiceURL string
	client            *resty.Client
}

func NewService(repo *Repository, publisher *rabbitmq.Publisher, flightServiceURL string, bookingServiceURL string) *Service {
	return &Service{
		repo:              repo,
		publisher:         publisher,
		flightServiceURL:  flightServiceURL,
		bookingServiceURL: bookingServiceURL,
		client:            resty.New().SetHeader("Content-Type", "application/json"),
	}
}
