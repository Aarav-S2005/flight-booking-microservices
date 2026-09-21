package booking

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/booking-service/internal/database"
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
	paymentServiceURL     string
	rdb                   *redis.Client
	publisher             *rabbitmq.Publisher
	flightGroup           singleflight.Group
}

func NewService(repo *Repository, flightServiceURL, reservationServiceURL, paymentServiceURL string, rdb *redis.Client, publisher *rabbitmq.Publisher) *Service {
	client := resty.New().SetHeader("Content-Type", "application/json")
	return &Service{
		repo:                  repo,
		client:                client,
		flightServiceURL:      flightServiceURL,
		reservationServiceURL: reservationServiceURL,
		rdb:                   rdb,
		publisher:             publisher,
		flightGroup:           singleflight.Group{},
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

	bookingID, updatedSeats, passengerIDs, err := s.repo.bookTicketAndSave(ctx, reqBody, flightIDs, bookingUserID)
	if err != nil {
		if errors.Is(err, ErrInsufficientSeatsLeft) {
			return BookTicketResponseDTO{}, app_error.Conflict("insufficient seats left", err)
		}
		return BookTicketResponseDTO{}, err
	}

	for _, updatedSeat := range updatedSeats {
		err = s.publisher.Publish(ctx, contract.RoutingFlightSeatUpdated, contract.FlightSeatUpdatedEvent{
			FlightID: updatedSeat.FlightID.String(),
			NewSeat:  updatedSeat.SeatsLeft,
			Version:  updatedSeat.Version,
		})
	}
	err = s.publisher.Publish(ctx, contract.RoutingBookingConfirmedPayment, contract.BookingConfirmedForPaymentEvent{
		UserID:    bookingUserID.String(),
		BookingID: bookingID.String(),
		TotalFare: reqBody.TotalFare,
	})
	value, err := json.Marshal(contract.BookingConfirmedForReservationEvent{
		BookingID:      bookingID.String(),
		UserID:         bookingUserID.String(),
		PassengerIDs:   UUIDsToStrings(passengerIDs),
		FlightSegments: stringFlightIDs,
	})
	if err != nil {
		log.Println(err)
	} else {
		err = s.rdb.Set(ctx, bookingID.String()+"-for-res-noti", value, 0).Err()
		if err != nil {
			log.Println(err)
		}
	}
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

func (s *Service) validateBooking(ctx context.Context, userID, bookingID uuid.UUID) (int, error) {
	totalFare, err := s.repo.getTotalFareByBookingIDAndUserID(ctx, bookingID, userID)
	if err != nil {
		if errors.Is(err, ErrBookingNotFound) {
			return 0, app_error.NotFound("booking not found", err)
		}
		return 0, err
	}

	return totalFare, nil
}

func (s *Service) validatePayment(ctx context.Context, reqBody ValidateBookingForPaymentRequestDTO) (bool, error) {
	userID, err := uuid.Parse(reqBody.UserID)
	if err != nil {
		return false, app_error.BadRequest("invalid user id", err)
	}
	bookingID, err := uuid.Parse(reqBody.BookingID)
	if err != nil {
		return false, app_error.BadRequest("invalid booking id", err)
	}
	creationTime, err := s.repo.getBookingCreationTime(ctx, userID, bookingID)
	if err != nil {
		if errors.Is(err, ErrBookingNotFound) {
			return false, app_error.NotFound("booking not found", err)
		}
		return false, err
	}
	if reqBody.PaymentTime.Before(creationTime) {
		return false, app_error.BadRequest("payment time cannot be before booking creation time", nil)
	}

	deadline := creationTime.Add(10 * time.Minute)
	if reqBody.PaymentTime.After(deadline) {
		seatUpdates, err := s.repo.increaseFlightSeatsByBookingID(ctx, bookingID)
		if err != nil {
			log.Fatal("could not update flight seats by booking id", err)
			return false, nil
		}
		for _, update := range seatUpdates {
			_ = s.publisher.Publish(ctx, contract.RoutingFlightSeatUpdated, contract.FlightSeatUpdatedEvent{
				FlightID: update.FlightID.String(),
				NewSeat:  update.SeatsLeft,
				Version:  update.Version,
			})
		}
		err = s.repo.updateStatusByBookingID(ctx, bookingID, "FAILED")
		if err != nil {
			if errors.Is(err, ErrBookingNotFound) {
				return false, app_error.NotFound("booking not found", err)
			}
			return false, err
		}
		return false, nil
	}
	err = s.repo.updateStatusByBookingID(ctx, bookingID, "CONFIRMED")
	if err != nil {
		log.Println("DB failed to update status ", err)
	}
	data, err := s.rdb.Get(ctx, bookingID.String()+"-for-res-noti").Bytes()
	if err != nil {
		log.Println("redis failed to get booking data: ", err)
	} else {
		var bookingMessageForReservation contract.BookingConfirmedForReservationEvent
		err = json.Unmarshal(data, &bookingMessageForReservation)
		if err != nil {
			log.Println("unmarshaling failed ", err)
		}
		err = s.publisher.Publish(ctx, contract.RoutingBookingConfirmedReservation, bookingMessageForReservation)
		if err != nil {
			log.Println("publishing failed ", err)
		}
	}
	return true, nil
}

func (s *Service) validateBookingForReservation(ctx context.Context, userID, bookingID uuid.UUID) (ValidateBookingForReservationResponseDTO, error) {
	status, err := s.repo.checkStatusByBookingID(ctx, bookingID, userID)
	if err != nil {
		if errors.Is(err, ErrBookingNotFound) {
			return ValidateBookingForReservationResponseDTO{}, app_error.NotFound("booking not found", err)
		}
		return ValidateBookingForReservationResponseDTO{}, err
	}
	if status == string(database.BookingFailed) {
		return ValidateBookingForReservationResponseDTO{}, app_error.BadRequest("booking failed, no reservation", errors.New("booking failed"))
	}
	if status != string(database.BookingConfirmed) {

		req := ValidatePaymentRequestDTO{
			BookingID: bookingID.String(),
			UserID:    userID.String(),
		}

		resp, err := s.client.R().SetBody(req).Get(s.paymentServiceURL + "/validate-payment")
		if err != nil {
			return ValidateBookingForReservationResponseDTO{}, err
		}
		if resp.StatusCode() >= 400 {
			return ValidateBookingForReservationResponseDTO{}, app_error.BadRequest("booking failed, no reservation", errors.New("booking failed"))
		}
		if resp.StatusCode() == http.StatusAccepted {
			return ValidateBookingForReservationResponseDTO{
				Status: "PAYMENT_PENDING",
			}, nil
		}
		err = s.repo.updateStatusByBookingID(ctx, bookingID, "CONFIRMED")
		if err != nil {
			log.Println("DB failed to update status ", err)
		}
	}

	data, err := s.rdb.Get(ctx, bookingID.String()+"-for-res-noti").Bytes()
	var respBody contract.BookingConfirmedForReservationEvent
	var strPassengerIDs []string
	var strFlightIDs []string
	if err != nil || json.Unmarshal(data, &respBody) != nil {
		log.Println("redis failed to get booking data: ", err)
		passengerIDs, err := s.repo.getPassengersIDbyBookingID(ctx, bookingID)
		if err != nil {
			if errors.Is(err, ErrBookingNotFound) {
				return ValidateBookingForReservationResponseDTO{}, app_error.NotFound("booking not found", err)
			}
			return ValidateBookingForReservationResponseDTO{}, err
		}
		flightIDs, err := s.repo.getFlightIDsBytBookingID(ctx, bookingID)
		if err != nil {
			if errors.Is(err, ErrBookingNotFound) {
				return ValidateBookingForReservationResponseDTO{}, app_error.NotFound("booking not found", err)
			}
			return ValidateBookingForReservationResponseDTO{}, err
		}
		strPassengerIDs = UUIDsToStrings(passengerIDs)
		strFlightIDs = UUIDsToStrings(flightIDs)
	} else {
		strPassengerIDs = respBody.PassengerIDs
		strFlightIDs = respBody.FlightSegments
	}
	return ValidateBookingForReservationResponseDTO{strPassengerIDs, strFlightIDs, "CONFIRMED"}, nil
}

// Helper

func (s *Service) resolveFlight(ctx context.Context, flightID string) (FlightRecord, error) {
	v, err, _ := s.flightGroup.Do(flightID, func() (interface{}, error) {
		flightIDUUID, err := uuid.Parse(flightID)
		if err != nil {
			return nil, err
		}
		record, err := s.repo.existsByFlightID(ctx, flightIDUUID)
		if errors.Is(err, ErrFlightNotFound) {
			var respBody GetFlightResponseFromFlightService
			resp, err := s.client.R().SetResult(&respBody).Get(s.flightServiceURL + "/flight/validate/" + flightID)
			if err != nil {
				return FlightRecord{}, err
			}
			if resp.StatusCode() != http.StatusOK {
				return FlightRecord{}, app_error.UnprocessableEntity("flight ID not valid", errors.New("flight ID not valid"))
			}
			err = s.repo.addSingleFlight(ctx, FlightRecord{
				FlightID:     flightIDUUID,
				TotalSeats:   respBody.SeatsLeft,
				AircraftType: respBody.AircraftType,
			})
			return FlightRecord{
				FlightID:     flightIDUUID,
				TotalSeats:   respBody.SeatsLeft,
				AircraftType: respBody.AircraftType,
			}, err
		}
		if err != nil {
			return nil, err
		}

		return record, nil
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
