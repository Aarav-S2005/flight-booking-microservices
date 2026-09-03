package store

import (
	"context"
	"fmt"
	"log"
	"maps"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Registry struct {
	snap        *FlightsSnapshot
	mu          sync.RWMutex
	lastApplied map[uuid.UUID]uint64 // tracking version
}

func NewRegistry(ctx context.Context, db *pgxpool.Pool) *Registry {
	snap, err := buildSnapShot(ctx, db)
	if err != nil {
		return nil
	}
	return &Registry{
		snap:        snap,
		lastApplied: make(map[uuid.UUID]uint64),
	}
}

func (reg *Registry) Get() *FlightsSnapshot {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	return reg.snap
}

func (reg *Registry) ApplySeatUpdate(ctx context.Context, flightID uuid.UUID, newSeat int, version uint64, db *pgxpool.Pool) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	flight, ok := reg.snap.FlightsByID[flightID]
	if !ok {
		return fmt.Errorf("flight %s not in snapshot", flightID)
	}

	if version <= reg.lastApplied[flightID] {
		return nil
	}

	_, err := db.Exec(ctx, ` UPDATE flights SET seats_left = $1 WHERE id = $2 `, newSeat, flightID)
	if err != nil {
		log.Printf("failed to update seats_left: %v", err)
	}

	flight.SeatsLeft = newSeat
	reg.snap.FlightsByID[flightID] = flight
	reg.lastApplied[flightID] = version

	return nil
}

func (reg *Registry) AddFlight(flight database.Flight) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	old := reg.snap
	next := cloneForInsert(old, flight)
	reg.snap = next
}

func cloneForInsert(old *FlightsSnapshot, f database.Flight) *FlightsSnapshot {
	next := &FlightsSnapshot{
		FlightsByID:     maps.Clone(old.FlightsByID),
		AirportsByCode:  old.AirportsByCode,
		ByAirlineName:   maps.Clone(old.ByAirlineName),
		ByDepartureDate: maps.Clone(old.ByDepartureDate),
		ByArrivalDate:   maps.Clone(old.ByArrivalDate),
		ByTimeBucket:    maps.Clone(old.ByTimeBucket),
		BySourceAirport: maps.Clone(old.BySourceAirport),
		ByRoute:         maps.Clone(old.ByRoute),
		AdjBySource:     maps.Clone(old.AdjBySource),

		ListByPriceAsc:         old.ListByPriceAsc,
		ListByDurationAsc:      old.ListByDurationAsc,
		ListByDepartureTimeAsc: old.ListByDepartureTimeAsc,
	}

	next.FlightsByID[f.Id] = f

	next.ByAirlineName[f.AirlineName] = appendCopy(next.ByAirlineName[f.AirlineName], f.Id)
	next.ByDepartureDate[NewDateKey(f.DepartureTime)] = appendCopy(next.ByDepartureDate[NewDateKey(f.DepartureTime)], f.Id)
	next.ByArrivalDate[NewDateKey(f.ArrivalTime)] = appendCopy(next.ByArrivalDate[NewDateKey(f.ArrivalTime)], f.Id)
	next.ByTimeBucket[NewTimeBucket(f.DepartureTime)] = appendCopy(next.ByTimeBucket[NewTimeBucket(f.DepartureTime)], f.Id)
	next.BySourceAirport[f.SourceAirportCode] = appendCopy(next.BySourceAirport[f.SourceAirportCode], f.Id)
	routeKey := RouteKey{
		Source: f.SourceAirportCode,
		Dest:   f.DestinationAirportCode,
	}
	next.ByRoute[routeKey] = appendCopy(next.ByRoute[routeKey], f.Id)
	next.AdjBySource[f.SourceAirportCode] = appendCopy(next.AdjBySource[f.SourceAirportCode], f.Id)

	next.ListByPriceAsc = insertSorted(old.ListByPriceAsc, f.Id, func(a, b uuid.UUID) bool {
		fa := old.FlightsByID[a]
		fb := old.FlightsByID[b]

		if fa.Price != fb.Price {
			return fa.Price < fb.Price
		}
		return a.String() < b.String()
	})
	next.ListByDurationAsc = insertSorted(old.ListByDurationAsc, f.Id, func(a, b uuid.UUID) bool {
		fa := old.FlightsByID[a]
		fb := old.FlightsByID[b]

		if fa.DurationInMins != fb.DurationInMins {
			return fa.DurationInMins < fb.DurationInMins
		}
		return a.String() < b.String()
	})
	next.ListByDepartureTimeAsc = insertSorted(old.ListByDepartureTimeAsc, f.Id, func(a, b uuid.UUID) bool {
		fa := old.FlightsByID[a]
		fb := old.FlightsByID[b]

		if !fa.DepartureTime.Equal(fb.DepartureTime) {
			return fa.DepartureTime.Before(fb.DepartureTime)
		}
		return a.String() < b.String()
	})

	return next
}

func appendCopy(old []uuid.UUID, id uuid.UUID) []uuid.UUID {
	next := make([]uuid.UUID, len(old)+1)
	copy(next, old)
	next[len(old)] = id
	return next
}

func insertSorted(old []uuid.UUID, id uuid.UUID, less func(a, b uuid.UUID) bool) []uuid.UUID {
	i := sort.Search(len(old), func(i int) bool { return less(id, old[i]) })
	next := make([]uuid.UUID, len(old)+1)
	copy(next, old[:i])
	next[i] = id
	copy(next[i+1:], old[i:])
	return next
}
