package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/config"
	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/async"
	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/database"
	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/endpoint"
	dbInitializer "github.com/Aarav-S2005/flight-booking-microservices/shared/db"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/keys"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/go-chi/jwtauth/v5"
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
	if err != nil {
		log.Fatal(err)
		return
	}
	err = database.InitSchema(ctx, db)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Schema Initialized and Aircraft seeded...")

	repo := database.NewRepository(db)

	conn, err := rabbitmq.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("RabbitMQ initialized...")
	defer conn.Close()
	if err := conn.DeclareTopology(async.Topology()); err != nil {
		log.Fatal(err)
	}
	consumer := rabbitmq.NewConsumer(conn, async.Queue)
	err = consumer.Consume(ctx, "notification-service", 10, async.HandleBookingConfirmed(repo))
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	h := endpoint.NewHandler(db, nil, cfg.FlightServiceURL, cfg.BookingServiceURL)

	r := h.InitRoutes(tokenAuth, logger)
	err = http.ListenAndServe(":"+cfg.Port, r)
}
