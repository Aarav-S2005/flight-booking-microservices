package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/config"
	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/async"
	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/database"
	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/endpoint"
	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/store"
	dbInitializer "github.com/Aarav-S2005/flight-booking-microservices/shared/db"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/keys"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

func main() {
	//ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
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

	err = database.InitSchema(ctx, db)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("DB initialized...")

	registry := store.NewRegistry(ctx, db)
	log.Println("Registry initialized...")

	conn, err := rabbitmq.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("RabbitMQ initialized...")
	defer conn.Close()

	if err := conn.DeclareTopology(async.Topology()); err != nil {
		log.Fatal(err)
	}

	consumer := rabbitmq.NewConsumer(conn, async.Queue)

	err = consumer.Consume(ctx, "flight-service-consumer", 20,
		async.WrapSeatUpdatedHandler(func(e contract.FlightSeatUpdatedEvent) error {
			flightUUID, err := uuid.Parse(e.FlightID)
			if err != nil {
				return err
			}
			return registry.ApplySeatUpdate(ctx, flightUUID, e.NewSeat, e.Version, db)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("RabbitMQ consumer ready to consume...")

	pubKey, err := keys.GetPublicKey(cfg.PublicJwtSecretPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("PubKey loaded...")

	tokenAuth := jwtauth.New("ES256", nil, pubKey)
	log.Println("Token auth initialized...")

	h := endpoint.NewHandler(db, registry, cfg.ReservationServiceURL)
	r := h.InitRoutes(tokenAuth)
	log.Println("Endpoint initialized...")

	err = http.ListenAndServe(":"+cfg.Port, r)
	if err != nil {
		log.Fatal(err)
		return
	}
}
