package store

import (
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/schema"
	"github.com/google/uuid"
)

type TimeBucket string

const (
	// all inclusive times
	Morning   TimeBucket = "morning"    // 06:00-12:00
	Afternoon TimeBucket = "afternoon"  // 12:00-18:00
	Night     TimeBucket = "night"      // 18:00-24:00
	LateNight TimeBucket = "late_night" // 00:00-06:00
)

type DateKey struct {
	Year  int
	Month time.Month
	Day   int
}

type RouteKey struct {
	Source string
	Dest   string
}

type FlightsSnapshot struct {
	FlightsByID    map[uuid.UUID]schema.Flight
	AirportsByCode map[string]schema.Airport
	Version        uint64

	ByAirlineName   map[string][]uuid.UUID
	ByDepartureDate map[DateKey][]uuid.UUID
	ByArrivalDate   map[DateKey][]uuid.UUID
	ByTimeBucket    map[TimeBucket][]uuid.UUID
	BySourceAirport map[string][]uuid.UUID
	ByRoute         map[RouteKey][]uuid.UUID

	AdjBySource map[string][]uuid.UUID

	ListByPriceAsc         []uuid.UUID
	ListByDurationAsc      []uuid.UUID
	ListByDepartureTimeAsc []uuid.UUID
}
