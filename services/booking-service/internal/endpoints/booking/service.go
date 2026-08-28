package booking

import (
	"context"
	"errors"
	"net/http"
	"sync"

	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq"
	"github.com/Aarav-S2005/flight-booking-microservices/shared/rabbitmq/contract"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

type Service struct {
	repo                  *Repository
	client                *resty.Client
	flightServiceURL      string
	reservationServiceURL string
	rdb                   *redis.Client
	publisher             *rabbitmq.Publisher
}

func NewService(repo *Repository, flightServiceURL, reservationServiceURL string, rdb *redis.Client, publisher *rabbitmq.Publisher) *Service {
	client := resty.New().
		SetHeader("Content-Type", "application/json")
	return &Service{
		repo:                  repo,
		client:                client,
		flightServiceURL:      flightServiceURL,
		reservationServiceURL: reservationServiceURL,
		rdb:                   rdb,
		publisher:             publisher,
	}
}

func (s *Service) bookTicket(ctx context.Context, reqBody BookTicketDTO, bookingUserID uuid.UUID) (BookTicketResponseDTO, error) {
	stringFlightIDs := make([]string, 0, len(reqBody.FlightSegments))
	for _, i := range reqBody.FlightSegments {
		stringFlightIDs = append(stringFlightIDs, i.FlightID)
	}
	validationResp, err := s.client.R().SetBody(ValidateFareRequestDTO{TotalFare: reqBody.TotalFare, FlightIDs: stringFlightIDs}).Post(s.flightServiceURL + "/flight/validate-fare")
	if err != nil {
		return BookTicketResponseDTO{}, err
	}
	if validationResp.StatusCode() != http.StatusOK {
		return BookTicketResponseDTO{}, app_error.Conflict("total fare does not match requested fare", errors.New("total fare does not match requested fare"))
	}

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

	bookingID, updatedSeats, err := s.repo.bookTicketAndSave(ctx, reqBody, flightIDs, bookingUserID)
	if err != nil {
		if errors.Is(err, ErrInsufficientSeatsLeft) {
			return BookTicketResponseDTO{}, app_error.Conflict("insufficient seats left", err)
		}
		return BookTicketResponseDTO{}, err
	}

	for _, updatedSeat := range updatedSeats {
		err = s.publisher.Publish(ctx, contract.FlightSeatUpdateRoutingKey, contract.SeatUpdatedEvent{
			FlightID: updatedSeat.FlightID.String(),
			NewSeat:  updatedSeat.SeatsLeft,
			Version:  updatedSeat.Version,
		})
	}
	err = s.publisher.Publish(ctx, contract.CreateReservationEventsRoutingKey, contract.CreateReservationEvent{
		BookingID:      bookingID.String(),
		PassengerCount: len(reqBody.PassengerDetails),
	})
	return BookTicketResponseDTO{BookingID: bookingID.String()}, nil
}

func (s *Service) getAllBookings(ctx context.Context, bookingUserID uuid.UUID) (GetAllBookingsDTO, error) {
	flightIDsGroupedByBookingID, err := s.repo.getAllFlightByBookingUserIDGroupedByBookingID(ctx, bookingUserID)
	if err != nil {
		if errors.Is(err, ErrBookingNotFound) {
			return GetAllBookingsDTO{}, app_error.NotFound("booking not found", err)
		}
		return GetAllBookingsDTO{}, err
	}
	uniqueFlightIDs := computeUniqueFlightIDs(flightIDsGroupedByBookingID)
	flightDetailsGroupedByFlightID, err := s.fetchFlightDetails(ctx, uniqueFlightIDs)
	if err != nil {
		return GetAllBookingsDTO{}, err
	}

	// save in redis to avoid repeated calls

	passengerDetails, err := s.repo.getAllPassengerByBookingUserIDGroupedByBookingID(ctx, bookingUserID)
	if err != nil {
		return GetAllBookingsDTO{}, err
	}

	totalFaresGroupedByBookingID, err := s.repo.getAllTotalFaresByBookingUserIDGroupedByBookingID(ctx, bookingUserID)
	if err != nil {
		return GetAllBookingsDTO{}, err
	}

	bookings := make([]GetBookingDTO, 0, len(flightIDsGroupedByBookingID))
	for bookingID, flightIDs := range flightIDsGroupedByBookingID {
		tempFlightDetailsArray := make([]FlightDetailsDTO, 0, len(flightIDs))
		for _, flightID := range flightIDs {
			tempFlightDetailsArray = append(tempFlightDetailsArray, flightDetailsGroupedByFlightID[flightID])
		}
		bookings = append(bookings, GetBookingDTO{
			BookingID:        bookingID.String(),
			PassengerDetails: passengerDetails[bookingID],
			FlightDetails:    tempFlightDetailsArray,
			TotalFare:        totalFaresGroupedByBookingID[bookingID],
		})
	}

	return GetAllBookingsDTO{
		Bookings: bookings,
	}, nil
}

// Helper

func (s *Service) resolveFlight(ctx context.Context, flightID string) (FlightRecord, error) {
	flightGroup := singleflight.Group{}
	v, err, _ := flightGroup.Do(flightID, func() (interface{}, error) {
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

func (s *Service) fetchFlightDetails(ctx context.Context, flightIDs []uuid.UUID) (map[uuid.UUID]FlightDetailsDTO, error) {
	g, gctx := errgroup.WithContext(ctx)
	flightDetails := make(map[uuid.UUID]FlightDetailsDTO)
	var mu sync.Mutex
	for i, flightID := range flightIDs {
		g.Go(func() error {
			var flight FlightDetailsDTO
			resp, err := s.client.R().SetContext(gctx).SetResult(&flight).Get(s.flightServiceURL + "/flight/" + flightID.String())
			if err != nil {
				return err
			}
			if resp.StatusCode() == http.StatusNotFound {
				return app_error.NotFound("flight not found", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return app_error.NotFound("internal server error", err)
			}
			mu.Lock()
			defer mu.Unlock()
			flight.SegmentOrder = i + 1
			flightDetails[flightID] = flight
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return flightDetails, nil
}
