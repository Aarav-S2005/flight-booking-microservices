package rabbitmq

import "fmt"

type ExchangeConfig struct {
	Name       string
	Kind       string // "topic", "direct", "fanout"
	Durable    bool
	AutoDelete bool
}

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
}

type BindingConfig struct {
	Queue      string
	Exchange   string
	RoutingKey string
}

type Topology struct {
	Exchanges []ExchangeConfig
	Queues    []QueueConfig
	Bindings  []BindingConfig
}

func (c *Connection) DeclareTopology(t Topology) error {
	ch, err := c.NewChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	for _, e := range t.Exchanges {
		if err := ch.ExchangeDeclare(e.Name, e.Kind, e.Durable, e.AutoDelete, false, false, nil); err != nil {
			return fmt.Errorf("rabbitmq: declare exchange %q failed: %w", e.Name, err)
		}
	}
	for _, q := range t.Queues {
		if _, err := ch.QueueDeclare(q.Name, q.Durable, q.AutoDelete, q.Exclusive, false, nil); err != nil {
			return fmt.Errorf("rabbitmq: declare queue %q failed: %w", q.Name, err)
		}
	}
	for _, b := range t.Bindings {
		if err := ch.QueueBind(b.Queue, b.RoutingKey, b.Exchange, false, nil); err != nil {
			return fmt.Errorf("rabbitmq: bind queue %q failed: %w", b.Queue, err)
		}
	}
	return nil
}
