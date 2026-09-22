package async

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/database"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandleBookingConfirmed(repo *database.Repository, client *resty.Client, flightServiceURL string) rabbitmq.Handler {
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

		flights := make([]database.FlightsSchema, 0, len(message.FlightSegments))

		for _, segment := range message.FlightSegments {
			flightID, err := uuid.Parse(segment)
			if err != nil {
				log.Println("failed to parse flight id:", err)
				return rabbitmq.NackDiscard
			}
			var flight FlightResponse
			resp, err := client.R().SetContext(ctx).SetResult(&flight).Get(fmt.Sprintf("%s/flight/%s", flightServiceURL, segment))
			if err != nil {
				log.Println("failed to get flight details:", err)
				return rabbitmq.NackRequeue
			}

			if resp.IsError() {
				log.Printf("flight service returned %d for flight %s", resp.StatusCode(), flightID.String())
				return rabbitmq.NackRequeue
			}

			flights = append(flights, database.FlightsSchema{
				FlightID:      flightID,
				AircraftType:  flight.AircraftType,
				DepartureTime: flight.DepartureTime,
			})
		}

		if err := repo.InsertFlights(ctx, flights); err != nil {
			if errors.Is(err, database.ErrExternal) {
				log.Println("failed to insert flights:", err)
				return rabbitmq.NackDiscard
			}
			return rabbitmq.NackRequeue
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

type FlightResponse struct {
	FlightID      string     `json:"flight_id"`
	AircraftType  string     `json:"aircraft_type"`
	DepartureTime *time.Time `json:"departure_time"`
}
