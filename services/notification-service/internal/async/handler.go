package async

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/services/notification-service/internal/notification"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewNotificationHandler(svc notification.Service) rabbitmq.Handler {
	return func(ctx context.Context, msg amqp.Delivery) rabbitmq.Action {
		parse, ok := registry[msg.RoutingKey]
		if !ok {
			log.Printf("no parser for routing key %q, discarding", msg.RoutingKey)
			return rabbitmq.NackDiscard
		}
		job, err := parse(msg.Body)
		if err != nil {
			log.Printf("unmarshal failed for %q: %v", msg.RoutingKey, err)
			return rabbitmq.NackDiscard
		}
		if err := svc.Notify(ctx, job); err != nil {
			log.Printf("notify failed for %q: %v", msg.RoutingKey, err)
			return rabbitmq.NackRequeue
		}

		return rabbitmq.Ack
	}
}

type parseFunc func(body []byte) (notification.Job, error)

var registry = map[string]parseFunc{
	contract.RoutingNotifyBookingConfirmed: func(b []byte) (notification.Job, error) {
		var p contract.NotifyBookingConfirmed
		if err := json.Unmarshal(b, &p); err != nil {
			return notification.Job{}, err
		}
		return notification.Job{UserID: p.UserID, TemplateName: "booking_confirmed", Data: p}, nil
	},
	contract.RoutingNotifyReservationConfirmed: func(b []byte) (notification.Job, error) {
		var p contract.NotifyReservationConfirmed
		if err := json.Unmarshal(b, &p); err != nil {
			return notification.Job{}, err
		}
		return notification.Job{UserID: p.UserID, TemplateName: "reservation_confirmed", Data: p}, nil
	},
	contract.RoutingNotifyPaymentCompleted: func(b []byte) (notification.Job, error) {
		var p contract.NotifyPaymentCompleted
		if err := json.Unmarshal(b, &p); err != nil {
			return notification.Job{}, err
		}
		return notification.Job{UserID: p.UserID, TemplateName: "payment_completed", Data: p}, nil
	},
	contract.RoutingNotifyPaymentFailed: func(b []byte) (notification.Job, error) {
		var p contract.NotifyPaymentFailed
		if err := json.Unmarshal(b, &p); err != nil {
			return notification.Job{}, err
		}
		return notification.Job{UserID: p.UserID, TemplateName: "payment_failed", Data: p}, nil
	},
}
