package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/config"
	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/async"
	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/endpoint"
	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/schema"
	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/store"
	dbInitializer "github.com/Aarav-S2005/flight-booking-microservices/shared/db"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/keys"
	"github.com/go-chi/jwtauth/v5"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("ENV loaded...")

	db, err := dbInitializer.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	err = schema.InitSchema(ctx, db)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("DB initialized...")

	registry := store.NewRegistry(ctx, db)
	log.Println("Registry initialized...")

	rabbitmq, err := async.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("RabbitMQ initialized...")
	defer rabbitmq.Close()
	if err := rabbitmq.DeclareFlightEventsExchange(); err != nil {
		log.Fatal(err)
	}

	if err := rabbitmq.DeclareFlightEventsQueue(); err != nil {
		log.Fatal(err)
	}
	err = rabbitmq.ConsumeFlightEvents(ctx, func(event async.SeatUpdatedEvent) error {
		return registry.ApplySeatUpdate(ctx, event.FlightID, event.NewSeat, event.Version, db)
	})
	log.Println("RabbitMQ consumer ready to consume...")
	if err != nil {
		log.Fatal(err)
	}

	pubKey, err := keys.GetPublicKey(cfg.PublicJwtSecretPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("PubKey loaded...")

	tokenAuth := jwtauth.New("ES256", nil, pubKey)
	log.Println("Token auth initialized...")

	h := endpoint.NewHandler(db, registry, cfg.BookingServiceURL)
	r := h.InitRoutes(tokenAuth)
	log.Println("Endpoint initialized...")

	err = http.ListenAndServe(":"+cfg.Port, r)
	if err != nil {
		log.Fatal(err)
		return
	}
}
