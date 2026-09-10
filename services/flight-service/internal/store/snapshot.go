package store

import (
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/database"
	"github.com/google/uuid"
)

type TimeBucket string

const (
	// [x, y), start time inclusive, end time exclusive
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
	FlightsByID    map[uuid.UUID]database.Flight
	AirportsByCode map[string]database.Airport

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

func NewDateKey(t time.Time) DateKey {
	return DateKey{
		Year:  t.Year(),
		Month: t.Month(),
		Day:   t.Day(),
	}
}

func NewTimeBucket(t time.Time) TimeBucket {
	hour := t.Hour()

	switch {
	case hour >= 6 && hour < 12:
		return Morning
	case hour >= 12 && hour < 18:
		return Afternoon
	case hour >= 18 && hour < 24:
		return Night
	default:
		return LateNight
	}
}

func MatchesTimeBucket(t time.Time, buckets []TimeBucket) bool {
	if len(buckets) == 0 {
		return true
	}

	bucket := NewTimeBucket(t)

	for _, allowed := range buckets {
		if bucket == allowed {
			return true
		}
	}

	return false
}

func MatchesDate(t time.Time, date *DateKey) bool {
	if date == nil {
		return true
	}

	return t.Year() == date.Year &&
		t.Month() == date.Month &&
		t.Day() == date.Day
}
