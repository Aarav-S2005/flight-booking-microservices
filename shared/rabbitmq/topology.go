package rabbitmq

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

func (c *Connection) DeclareExchange(cfg ExchangeConfig) error {
	return c.ch.ExchangeDeclare(
		cfg.Name, cfg.Kind, cfg.Durable, cfg.AutoDelete, false, false, nil,
	)
}

func (c *Connection) DeclareQueue(cfg QueueConfig) error {
	_, err := c.ch.QueueDeclare(
		cfg.Name, cfg.Durable, cfg.AutoDelete, cfg.Exclusive, false, nil,
	)
	return err
}

func (c *Connection) BindQueue(b BindingConfig) error {
	return c.ch.QueueBind(b.Queue, b.RoutingKey, b.Exchange, false, nil)
}

type Topology struct {
	Exchanges []ExchangeConfig
	Queues    []QueueConfig
	Bindings  []BindingConfig
}

func (c *Connection) DeclareTopology(t Topology) error {
	for _, e := range t.Exchanges {
		if err := c.DeclareExchange(e); err != nil {
			return err
		}
	}
	for _, q := range t.Queues {
		if err := c.DeclareQueue(q); err != nil {
			return err
		}
	}
	for _, b := range t.Bindings {
		if err := c.BindQueue(b); err != nil {
			return err
		}
	}
	return nil
}
