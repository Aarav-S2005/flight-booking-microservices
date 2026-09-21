package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string `env:"CONTROL_PLANE_PORT" envDefault:"3301"`
	PostgresDSN           string `env:"POSTGRES_DSN"`
	PublicJwtSecretPath   string `env:"PUBLIC_KEY_JWT_PATH"`
	RabbitMQURL           string `env:"RABBITMQ_URL"`
	FlightServiceURL      string `env:"FLIGHT_SERVICE_URL"`
	ReservationServiceURL string `env:"RESERVATION_SERVICE_URL"`
	PaymentServiceURL     string `env:"PAYMENT_SERVICE_URL"`
	RedisURI              string `env:"REDIS_URI"`
}

func LoadEnv() (Config, error) {
	_ = godotenv.Load()
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
