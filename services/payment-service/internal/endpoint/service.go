package endpoint

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/services/payment-service/internal/database"
	app_error "github.com/Aarav-S2005/flight-booking-microservices/shared/app-error"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo              *Repository
	bookingServiceURL string
	httpClient        *resty.Client
}

func NewService(db *pgxpool.Pool, bookingServiceURL string) *Service {
	return &Service{
		bookingServiceURL: bookingServiceURL,
		httpClient:        resty.New().SetBaseURL(bookingServiceURL).SetHeader("Content-Type", "application/json"),
		repo:              NewRepository(db),
	}
}

func (s *Service) Pay(ctx context.Context, userID uuid.UUID, reqBody MakePaymentDTO) error {
	bookingID, err := uuid.Parse(reqBody.BookingID)
	if err != nil {
		return app_error.BadRequest("could not parse bookingID", err)
	}

	paymentRecord, err := s.repo.findRecordByUserIDAndBookingID(ctx, userID, bookingID)

	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			var respBody ValidateBookingResponseDTO
			res, err := s.httpClient.R().SetResult(&respBody).SetBody(ValidateBookingRequestDTO{
				BookingID: bookingID.String(),
				UserID:    userID.String(),
			}).Post("/validate-booking")
			if err != nil {
				return err
			}
			if res.StatusCode() != http.StatusOK {
				return errors.New("booking does not exists")
			}
			paymentRecord = database.Payment{
				UserID:    userID,
				BookingID: bookingID,
				Amount:    reqBody.Amount,
			}
		} else {
			return err
		}
	}
	if paymentRecord.PaymentCompletedAt != nil {
		return app_error.BadRequest("payment already completed", errors.New("payment already completed"))
	}
	if paymentRecord.Amount != reqBody.Amount {
		return app_error.BadRequest("amount does not match", errors.New("amount does not match"))
	}
	now := time.Now().UTC()
	var resBody ValidatePaymentToBookingResponseDTO
	resp, err := s.httpClient.R().SetResult(&resBody).SetBody(ValidatePaymentToBookingRequestDTO{
		PaymentTime: now,
		UserID:      userID.String(),
		BookingID:   bookingID.String(),
	}).Post("/validate-payment")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return app_error.NotFound("booking not found", errors.New("booking not found"))
	} else if resp.StatusCode() == http.StatusBadRequest {
		return app_error.BadRequest("could not validate payment", errors.New("could not validate payment"))
	} else if resp.StatusCode() == http.StatusOK {
		if !resBody.Valid {
			return app_error.BadRequest("payment time over", errors.New("payment time over"))
		}
		err = s.repo.savePayment(ctx, now, userID, bookingID, paymentRecord.Amount)
		if err != nil {
			return err
		}
	} else {
		return app_error.InternalServer(errors.New("could not validate payment"))
	}
	return nil
}

func (s *Service) validatePayment(ctx context.Context, bookingID, userID uuid.UUID) (bool, error) {
	paymentRecord, err := s.repo.findRecordByUserIDAndBookingID(ctx, userID, bookingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, app_error.NotFound("payment not found", errors.New("payment not found"))
		}
		return false, err
	}
	if paymentRecord.PaymentCompletedAt == nil {
		return false, nil
	}
	return true, nil
}
