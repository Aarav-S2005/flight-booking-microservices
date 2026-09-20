package main

import (
	"context"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/config"
	"github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/async"
	"github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/database"
	"github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/email"
	http_client "github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/http-client"
	"github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/notification"
	db2 "github.com/Aarav-S2005/flight-booking-microservices/shared/db"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadEnv()
	if err != nil {
		panic(err)
	}

	db, err := db2.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = database.InitSchema(ctx, db)
	if err != nil {
		panic(err)
	}

	emailSender := email.NewSender(cfg.GmailUsername, cfg.GmailPassword)
	client := http_client.NewClient(cfg.AuthServiceURL)
	repo := database.NewRepository(db)
	templates := email.NewTemplates()
	svc := notification.NewService(client, emailSender, templates, repo)

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
	err = consumer.Consume(ctx, "notification-service", 10, async.NewNotificationHandler(svc))
	if err != nil {
		panic(err)
	}

}
