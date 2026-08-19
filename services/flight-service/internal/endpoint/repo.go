package endpoint

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (repo *Repository) getAirportByCode(ctx context.Context, airportCode string) (string, error) {
	var name string
	err := repo.db.QueryRow(ctx, "select name from airports where airport_code = $1", airportCode).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (repo *Repository) createFlight(ctx context.Context, reqBody CreateFlightDTO, seatsLeft int) (uuid.UUID, error) {
	var flightID uuid.UUID
	err := repo.db.QueryRow(ctx, `
		INSERT INTO flights (
			flight_number,
			airline_name,
			aircraft_type,
			seats_left,
			source_airport_code,
			destination_airport_code,
			departure_time,
			arrival_time,
			duration,
			price
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			EXTRACT(EPOCH FROM ($8 - $7)) / 60)::INTEGER,
			$9
		)
		RETURNING flight_id
	`,
		reqBody.FlightNumber,
		reqBody.AirlineName,
		reqBody.AircraftType,
		seatsLeft,
		reqBody.SourceAirportCode,
		reqBody.DestinationAirportCode,
		reqBody.DepartureTime,
		reqBody.ArrivalTime,
		reqBody.Price,
	).Scan(&flightID)
	if err != nil {
		return uuid.Nil, err
	}

	return flightID, nil
}
