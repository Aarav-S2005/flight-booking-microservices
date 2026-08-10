package async

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	FlightEventsExchange = "flight.events"
	FlightEventsQueue    = "flight-service.events"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open rabbitmq channel failed: %w", err)
	}

	return &RabbitMQ{
		conn: conn,
		ch:   ch,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if err := r.ch.Close(); err != nil {
		_ = r.conn.Close()
		return err
	}

	return r.conn.Close()
}

func (r *RabbitMQ) DeclareFlightEventsExchange() error {
	return r.ch.ExchangeDeclare(
		FlightEventsExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) DeclareFlightEventsQueue() error {
	_, err := r.ch.QueueDeclare(
		FlightEventsQueue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	return r.ch.QueueBind(
		FlightEventsQueue,
		"flight.seat.updated",
		FlightEventsExchange,
		false,
		nil,
	)
}
