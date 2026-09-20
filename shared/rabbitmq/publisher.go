package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	mu       sync.Mutex
	ch       *amqp.Channel
	exchange string
}

func NewPublisher(conn *Connection, exchange string) (*Publisher, error) {
	ch, err := conn.NewChannel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: publisher channel failed: %w", err)
	}
	return &Publisher{ch: ch, exchange: exchange}, nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("rabbitmq: marshal failed: %w", err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ch.PublishWithContext(ctx, p.exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         data,
		DeliveryMode: amqp.Persistent,
	})
}
