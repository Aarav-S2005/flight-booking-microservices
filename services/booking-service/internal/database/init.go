package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	bookingStatusEnum = `
	DO $$
		BEGIN
			CREATE TYPE booking_status AS ENUM (
				'PAYMENT_PENDING',
				'CONFIRMED',
				'FAILED'
			);
		EXCEPTION
			WHEN duplicate_object THEN NULL;
	END
	$$;
	`

	bookingsTable = `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS bookings (
		  	booking_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			booking_user_id UUID NOT NULL,
			reservation_id UUID,
			email TEXT NOT NULL,
			phone TEXT NOT NULL,
			total_fare INTEGER NOT NULL DEFAULT 0,
			status booking_status NOT NULL DEFAULT 'PAYMENT_PENDING',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`

	passengersTable = `
		CREATE TABLE IF NOT EXISTS passengers (
		   passenger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		   booking_id UUID NOT NULL,
		   first_name VARCHAR(100) NOT NULL,
		   last_name VARCHAR(100) NOT NULL,
		   age int NOT NULL,
		   gender VARCHAR(20) NOT NULL,
		   passport_number VARCHAR(50) NOT NULL,
		   CONSTRAINT fk_passengers_booking
			  FOREIGN KEY (booking_id)
			  REFERENCES bookings(booking_id)
			  ON DELETE CASCADE
		);
	`

	flightSegmentTable = `
		CREATE TABLE IF NOT EXISTS flight_segments (
			booking_id UUID NOT NULL,
			flight_id UUID NOT NULL,
			segment_order INTEGER NOT NULL,
		
			PRIMARY KEY (booking_id, segment_order),
		
			CONSTRAINT fk_flight_segments_booking
				FOREIGN KEY (booking_id)
				REFERENCES bookings(booking_id)
				ON DELETE CASCADE,
		
			CONSTRAINT uq_flight_segments_booking_flight
				UNIQUE (booking_id, flight_id),
		
			CONSTRAINT chk_segment_order
				CHECK (segment_order > 0)
		);
	`

	flightsTable = `
		CREATE TABLE IF NOT EXISTS flights (
			flight_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			aircraft_type VARCHAR(40) NOT NULL,
			seats_left INT NOT NULL,
			total_seats INT NOT NULL,
			version BIGINT NOT NULL,
			CHECK (seats_left >= 0)
		);
	`
)

func InitSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, bookingStatusEnum)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, bookingsTable)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, passengersTable)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, flightSegmentTable)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, flightsTable)
	if err != nil {
		return err
	}
	return nil
}
