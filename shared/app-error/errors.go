package app_error

import (
	"net/http"
)

func BadRequest(message string, err error) *AppError {
	return New(http.StatusBadRequest, message, err)
}

func Unauthorized(message string, err error) *AppError {
	return New(http.StatusUnauthorized, message, err)
}

func NotFound(message string, err error) *AppError {
	return New(http.StatusNotFound, message, err)
}

func Conflict(message string, err error) *AppError {
	return New(http.StatusConflict, message, err)
}

func InternalServer(err error) *AppError {
	return New(http.StatusInternalServerError, "Internal Server Error", err)
}
