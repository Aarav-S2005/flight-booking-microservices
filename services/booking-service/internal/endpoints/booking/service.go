package booking

import "github.com/go-resty/resty/v2"

type Service struct {
	repo   *Repository
	client *resty.Client
	// RabbitMQ also needed
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}
