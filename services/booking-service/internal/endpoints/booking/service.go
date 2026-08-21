package booking

import (
	"context"
	"errors"
	"net/http"

	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

type Service struct {
	repo                  *Repository
	client                *resty.Client
	flightServiceURL      string
	reservationServiceURL string
	// RabbitMQ also needed

	flightGroup singleflight.Group
}

func NewService(repo *Repository, flightServiceURL, reservationServiceURL string) *Service {
	client := resty.New().
		SetHeader("Content-Type", "application/json")
	return &Service{
		repo:                  repo,
		client:                client,
		flightServiceURL:      flightServiceURL,
		reservationServiceURL: reservationServiceURL,
	}
}

func (s *Service) bookTicket(ctx context.Context, reqBody BookTicketDTO, bookingUserID uuid.UUID) (BookTicketResponseDTO, error) {
	g, gctx := errgroup.WithContext(ctx)
	for _, segment := range reqBody.FlightSegments {
		g.Go(func() error {
			_, err := s.resolveFlight(gctx, segment.FlightID)
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return BookTicketResponseDTO{}, err
	}

	flightIDs := make([]uuid.UUID, 0, len(reqBody.FlightSegments))
	for _, seg := range reqBody.FlightSegments {
		flightIDUUID, err := uuid.Parse(seg.FlightID)
		if err != nil {
			return BookTicketResponseDTO{}, err
		}
		flightIDs = append(flightIDs, flightIDUUID)
	}

	bookingID, err := s.repo.bookTicketAndSave(ctx, reqBody, flightIDs, bookingUserID)
	if err != nil {
		if errors.Is(err, ErrInsufficientSeatsLeft) {
			return BookTicketResponseDTO{}, app_error.Conflict("insufficient seats left", err)
		}
		return BookTicketResponseDTO{}, err
	}
	return BookTicketResponseDTO{bookingID: bookingID.String()}, nil
}

func (s *Service) resolveFlight(ctx context.Context, flightID string) (FlightRecord, error) {
	v, err, _ := s.flightGroup.Do(flightID, func() (interface{}, error) {
		flightIDUUID, err := uuid.Parse(flightID)
		if err != nil {
			return nil, err
		}
		exists, err := s.repo.existsByFlightID(ctx, flightIDUUID)
		if err != nil {
			return nil, err
		}
		if exists {
			return FlightRecord{}, nil
		}
		resp, err := s.client.R().SetQueryParam("flight-id", flightID).Get(s.flightServiceURL + "/flight/validate")
		if err != nil {
			return FlightRecord{}, err
		}
		if resp.StatusCode() != http.StatusOK {
			return FlightRecord{}, app_error.UnprocessableEntity("flight ID not valid", errors.New("flight ID not valid"))
		}

		var flightRecord FlightRecord
		resp, err = s.client.R().SetBody(resp.Body()).SetResult(&flightRecord).Post(s.reservationServiceURL + "/aircraft/aircraft-details")
		if err != nil {
			return FlightRecord{}, err
		}
		if resp.StatusCode() != http.StatusOK {
			return FlightRecord{}, errors.New("error from reservation service: " + resp.Status())
		}
		err = s.repo.addSingleFlight(ctx, flightRecord)
		if err != nil {
			return FlightRecord{}, err
		}
		return flightRecord, nil
	})
	if err != nil {
		return FlightRecord{}, err
	}
	return v.(FlightRecord), nil
}
