package store

import (
	"context"
	"fmt"
	"log"
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
