package store

import (
	"context"
	"fmt"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/schema"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRegistry() *Registry {
	return &Registry{}
}

func (reg *Registry) buildSnapShot(ctx context.Context, db *pgxpool.Pool) error {
	airports, err := getAllAirportsFromDB(ctx, db)
	if err != nil {
		return err
	}
	airportsByCode := make(map[string]schema.Airport)
	for _, airport := range airports {
		airportsByCode[airport.AirportCode] = airport
	}
	return nil
}

func getAllAirportsFromDB(ctx context.Context, db *pgxpool.Pool) ([]schema.Airport, error) {
	rows, err := db.Query(ctx, "SELECT * FROM airports")
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var airports []schema.Airport
	for rows.Next() {
		var a schema.Airport

		err := rows.Scan(&a.AirportCode, &a.AirportName, &a.City, &a.Country)
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		airports = append(airports, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return airports, nil
}

func getFlightsFromDB(ctx context.Context, db *pgxpool.Pool) ([]schema.Flight, error) {
	now := time.Now()
	rows, err := db.Query(ctx, "SELECT * FROM flights where departure_time < $1", now)
	return nil, nil
}
