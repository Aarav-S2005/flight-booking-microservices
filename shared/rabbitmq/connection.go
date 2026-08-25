package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func Connect(url string) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: connect failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq: open channel failed: %w", err)
	}

	return &Connection{conn: conn, ch: ch}, nil
}

func (c *Connection) Channel() *amqp.Channel {
	return c.ch
}

func (c *Connection) Close() error {
	if err := c.ch.Close(); err != nil {
		_ = c.conn.Close()
		return fmt.Errorf("rabbitmq: close channel failed: %w", err)
	}
	return c.conn.Close()
}
