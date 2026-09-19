package async

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/database"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandleBookingConfirmed(repo *database.Repository) rabbitmq.Handler {
	return func(ctx context.Context, msg amqp.Delivery) rabbitmq.Action {
		var message contract.BookingConfirmedForReservationEvent
		err := json.Unmarshal(msg.Body, &message)
		if err != nil {
			log.Println("failed to unmarshal message: ", err)
			return rabbitmq.NackDiscard
		}
		bookingID, err := uuid.Parse(message.BookingID)
		if err != nil {
			log.Println("failed to parse booking id: ", err)
			return rabbitmq.NackDiscard
		}
		userID, err := uuid.Parse(message.UserID)
		if err != nil {
			log.Println("failed to parse user id: ", err)
			return rabbitmq.NackDiscard
		}
		err = repo.InsertReservationWithoutSeatReservation(ctx, bookingID, userID, message.PassengerIDs, message.FlightSegments)
		if err != nil {
			if errors.Is(err, database.ErrExternal) {
				log.Println("failed to insert reservation: ", err)
				return rabbitmq.NackDiscard
			}
			return rabbitmq.NackRequeue
		}
		return rabbitmq.Ack
	}
}
