package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func NewAPIError(c echo.Context, status int, message string) error {
	return c.JSON(status, APIError{
		Error:   http.StatusText(status),
		Code:    status,
		Message: message,
	})
}

func BadRequest(c echo.Context, message string) error {
	return NewAPIError(c, http.StatusBadRequest, message)
}

func Unauthorized(c echo.Context, message string) error {
	return NewAPIError(c, http.StatusUnauthorized, message)
}

func InternalError(c echo.Context, message string) error {
	return NewAPIError(c, http.StatusInternalServerError, message)
}

func BadGateway(c echo.Context, message string) error {
	return NewAPIError(c, http.StatusBadGateway, message)
}
