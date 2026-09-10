package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port              string `env:"CONTROL_PLANE_PORT"`
	PostgresDSN       string `env:"POSTGRES_DSN"`
	RabbitMQURL       string `env:"RABBITMQ_URL"`
	BookingServiceURL string `env:"BOOKING_SERVICE_URL"`
}

func LoadEnv() (Config, error) {
	_ = godotenv.Load()
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
