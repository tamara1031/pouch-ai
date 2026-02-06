package api

import (
	"net/http"
	"pouch-ai/internal/service"
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(keyService *service.KeyService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization Header")
			}

			keyStr := strings.TrimPrefix(authHeader, "Bearer ")
			k, err := keyService.VerifyKey(c.Request().Context(), keyStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			c.Set("app_key", k)
			return next(c)
		}
	}
}
