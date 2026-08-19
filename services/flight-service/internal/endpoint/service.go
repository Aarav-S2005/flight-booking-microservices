package endpoint

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/flight-service/internal/store"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type Service struct {
	registry *store.Registry
	repo     *Repository
	client   *resty.Client
}

func NewService(snapshot *store.Registry, repo *Repository, reservationURL string) *Service {
	client := resty.New().
		SetBaseURL(reservationURL).
		SetHeader("Content-Type", "application/json")
	return &Service{registry: snapshot, repo: repo, client: client}
}

func (s *Service) searchFlights(ctx context.Context, query Query) SearchResponse {
	snap := s.registry.Get()
	routes := searchValidRoutes(snap, query)
	return SearchResponse{Flights: routes}
}

func (s *Service) getFlight(ctx context.Context, flightID uuid.UUID) (GetFlightResponse, error) {
	snap := s.registry.Get()
	flight, ok := snap.FlightsByID[flightID]
	if !ok {
		return GetFlightResponse{}, app_error.NotFound("flight not found", errors.New("could not fin flight:"+flightID.String()))
	}

	sourceAirportName, err := s.repo.getAirportByCode(ctx, flight.SourceAirportCode)
	if err != nil {
		return GetFlightResponse{}, err
	}
	destinationAirportName, err := s.repo.getAirportByCode(ctx, flight.DestinationAirportCode)
	if err != nil {
		return GetFlightResponse{}, err
	}
	return GetFlightResponse{
		FlightID:               flightID.String(),
		FlightNumber:           flight.FlightNumber,
		AirlineName:            flight.AirlineName,
		AircraftType:           flight.AircraftType,
		SourceAirportCode:      flight.SourceAirportCode,
		SourceAirportName:      sourceAirportName,
		DestinationAirportCode: flight.DestinationAirportCode,
		DestinationAirportName: destinationAirportName,
		DepartureTime:          flight.DepartureTime,
		ArrivalTime:            flight.ArrivalTime,
		DurationInMins:         flight.DurationInMins,
		Price:                  flight.Price,
	}, nil
}

func (s *Service) createFlight(ctx context.Context, reqBody CreateFlightDTO) error {
	var resBody GetFlightSeatsFromBookingResponse
	resp, err := s.client.R().SetQueryParam("aircraftType", reqBody.AircraftType).SetResult(&resBody).Get("/flight-type")
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("failed to get flight type: status %d", resp.StatusCode())
	}
	err = s.repo.createFlight(ctx, reqBody, resBody.SeatsLeft)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) checkAllFlightIds(ctx context.Context, flightIDs []uuid.UUID) error {
	snap := s.registry.Get()
	for _, id := range flightIDs {
		_, ok := snap.FlightsByID[id]
		if !ok {
			return app_error.NotFound("flight not found", errors.New("could not find flight:"+id.String()))
		}
	}
	return nil
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
			FlightID:           flightID.String(),
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
