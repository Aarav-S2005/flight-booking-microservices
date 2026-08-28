package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Aarav-S2005/flight-booking-microservices/services/booking-service/config"
	"github.com/Aarav-S2005/flight-booking-microservices/services/booking-service/internal/async"
	"github.com/Aarav-S2005/flight-booking-microservices/services/booking-service/internal/endpoints/booking"
	"github.com/Aarav-S2005/flight-booking-microservices/services/booking-service/internal/schema"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/db"
	dbInitializer "github.com/Aarav-S2005/flight-booking-microservices/shared/db"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/keys"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
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
	log.Println("Env Load Complete...")

	rdb := db.NewRedis(cfg.RedisURI)
	defer rdb.Close()
	log.Println("Connected to Redis...")

	db, err := dbInitializer.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()
	log.Println("Connected to postgres...")

	err = schema.InitSchema(ctx, db)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Schema Init Complete...")

	conn, err := rabbitmq.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("RabbitMQ initialized...")
	defer conn.Close()
	if err := conn.DeclareTopology(async.Topology()); err != nil {
		log.Fatal(err)
		return
	}
	publisher := rabbitmq.NewPublisher(conn, contract.FlightUpdateEventsExchange)

	pubKey, err := keys.GetPublicKey(cfg.PublicJwtSecretPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("PubKey loaded...")

	tokenAuth := jwtauth.New("ES256", nil, pubKey)
	log.Println("Token auth initialized...")

	h := booking.NewHandler(db, cfg.FlightServiceURL, cfg.ReservationServiceURL, rdb, publisher)
	r := h.InitRoutes(tokenAuth)
	log.Println("Endpoints Initialized...")

	err = http.ListenAndServe(":"+cfg.Port, r)
	if err != nil {
		log.Fatal(err)
		return
	}
}
