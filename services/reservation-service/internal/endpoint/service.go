package endpoint

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/reservation-service/internal/database"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

const (
	maxLeadTime = 48 * time.Hour
	minLeadTime = 4 * time.Hour
)

type Service struct {
	repo              *database.Repository
	publisher         *rabbitmq.Publisher
	flightServiceURL  string
	bookingServiceURL string
	client            *resty.Client
}

func NewService(repo *database.Repository, publisher *rabbitmq.Publisher, flightServiceURL string, bookingServiceURL string) *Service {
	return &Service{
		repo:              repo,
		publisher:         publisher,
		flightServiceURL:  flightServiceURL,
		bookingServiceURL: bookingServiceURL,
		client:            resty.New().SetHeader("Content-Type", "application/json"),
	}
}

type seatPick struct {
	Column string
	Number int
}

type flightPick struct {
	FlightID uuid.UUID
	Seats    []seatPick
}

func (s *Service) reserveSeats(ctx context.Context, reqBody ReserveSeatsRequestDTO, userID uuid.UUID) error {
	bookingID, err := uuid.Parse(reqBody.BookingID)
	if err != nil {
		return app_error.BadRequest("failed to parse booking id", err)
	}
	exists, err := s.repo.ExistsByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}
	if !exists {
		request := ValidateBookingForReservationRequestDTO{
			reqBody.BookingID,
			userID.String(),
		}
		var response ValidateBookingForReservationResponseDTO
		resp, err := s.client.R().SetBody(&request).SetResult(&response).Post(s.bookingServiceURL + "/validate-reservation")
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			if response.Status == "PAYMENT_PENDING" {
				return app_error.BadRequest("payment pending", errors.New(response.Status))
			}
			err = s.repo.InsertReservationWithoutSeatReservation(ctx, bookingID, userID, response.Passengers, response.FlightIDs)
			if err != nil {
				return err
			}
		}
	}
	// reserve tickets atomically
	picks, err := parseReserveRequest(reqBody)
	if err != nil {
		return err
	}
	err = s.repo.WithTx(ctx, func(tx database.SeatTx) error {
		res, err := tx.LockReservationByBooking(ctx, bookingID)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return app_error.NotFound("booking not found", err)
			}
			return err
		}
		if res.UserID != userID {
			return app_error.NotFound("booking not found", nil)
		}
		segments, err := tx.ListFlightSegments(ctx, res.ID)
		if err != nil {
			return err
		}
		segByFlight := make(map[uuid.UUID]database.FlightSegmentTxModel, len(segments))
		for _, sg := range segments {
			if _, exists := segByFlight[sg.FlightID]; !exists {
				segByFlight[sg.FlightID] = sg
			}
		}
		now := time.Now().UTC()
		for _, fp := range picks {
			seg, ok := segByFlight[fp.FlightID]
			if !ok {
				return app_error.BadRequest(fmt.Sprintf("flight %s is not part of this booking", fp.FlightID), nil)
			}
			untilDeparture := seg.DepartureTime.UTC().Sub(now)
			if untilDeparture > maxLeadTime {
				return app_error.BadRequest(fmt.Sprintf("seat selection for flight %s opens 48 hours before departure", fp.FlightID), nil)
			}
			if untilDeparture < minLeadTime {
				return app_error.BadRequest(fmt.Sprintf("seat selection for flight %s closed 4 hours before departure", fp.FlightID), nil)
			}
			validCols := make(map[string]struct{}, len(seg.Columns))
			for _, c := range seg.Columns {
				validCols[strings.ToUpper(strings.TrimSpace(c))] = struct{}{}
			}
			for _, p := range fp.Seats {
				if _, ok := validCols[p.Column]; !ok {
					return app_error.BadRequest(fmt.Sprintf("column %s does not exist on aircraft %s", p.Column, seg.AircraftType), nil)
				}
				if p.Number < 1 || p.Number > seg.TotalRows {
					return app_error.BadRequest(fmt.Sprintf("row %d does not exist on aircraft %s (rows 1-%d)", p.Number, seg.AircraftType, seg.TotalRows), nil)
				}
			}
			allocs, err := tx.LockAllocations(ctx, res.ID, fp.FlightID)
			if err != nil {
				return err
			}
			if len(allocs) == 0 {
				return fmt.Errorf("no passenger allocations found for flight %s", fp.FlightID)
			}
			for _, a := range allocs {
				if a.Column != nil || a.SeatNumber != nil {
					return app_error.Conflict(fmt.Sprintf("seats for flight %s have already been selected", fp.FlightID), nil)
				}
			}
			if len(fp.Seats) != len(allocs) {
				return app_error.BadRequest(fmt.Sprintf("flight %s requires exactly %d seat(s), got %d", fp.FlightID, len(allocs), len(fp.Seats)), nil)
			}
			for i, p := range fp.Seats {
				err := tx.AssignSeat(ctx, res.ID, fp.FlightID, allocs[i].PassengerID, p.Column, p.Number)
				if err != nil {
					if errors.Is(err, database.ErrSeatTaken) {
						return app_error.Conflict(fmt.Sprintf("seat %d%s on flight %s is already taken", p.Number, p.Column, fp.FlightID), err)
					}
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return toAppErr(err)
	}
	return nil
}

func (s *Service) getAircraftDetails(ctx context.Context, aircraftType string) (GetFlightSeatsFromReservationRequest, error) {
	totalSeats, err := s.repo.GetSeatsLeftByAircraftType(ctx, aircraftType)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return GetFlightSeatsFromReservationRequest{}, app_error.NotFound("aircraft not found", err)
		}
		return GetFlightSeatsFromReservationRequest{}, err
	}
	return GetFlightSeatsFromReservationRequest{SeatsLeft: totalSeats}, nil
}
