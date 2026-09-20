package rabbitmq

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	url  string
	mu   sync.Mutex
	conn *amqp.Connection
}

func Connect(url string) (*Connection, error) {
	c := &Connection{url: url}
	if err := c.dial(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Connection) dial() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("rabbitmq: connect failed: %w", err)
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	return nil
}

func (c *Connection) NewChannel() (*amqp.Channel, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil || conn.IsClosed() {
		if err := c.dial(); err != nil {
			return nil, err
		}
		c.mu.Lock()
		conn = c.conn
		c.mu.Unlock()
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: open channel failed: %w", err)
	}
	return ch, nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
