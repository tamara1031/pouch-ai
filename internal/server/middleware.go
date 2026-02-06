package server

import (
	"net/http"
	"pouch-ai/internal/auth"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware creates a middleware that verifies API keys.
// If valid, sets "app_key_id" in context.
// If Mock key, returns Mock response immediately.
func AuthMiddleware(keyMgr *auth.KeyManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				// No key provided
				// Check if any keys exist in DB. If yes, require auth.
				keys, _ := keyMgr.ListKeys()
				if len(keys) > 0 {
					return echo.NewHTTPError(http.StatusUnauthorized, "Missing API Key")
				}
				// If no keys exist, allow anonymous (first setup or personal use without keys)
				return next(c)
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
			}
			key := strings.TrimPrefix(authHeader, "Bearer ")

			info, err := keyMgr.VerifyKey(key)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired API Key: "+err.Error())
			}
			if info == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API Key")
			}

			if info.IsMock {
				// Mock Mode
				return c.JSONBlob(http.StatusOK, []byte(info.MockConfig))
			}

			// Store Key ID in context for UsageCallback
			c.Set("app_key_id", info.ID)
			return next(c)
		}
	}
}
