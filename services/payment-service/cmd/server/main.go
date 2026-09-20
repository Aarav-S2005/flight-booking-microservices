package main

import (
	"context"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/services/payment-service/config"
	"github.com/Aarav-S2005/flight-booking-microservices/services/payment-service/internal/async"
	dbInitializer "github.com/Aarav-S2005/flight-booking-microservices/shared/db"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadEnv()
	if err != nil {
		panic(err)
	}

	db, err := dbInitializer.NewPostgres(ctx, cfg.PostgresDSN)

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
	err = consumer.Consume(ctx, "payment-service", 10, async.AddNewDuePaymentRecord(db))
	if err != nil {
		log.Fatal(err)
	}
}
