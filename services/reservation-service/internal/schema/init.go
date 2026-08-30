package schema

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	aircraftSchema = `
		CREATE TABLE IF NOT EXISTS aircraft (
			aircraft_type VARCHAR(12) PRIMARY KEY,
			total_seats INTEGER NOT NULL CHECK (total_seats > 0),
			columns_available VARCHAR(1)[] NOT NULL,
			total_rows INTEGER NOT NULL CHECK (total_rows > 0)
		);
	`
	reservationSchema = `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS reservation (
		  	reservation_id UUID PRIMARY KEY,
			booking_id UUID NOT NULL UNIQUE,
			passenger_count INTEGER NOT NULL
		)
	`
	reservationFlightsSchema = `
		create table reservation_flights (
		    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
			flight_id UUID NOT NULL,
			segment_number INTEGER NOT NULL CHECK (segment_number > 0)
		)
	`
	seatAllocationSchema = `
		CREATE TABLE IF NOT EXISTS seat_allocation (
		    reservation_id UUID NOT NULL REFERENCES reservations(id)ON DELETE CASCADE,
			flight_id UUID NOT NULL,
		    seat_alloted varchar(4) NOT NULL,
		    passenger_id UUID NOT NULL,
		)
	`
)

func InitSchema(ctx context.Context, db *pgxpool.Pool) error {

	if _, err := db.Exec(ctx, aircraftSchema); err != nil {
		return err
	}
	if err := SeedAircraft(ctx, db); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, aircraftSchema); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, aircraftSchema); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, aircraftSchema); err != nil {
		return err
	}
	return nil
}

type aircraftSeed struct {
	aircraftType string
	totalSeats   int
	columns      []string
	totalRows    int
}

func SeedAircraft(ctx context.Context, pool *pgxpool.Pool) error {
	aircraft := []aircraftSeed{
		// Boeing
		{
			aircraftType: "B737",
			totalSeats:   162,
			columns:      []string{"A", "B", "C", "D", "E", "F"},
			totalRows:    27,
		},
		{
			aircraftType: "B737-MAX8",
			totalSeats:   178,
			columns:      []string{"A", "B", "C", "D", "E", "F"},
			totalRows:    30,
		},
		{
			aircraftType: "B747",
			totalSeats:   416,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K"},
			totalRows:    42,
		},
		{
			aircraftType: "B747-8I",
			totalSeats:   410,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K"},
			totalRows:    41,
		},
		{
			aircraftType: "B777",
			totalSeats:   396,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K"},
			totalRows:    40,
		},
		{
			aircraftType: "B777-300",
			totalSeats:   396,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K"},
			totalRows:    40,
		},
		{
			aircraftType: "B777-300ER",
			totalSeats:   396,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K"},
			totalRows:    40,
		},
		{
			aircraftType: "B787-9",
			totalSeats:   296,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J"},
			totalRows:    33,
		},
		{
			aircraftType: "B787-10",
			totalSeats:   330,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J"},
			totalRows:    37,
		},

		// Airbus
		{
			aircraftType: "A320",
			totalSeats:   180,
			columns:      []string{"A", "B", "C", "D", "E", "F"},
			totalRows:    30,
		},
		{
			aircraftType: "A321",
			totalSeats:   220,
			columns:      []string{"A", "B", "C", "D", "E", "F"},
			totalRows:    37,
		},
		{
			aircraftType: "A330",
			totalSeats:   300,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H"},
			totalRows:    38,
		},
		{
			aircraftType: "A340",
			totalSeats:   350,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H"},
			totalRows:    44,
		},
		{
			aircraftType: "A350-900",
			totalSeats:   325,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J"},
			totalRows:    36,
		},
		{
			aircraftType: "A350-1000",
			totalSeats:   366,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K"},
			totalRows:    37,
		},
		{
			aircraftType: "A380-800",
			totalSeats:   525,
			columns:      []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K"},
			totalRows:    53,
		},
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin aircraft seed transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const query = `
		INSERT INTO aircraft (
			aircraft_type,
			total_seats,
			columns_available,
			total_rows
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (aircraft_type)
		DO UPDATE SET
			total_seats = EXCLUDED.total_seats,
			columns_available = EXCLUDED.columns_available,
			total_rows = EXCLUDED.total_rows
	`

	for _, a := range aircraft {
		if _, err := tx.Exec(
			ctx,
			query,
			a.aircraftType,
			a.totalSeats,
			a.columns,
			a.totalRows,
		); err != nil {
			return fmt.Errorf("seed aircraft %q: %w", a.aircraftType, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit aircraft seed: %w", err)
	}

	return nil
}
