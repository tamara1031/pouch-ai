package server

import (
	"net/http"
	"pouch-ai/internal/auth"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// Per-key rate limiters cache
var (
	keyLimiters   = make(map[int64]*rate.Limiter)
	keyLimitersMu sync.RWMutex
)

// getOrCreateLimiter retrieves or creates a rate limiter for a key.
func getOrCreateLimiter(keyID int64, rateLimit int, ratePeriod string) *rate.Limiter {
	keyLimitersMu.RLock()
	limiter, exists := keyLimiters[keyID]
	keyLimitersMu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter
	var r rate.Limit
	if ratePeriod == "none" || rateLimit <= 0 {
		r = rate.Inf // Unlimited
	} else if ratePeriod == "second" {
		r = rate.Limit(rateLimit) // rateLimit per second
	} else {
		// Default: minute
		r = rate.Limit(float64(rateLimit) / 60.0) // rateLimit per minute
	}

	limiter = rate.NewLimiter(r, max(rateLimit, 1)) // Burst = rate limit

	keyLimitersMu.Lock()
	keyLimiters[keyID] = limiter
	keyLimitersMu.Unlock()

	return limiter
}

// AuthMiddleware creates a middleware that verifies API keys.
// If valid, sets "app_key_id" in context.
// If Mock key, returns Mock response immediately.
// Also enforces per-key rate limiting.
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

			// Per-key rate limiting (applies to all keys including mock)
			if info.RatePeriod != "none" && info.RateLimit > 0 {
				limiter := getOrCreateLimiter(info.ID, info.RateLimit, info.RatePeriod)
				if !limiter.Allow() {
					return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded for this API key")
				}
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
