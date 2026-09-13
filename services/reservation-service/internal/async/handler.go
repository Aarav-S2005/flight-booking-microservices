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

func HandleBookingConfirmed(db *pgxpool.Pool) rabbitmq.Handler {
	return func(ctx context.Context, msg amqp.Delivery) rabbitmq.Action {
		var message contract.BookingConfirmedForReservationEvent
		err := json.Unmarshal(msg.Body, &message)
		if err != nil {
			log.Println("failed to unmarshal message: ", err)
			return rabbitmq.NackDiscard
		}
		// add to db
		return InsertReservation(ctx, db, message)
	}
}

func InsertReservation(ctx context.Context, db *pgxpool.Pool, event contract.BookingConfirmedForReservationEvent) rabbitmq.Action {
	tx, err := db.Begin(ctx)
	if err != nil {
		return rabbitmq.NackRequeue
	}
	defer tx.Rollback(ctx)

	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		log.Println("invalid booking_id: %w", err)
		return rabbitmq.NackDiscard
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		log.Println("invalid user_id: %w", err)
		return rabbitmq.NackDiscard
	}

	var reservationID uuid.UUID
	err = tx.QueryRow(ctx, `
		insert into reservations (
			user_id,
			booking_id,
			passenger_count
		)
		values ($1, $2, $3)
		ON CONFLICT (booking_id)
		DO UPDATE SET booking_id = EXCLUDED.booking_id
		RETURNING reservation_id;
		`, userID, bookingID, len(event.PassengerIDs)).Scan(&reservationID)

	if err != nil {
		log.Println("failed to insert reservation: %w", err)
		return rabbitmq.NackRequeue
	}

	for i, flightIDStr := range event.FlightSegments {
		flightID, err := uuid.Parse(flightIDStr)
		if err != nil {
			log.Println("invalid flight_id: %w", err)
			return rabbitmq.NackDiscard
		}

		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO reservation_flights (
				reservation_id,
				flight_id,
				aircraft_type,
				flight_departure_time,
				segment_number
			)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (reservation_id, flight_id)
			DO NOTHING;
			`,
			reservationID,
			flightID,
			nil,
			nil,
			i+1,
		)

		if err != nil {
			log.Println("failed to insert reservation flight: %w", err)
			return rabbitmq.NackRequeue
		}
	}

	for _, flightIDStr := range event.FlightSegments {
		flightID, err := uuid.Parse(flightIDStr)
		if err != nil {
			log.Println("invalid flight_id: %w", err)
			return rabbitmq.NackDiscard
		}

		for _, passengerIDStr := range event.PassengerIDs {
			passengerID, err := uuid.Parse(passengerIDStr)
			if err != nil {
				log.Println("invalid passenger_id: %w", err)
				return rabbitmq.NackDiscard
			}

			_, err = tx.Exec(
				ctx,
				`
				INSERT INTO seat_allocation (
					reservation_id,
					flight_id,
					column_allocated,
					seat_number,
					passenger_id
				)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (reservation_id, flight_id, passenger_id)
				DO NOTHING;
				`,
				reservationID,
				flightID,
				nil,
				nil,
				passengerID,
			)

			if err != nil {
				log.Println("failed to insert seat allocation: %w", err)
				return rabbitmq.NackRequeue
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("failed to commit reservation: %w", err)
		return rabbitmq.NackRequeue
	}

	return rabbitmq.Ack
}
