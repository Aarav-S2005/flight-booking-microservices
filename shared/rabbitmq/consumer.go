package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Action int

const (
	Ack Action = iota
	NackRequeue
	NackDiscard
)

type Handler func(ctx context.Context, msg amqp.Delivery) Action

type Consumer struct {
	ch    *amqp.Channel
	queue string
}

func NewConsumer(conn *Connection, queue string) *Consumer {
	return &Consumer{ch: conn.Channel(), queue: queue}
}

// Consume starts a background goroutine consuming `queue` and returns
func (c *Consumer) Consume(ctx context.Context, consumerTag string, prefetch int, handler Handler) error {
	if err := c.ch.Qos(prefetch, 0, false); err != nil {
		return fmt.Errorf("rabbitmq: qos failed: %w", err)
	}

	msgs, err := c.ch.Consume(c.queue, consumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("rabbitmq: consume failed: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				switch handler(ctx, msg) {
				case Ack:
					_ = msg.Ack(false)
				case NackRequeue:
					_ = msg.Nack(false, true)
				case NackDiscard:
					_ = msg.Nack(false, false)
				}
			}
		}
	}()

	return nil
}
