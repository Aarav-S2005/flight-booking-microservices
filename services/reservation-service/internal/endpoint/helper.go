package endpoint

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/database"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/google/uuid"
)

func toAppErr(err error) error {
	switch {
	case errors.Is(err, database.ErrSeatTaken):
		return app_error.Conflict("one or more selected seats are already taken", err)
	case errors.Is(err, database.ErrRetryable):
		return app_error.Conflict("seat selection conflicted with another request, please try again", err)
	case errors.Is(err, database.ErrNotFound):
		return app_error.NotFound("resource not found", err)
	default:
		return err
	}
}

func parseReserveRequest(req ReserveSeatsRequestDTO) ([]flightPick, error) {
	if len(req.Seats.Flights) == 0 {
		return nil, app_error.BadRequest("no flights provided", nil)
	}

	seenFlights := make(map[uuid.UUID]struct{}, len(req.Seats.Flights))
	picks := make([]flightPick, 0, len(req.Seats.Flights))

	for _, f := range req.Seats.Flights {
		flightID, err := uuid.Parse(f.FlightID)
		if err != nil {
			return nil, app_error.BadRequest("invalid flight_id: "+f.FlightID, err)
		}
		if _, dup := seenFlights[flightID]; dup {
			return nil, app_error.BadRequest("duplicate flight_id: "+f.FlightID, nil)
		}
		seenFlights[flightID] = struct{}{}

		if len(f.SeatsSelected) == 0 {
			return nil, app_error.BadRequest("no seats provided for flight "+f.FlightID, nil)
		}

		seenSeats := make(map[seatPick]struct{}, len(f.SeatsSelected))
		seats := make([]seatPick, 0, len(f.SeatsSelected))

		for _, s := range f.SeatsSelected {
			col := strings.ToUpper(strings.TrimSpace(s.Column))
			if len(col) != 1 || col[0] < 'A' || col[0] > 'Z' {
				return nil, app_error.BadRequest("invalid seat column: "+s.Column, nil)
			}
			num, err := strconv.Atoi(strings.TrimSpace(s.SeatNumber))
			if err != nil || num < 1 {
				return nil, app_error.BadRequest("invalid seat number: "+s.SeatNumber, err)
			}

			p := seatPick{Column: col, Number: num}
			if _, dup := seenSeats[p]; dup {
				return nil, app_error.BadRequest(fmt.Sprintf("seat %d%s selected more than once", num, col), nil)
			}
			seenSeats[p] = struct{}{}
			seats = append(seats, p)
		}

		slices.SortFunc(seats, func(a, b seatPick) int {
			if c := cmp.Compare(a.Column, b.Column); c != 0 {
				return c
			}
			return cmp.Compare(a.Number, b.Number)
		})
		picks = append(picks, flightPick{FlightID: flightID, Seats: seats})
	}

	slices.SortFunc(picks, func(a, b flightPick) int {
		return strings.Compare(a.FlightID.String(), b.FlightID.String())
	})
	return picks, nil
}
