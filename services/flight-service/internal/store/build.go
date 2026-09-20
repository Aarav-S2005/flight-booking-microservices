package store

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildSnapShot(ctx context.Context, db *pgxpool.Pool) (*FlightsSnapshot, error) {
	airports, err := getAllAirportsFromDB(ctx, db)
	if err != nil {
		return nil, err
	}
	flights, err := getFlightsFromDB(ctx, db)
	if err != nil {
		return nil, err
	}

	snapshot := FlightsSnapshot{
		FlightsByID:            make(map[uuid.UUID]database.Flight),
		AirportsByCode:         make(map[string]database.Airport),
		ByAirlineName:          make(map[string][]uuid.UUID),
		ByDepartureDate:        make(map[DateKey][]uuid.UUID),
		ByArrivalDate:          make(map[DateKey][]uuid.UUID),
		ByTimeBucket:           make(map[TimeBucket][]uuid.UUID),
		BySourceAirport:        make(map[string][]uuid.UUID),
		ByRoute:                make(map[RouteKey][]uuid.UUID),
		AdjBySource:            make(map[string][]uuid.UUID),
		ListByPriceAsc:         make([]uuid.UUID, 0),
		ListByDurationAsc:      make([]uuid.UUID, 0),
		ListByDepartureTimeAsc: make([]uuid.UUID, 0),
	}

	// PUTTING AIRPORT
	for _, airport := range airports {
		snapshot.AirportsByCode[airport.AirportCode] = airport
	}

	// PITTING FLIGHTS
	for _, flight := range flights {
		snapshot.FlightsByID[flight.Id] = flight
		snapshot.ByAirlineName[flight.AirlineName] = append(snapshot.ByAirlineName[flight.AirlineName], flight.Id)
		snapshot.BySourceAirport[flight.SourceAirportCode] = append(snapshot.BySourceAirport[flight.SourceAirportCode], flight.Id)
		routeKey := RouteKey{
			Source: flight.SourceAirportCode,
			Dest:   flight.DestinationAirportCode,
		}
		snapshot.ByRoute[routeKey] = append(snapshot.ByRoute[routeKey], flight.Id)
		snapshot.AdjBySource[flight.SourceAirportCode] = append(snapshot.AdjBySource[flight.SourceAirportCode], flight.Id)
		departureKey := DateKey{
			Year:  flight.DepartureTime.Year(),
			Month: flight.DepartureTime.Month(),
			Day:   flight.DepartureTime.Day(),
		}
		snapshot.ByDepartureDate[departureKey] = append(snapshot.ByDepartureDate[departureKey], flight.Id)
		arrivalKey := DateKey{
			Year:  flight.ArrivalTime.Year(),
			Month: flight.ArrivalTime.Month(),
			Day:   flight.ArrivalTime.Day(),
		}
		snapshot.ByArrivalDate[arrivalKey] = append(snapshot.ByArrivalDate[arrivalKey], flight.Id)

		hour := flight.DepartureTime.Hour()

		var bucket TimeBucket
		switch {
		case hour >= 6 && hour < 12:
			bucket = Morning
		case hour >= 12 && hour < 18:
			bucket = Afternoon
		case hour >= 18 && hour < 24:
			bucket = Night
		default: // 00:00 - 05:59
			bucket = LateNight
		}

		snapshot.ByTimeBucket[bucket] = append(snapshot.ByTimeBucket[bucket], flight.Id)

		snapshot.ListByPriceAsc = append(snapshot.ListByPriceAsc, flight.Id)
		snapshot.ListByDurationAsc = append(snapshot.ListByDurationAsc, flight.Id)
		snapshot.ListByDepartureTimeAsc = append(snapshot.ListByDepartureTimeAsc, flight.Id)
	}
	slices.SortFunc(snapshot.ListByPriceAsc, func(a, b uuid.UUID) int {
		fa := snapshot.FlightsByID[a]
		fb := snapshot.FlightsByID[b]

		if fa.Price < fb.Price {
			return -1
		}
		if fa.Price > fb.Price {
			return 1
		}
		return strings.Compare(a.String(), b.String())
	})
	slices.SortFunc(snapshot.ListByDurationAsc, func(a, b uuid.UUID) int {
		fa := snapshot.FlightsByID[a]
		fb := snapshot.FlightsByID[b]

		if fa.DurationInMins < fb.DurationInMins {
			return -1
		}
		if fa.DurationInMins > fb.DurationInMins {
			return 1
		}
		return strings.Compare(a.String(), b.String())
	})
	slices.SortFunc(snapshot.ListByDepartureTimeAsc, func(a, b uuid.UUID) int {
		fa := snapshot.FlightsByID[a]
		fb := snapshot.FlightsByID[b]

		if fa.DepartureTime.Before(fb.DepartureTime) {
			return -1
		}
		if fa.DepartureTime.After(fb.DepartureTime) {
			return 1
		}
		return strings.Compare(a.String(), b.String())
	})

	return &snapshot, nil
}

func getAllAirportsFromDB(ctx context.Context, db *pgxpool.Pool) ([]database.Airport, error) {
	rows, err := db.Query(ctx, "SELECT * FROM airports")
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	airports, err := pgx.CollectRows(rows, pgx.RowToStructByName[database.Airport])
	if err != nil {
		return nil, fmt.Errorf("collect failed: %w", err)
	}

	return airports, nil
}

func getFlightsFromDB(ctx context.Context, db *pgxpool.Pool) ([]database.Flight, error) {
	now := time.Now().UTC()
	today := now
	after45Days := now.AddDate(0, 0, 45)
	rows, err := db.Query(ctx, "SELECT * FROM flights WHERE departure_time > $1 AND departure_time < $2", today, after45Days)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	flights, err := pgx.CollectRows(rows, pgx.RowToStructByName[database.Flight])
	if err != nil {
		return nil, fmt.Errorf("collect failed: %w", err)
	}
	return flights, nil
}
