package async

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
)

const SeatUpdatedRoutingKey = "flight.seat.updated"

type SeatUpdatedEvent struct {
	FlightID uuid.UUID `json:"flightId"`
	NewSeat  int       `json:"newSeat"`
	Version  uint64    `json:"version"`
}

func (r *RabbitMQ) ConsumeFlightEvents(ctx context.Context, handler func(SeatUpdatedEvent) error) error {
	if err := r.ch.Qos(20, 0, false); err != nil {
		return fmt.Errorf("set qos failed: %w", err)
	}
	msgs, err := r.ch.Consume(
		FlightEventsQueue,
		"flight-service-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consumer failed: %w", err)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				var event SeatUpdatedEvent
				if err := json.Unmarshal(d.Body, &event); err != nil {
					log.Printf("bad payload, discarding: %v", err)
					_ = d.Nack(false, false)
					continue
				}
				if err := handler(event); err != nil {
					log.Printf("handler error, requeueing: %v", err)
					_ = d.Nack(false, true)
					continue
				}
				_ = d.Ack(false)
			}
		}
	}()
	return nil
}
