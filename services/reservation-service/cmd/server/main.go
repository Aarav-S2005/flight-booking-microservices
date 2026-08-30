package main

import (
	"context"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/config"
	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/schema"
	dbInitializer "github.com/Aarav-S2005/flight-booking-microservices/shared/db"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Env Loaded...")

	db, err := dbInitializer.NewPostgres(ctx, cfg.PostgresDSN)
	err = schema.InitSchema(ctx, db)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Schema Initialized and Aircraft seeded...")
}
