package async

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

func AddNewDuePaymentRecord(db *pgxpool.Pool) rabbitmq.Handler {
	return func(ctx context.Context, msg amqp.Delivery) rabbitmq.Action {
		var body contract.BookingConfirmedForPaymentEvent
		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			log.Printf("unmarshal failed for %q: %v", msg.RoutingKey, err)
			return rabbitmq.NackDiscard
		}
		userID, err := uuid.Parse(body.UserID)
		if err != nil {
			log.Printf("parse failed for %q: %v", msg.RoutingKey, err)
			return rabbitmq.NackDiscard
		}
		bookingID, err := uuid.Parse(body.BookingID)
		if err != nil {
			log.Printf("parse failed for %q: %v", msg.RoutingKey, err)
			return rabbitmq.NackDiscard
		}
		_, err = db.Exec(ctx, "insert into payments(user_id, booking_id, amount) values ($1, $2, $3) ON CONFLICT (user_id, booking_id) DO NOTHING", userID, bookingID, body.TotalFare)
		if err != nil {
			log.Printf("insert failed for %q: %v", msg.RoutingKey, err)
			return rabbitmq.NackRequeue
		}
		return rabbitmq.Ack
	}
}
