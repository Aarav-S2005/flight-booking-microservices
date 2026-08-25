package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch       *amqp.Channel
	exchange string
}

func NewPublisher(conn *Connection, exchange string) *Publisher {
	return &Publisher{ch: conn.Channel(), exchange: exchange}
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("rabbitmq: marshal failed: %w", err)
	}

	return p.ch.PublishWithContext(ctx, p.exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         data,
		DeliveryMode: amqp.Persistent,
	})
}
