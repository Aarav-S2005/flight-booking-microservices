package endpoint

import (
	"context"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/store"
)

type Service struct {
	snapshot *store.Registry
}

func NewService(snapshot *store.Registry) *Service {
	return &Service{snapshot: snapshot}
}

func (s *Service) searchFlights(ctx context.Context, query Query) SearchResponse {
	snap := s.snapshot.Snap.Load()
	routes := searchValidRoutes(snap, query)
	return SearchResponse{Flights: routes}
}

func searchValidRoutes(snap *store.FlightsSnapshot, query Query) []Route {
	routes := make([]Route, 0)
	dfs(snap, query.SourceAirport, 0, query, &routes, Route{}, 0, 0, make(map[string]bool))
	return routes
}

func dfs(snap *store.FlightsSnapshot, curAirport string, stops int, query Query, routes *[]Route, curRoute Route, price, duration int, visited map[string]bool) {
	if curAirport == query.DestinationAirport {
		if len(curRoute.Segments) == 0 {
			return
		}
		if query.ArrivalDate != nil {
			lastSegment := curRoute.Segments[len(curRoute.Segments)-1]

			if !store.MatchesDate(lastSegment.ArrivalTime, query.ArrivalDate) {
				return
			}
		}
		curRoute.Price = price
		curRoute.TotalDuration = duration

		route := curRoute
		route.Segments = append([]Segment(nil), curRoute.Segments...)
		*routes = append(*routes, route)
		return
	}
	for _, flightID := range snap.AdjBySource[curAirport] {
		flight, ok := snap.FlightsByID[flightID]
		// flight existence check
		if !ok {
			continue
		}
		// airline check
		if query.AirlineName != "" && flight.AirlineName != query.AirlineName {
			continue
		}
		// departure check
		isFirstSegment := len(curRoute.Segments) == 0
		if isFirstSegment && query.DepartureDate != nil {
			if !store.MatchesDate(flight.DepartureTime, query.DepartureDate) {
				continue
			}
			if !store.MatchesTimeBucket(flight.DepartureTime, query.TimeBucket) {
				continue
			}
		}
		// is next destination visited already
		nextAirport := flight.DestinationAirportCode
		if visited[nextAirport] {
			continue
		}
		// check layover
		if !isFirstSegment {
			lastSegment := curRoute.Segments[len(curRoute.Segments)-1]
			var minLayover time.Duration
			if lastSegment.AirlineName == flight.AirlineName {
				minLayover = 45 * time.Minute
			} else {
				minLayover = 90 * time.Minute
			}

			earliestDeparture := lastSegment.ArrivalTime.Add(minLayover)
			latestDeparture := lastSegment.ArrivalTime.Add(24 * time.Hour)
			if flight.DepartureTime.Before(earliestDeparture) || flight.DepartureTime.After(latestDeparture) {
				continue
			}
		}
		nextStops := stops

		if nextAirport != query.DestinationAirport {
			nextStops++
		}
		if nextStops > query.Stops || nextStops > 3 {
			continue
		}
		segment := Segment{
			FlightNumber:       flight.FlightNumber,
			AirlineName:        flight.AirlineName,
			SourceAirport:      flight.SourceAirportCode,
			DestinationAirport: flight.DestinationAirportCode,
			DepartureTime:      flight.DepartureTime,
			ArrivalTime:        flight.ArrivalTime,
			Duration:           flight.DurationInMins,
		}
		nextRoute := curRoute

		nextRoute.Segments = append(append([]Segment(nil), curRoute.Segments...), segment)
		nextRoute.SourceAirport = query.SourceAirport
		nextRoute.DestinationAirport = query.DestinationAirport

		segPrice := flight.Price
		if !isFirstSegment {
			segPrice = flight.Price * 90 / 100
		}
		nextPrice := price + segPrice
		firstDeparture := nextRoute.Segments[0].DepartureTime

		nextDuration := int(flight.ArrivalTime.Sub(firstDeparture).Minutes())
		visited[nextAirport] = true
		dfs(snap, nextAirport, nextStops, query, routes, nextRoute, nextPrice, nextDuration, visited)
		delete(visited, nextAirport)
	}
}
