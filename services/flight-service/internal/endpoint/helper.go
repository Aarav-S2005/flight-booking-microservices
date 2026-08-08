package endpoint

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/store"
)

type SortBy string

const (
	Price         SortBy = "price"
	DepartureDate SortBy = "departure_date"
	Duration      SortBy = "duration"
)

type Query struct {
	AirlineName        string
	TimeBucket         []store.TimeBucket // enum in store
	DepartureDate      *store.DateKey
	ArrivalDate        *store.DateKey
	SourceAirport      string
	DestinationAirport string
	Stops              int
	SortedBy           SortBy
}

func ParseQuery(r *http.Request) (Query, error) {
	q := r.URL.Query()

	depDate, err := parseDateKey(q.Get("departureDate"))
	if err != nil {
		return Query{}, fmt.Errorf("invalid departureDate: %w", err)
	}
	arrDate, err := parseDateKey(q.Get("arrivalDate"))
	if err != nil {
		return Query{}, fmt.Errorf("invalid arrivalDate: %w", err)
	}

	var stops int
	if q.Get("stops") == "" {
		stops = 0
	} else {
		var err error
		stops, err = strconv.Atoi(q.Get("stops"))
		if err != nil {
			return Query{}, err
		}
	}

	query := Query{
		AirlineName:        q.Get("airlineName"),
		TimeBucket:         make([]store.TimeBucket, 0),
		DepartureDate:      depDate,
		ArrivalDate:        arrDate,
		SourceAirport:      q.Get("sourceAirport"),
		DestinationAirport: q.Get("destinationAirport"),
		Stops:              stops,
		SortedBy:           SortBy(q.Get("sortedBy")),
	}

	for _, tb := range q["timeBucket"] {
		temp := store.TimeBucket(tb)
		if temp != "" {
			switch temp {
			case store.Morning, store.Afternoon, store.Night, store.LateNight:
			default:
				return Query{}, fmt.Errorf("invalid timeBucket: %q", temp)
			}
		}
		query.TimeBucket = append(query.TimeBucket, temp)
	}

	if err := validateQuery(query); err != nil {
		return Query{}, err
	}
	return query, nil
}

func validateQuery(q Query) error {
	if q.SortedBy != "" {
		switch q.SortedBy {
		case Price, DepartureDate, Duration:
		default:
			return fmt.Errorf("invalid sortedBy: %q", q.SortedBy)
		}
	}

	if q.Stops > 3 {
		return fmt.Errorf("invalid stops: %d, maximum 3 allowed", q.Stops)
	}
	if q.Stops < 0 {
		return fmt.Errorf("invalid stops: %d, minimum 0 allowed", q.Stops)
	}

	if q.SourceAirport == "" || q.DestinationAirport == "" {
		return fmt.Errorf("source or Destination airport is required")
	}

	if q.SourceAirport == q.DestinationAirport {
		return fmt.Errorf("sourceAirport and sourceDestination cannot be the same")
	}

	return nil
}

func parseDateKey(s string) (*store.DateKey, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}

	return &store.DateKey{
		Year:  t.Year(),
		Month: t.Month(),
		Day:   t.Day(),
	}, nil
}
