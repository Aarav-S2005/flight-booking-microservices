package schema

const (
	bookingStatusEnum = `
	DO $$
		BEGIN
			CREATE TYPE booking_status AS ENUM (
				'PAYMENT_PENDING',
				'CONFIRMED',
				'FAILED',
				'CANCELLED',
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
		    booking_user_id UUID NOT NULL UNIQUE,
		    flight_id UUID NOT NULL UNIQUE,
		    reservation_id UUID NOT NULL UNIQUE,
			payment_id UUID NOT NULL UNIQUE,
			email STRING,
			phone STRING,
			total_fare INTEGER NOT NULL,
			status booking_status NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		);
	`

	passengersTable = `
		CREATE TABLE IF NOT EXISTS passengers (
			passenger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			booking_id UUID REFERENCES bookings(id) ON DELETE CASCADE,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			age INT NOT NULL CHECK (age >= 0),
			gender VARCHAR(20) NOT NULL,
			passport_number VARCHAR(50),
			seat_number VARCHAR(20),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`

	flightsTable = `
		CREATE TABLE IF NOT EXISTS flights (
			flight_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			aircraft_type VARCHAR(40) NOT NULL,
			seats_left INT NOT NULL DEFAULT 0
			CHECK (seats_left >= 0),			
		);
	`
)
