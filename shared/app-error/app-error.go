package app_error

import (
	"errors"
	"net/http"

	"github.com/Aarav-S2005/flight-booking-microservices/shared/utility"
)

type AppError struct {
	Err        error
	StatusCode int
	Message    string
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func HandleError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = InternalServer(err)
	}
	_ = utility.ConvertStructToJSON(w, appErr.StatusCode, Response{Success: false, Message: appErr.Message + ": " + err.Error()})
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string, err error) *AppError {
	return &AppError{
		StatusCode: code,
		Message:    message,
		Err:        err,
	}
}
