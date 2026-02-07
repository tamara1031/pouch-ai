package middleware

import (
	"fmt"
	"pouch-ai/backend/domain"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func NewRateLimitMiddleware(_ map[string]string) domain.Middleware {
	var (
		limiters = make(map[int64]*rate.Limiter)
		mu       sync.Mutex
	)

	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && req.Key.Configuration != nil {
			var limit int
			var period string

			for _, m := range req.Key.Configuration.Middlewares {
				if m.ID == "rate_limit" {
					fmt.Sscanf(m.Config["limit"], "%d", &limit)
					period = m.Config["period"]
					break
				}
			}

			if period != "" && period != "none" && limit > 0 {
				mu.Lock()
				l, ok := limiters[int64(req.Key.ID)]
				// Note: If limit/period changes for the same key, we should ideally recreate the limiter.
				// For simplicity here, we stick to the first one or we can check if it changed.
				if !ok {
					var r rate.Limit
					switch period {
					case "second":
						r = rate.Limit(limit)
					case "minute":
						r = rate.Every(time.Minute / time.Duration(limit))
					default:
						r = rate.Inf
					}
					l = rate.NewLimiter(r, limit)
					limiters[int64(req.Key.ID)] = l
				}
				mu.Unlock()

				if !l.Allow() {
					return nil, fmt.Errorf("rate limit exceeded")
				}
			}
		}
		return next.Handle(req)
	})
}
