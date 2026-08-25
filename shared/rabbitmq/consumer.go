package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler func(ctx context.Context, msg amqp.Delivery) error

type Consumer struct {
	ch    *amqp.Channel
	queue string
}

func NewConsumer(conn *Connection, queue string) *Consumer {
	return &Consumer{ch: conn.Channel(), queue: queue}
}

func (c *Consumer) Consume(ctx context.Context, consumerTag string, prefetch int, handler Handler) error {
	if err := c.ch.Qos(prefetch, 0, false); err != nil {
		return fmt.Errorf("rabbitmq: qos failed: %w", err)
	}

	msgs, err := c.ch.Consume(c.queue, consumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("rabbitmq: consume failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("rabbitmq: delivery channel closed")
			}
			if err := handler(ctx, msg); err != nil {
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}
