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
			booking_user_id UUID NOT NULL,
			reservation_id UUID,
			payment_id UUID,	
			email TEXT,
			phone TEXT,
			total_fare INTEGER NOT NULL DEFAULT 0,
			status booking_status NOT NULL DEFAULT 'PAYMENT_PENDING',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		    
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
			seats_left INT NOT NULL DEFAULT 0
			CHECK (seats_left >= 0),			
		);
	`
)
