package rabbitmq

import (
	"context"
	"log"
	"sync/atomic"
	"time"

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
	conn    *Connection
	queue   string
	healthy atomic.Bool
}

func NewConsumer(conn *Connection, queue string) *Consumer {
	return &Consumer{conn: conn, queue: queue}
}

// Healthy reports whether the consumer currently has an active subscription.
func (c *Consumer) Healthy() bool {
	return c.healthy.Load()
}

func (c *Consumer) Consume(ctx context.Context, consumerTag string, prefetch int, handler Handler) error {
	go c.runLoop(ctx, consumerTag, prefetch, handler)
	return nil
}

func (c *Consumer) runLoop(ctx context.Context, consumerTag string, prefetch int, handler Handler) {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for ctx.Err() == nil {
		ch, err := c.conn.NewChannel()
		if err != nil {
			c.healthy.Store(false)
			log.Printf("rabbitmq: consumer %q: open channel failed: %v, retrying in %s", c.queue, err, backoff)
			if !sleepCtx(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff, maxBackoff)
			continue
		}

		if err := ch.Qos(prefetch, 0, false); err != nil {
			c.healthy.Store(false)
			log.Printf("rabbitmq: consumer %q: qos failed: %v", c.queue, err)
			_ = ch.Close()
			if !sleepCtx(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff, maxBackoff)
			continue
		}

		msgs, err := ch.Consume(c.queue, consumerTag, false, false, false, false, nil)
		if err != nil {
			c.healthy.Store(false)
			log.Printf("rabbitmq: consumer %q: consume failed: %v", c.queue, err)
			_ = ch.Close()
			if !sleepCtx(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff, maxBackoff)
			continue
		}

		closeNotify := ch.NotifyClose(make(chan *amqp.Error, 1))
		log.Printf("rabbitmq: consumer %q: subscribed", c.queue)
		c.healthy.Store(true)
		backoff = time.Second

		c.drain(ctx, msgs, closeNotify, handler)

		c.healthy.Store(false)
		if ctx.Err() != nil {
			return
		}
		log.Printf("rabbitmq: consumer %q: disconnected, reconnecting in %s", c.queue, backoff)
		if !sleepCtx(ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff, maxBackoff)
	}
}

// drain processes deliveries until the channel closes or ctx is cancelled.
func (c *Consumer) drain(ctx context.Context, msgs <-chan amqp.Delivery, closeNotify <-chan *amqp.Error, handler Handler) {
	for {
		select {
		case <-ctx.Done():
			return
		case amqErr, ok := <-closeNotify:
			if ok {
				log.Printf("rabbitmq: consumer %q: channel closed: %v", c.queue, amqErr)
			} else {
				log.Printf("rabbitmq: consumer %q: channel closed", c.queue)
			}
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("rabbitmq: consumer %q: delivery channel closed", c.queue)
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
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}

func nextBackoff(cur, max time.Duration) time.Duration {
	next := cur * 2
	if next > max {
		return max
	}
	return next
}
