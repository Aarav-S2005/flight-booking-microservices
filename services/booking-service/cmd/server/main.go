package main

import (
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/services/booking-service/config"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/db"
)

func main() {
	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatal(err)
		return
	}
	rdb := db.NewRedis(cfg.RedisURI)
	defer rdb.Close()
}
