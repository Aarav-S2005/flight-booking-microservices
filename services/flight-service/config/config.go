package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                string `env:"CONTROL_PLANE_PORT" envDefault:"3001"`
	PostgresDSN         string `env:"POSTGRES_DSN"`
	PublicJwtSecretPath string `env:"PUBLIC_KEY_JWT_PATH"`
}

func LoadEnv() (Config, error) {
	_ = godotenv.Load()
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
