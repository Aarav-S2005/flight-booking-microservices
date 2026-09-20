package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createAirportTableSQL = `
		CREATE TABLE IF NOT EXISTS airports (
			airport_code CHAR(3) PRIMARY KEY,
			airport_name VARCHAR(255) NOT NULL,
			city VARCHAR(100) NOT NULL,
			country VARCHAR(100) NOT NULL
		);
	`

	createFlightTableSQL = `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;

		CREATE TABLE IF NOT EXISTS flights (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			flight_number VARCHAR(6) NOT NULL UNIQUE,
			airline_name VARCHAR(255) NOT NULL,
			aircraft_type varchar(20) NOT NULL,
			seats_left int NOT NULL,
			source_airport_code CHAR(3) NOT NULL,
			destination_airport_code CHAR(3) NOT NULL,
			departure_time TIMESTAMPTZ NOT NULL,
			arrival_time TIMESTAMPTZ NOT NULL,
			duration INTEGER NOT NULL,
			price INTEGER NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		
			CONSTRAINT fk_source_airport
				FOREIGN KEY (source_airport_code)
				REFERENCES airports (airport_code),
		
			CONSTRAINT fk_destination_airport
				FOREIGN KEY (destination_airport_code)
				REFERENCES airports (airport_code),
		
			CONSTRAINT chk_duration_positive
				CHECK (duration > 0),
		
			CONSTRAINT chk_different_airports
				CHECK (source_airport_code <> destination_airport_code),

			CONSTRAINT chk_arrival_after_departure
				CHECK (arrival_time > departure_time)
		);
	`
)

func InitSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, createAirportTableSQL)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, createFlightTableSQL)
	if err != nil {
		return err
	}
	return nil
}
