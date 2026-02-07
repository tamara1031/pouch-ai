package middleware

import (
	"fmt"
	"pouch-ai/backend/domain"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func NewRateLimitMiddleware() domain.Middleware {
	var (
		limiters = make(map[int64]*rate.Limiter)
		mu       sync.Mutex
	)

	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && req.Key.RateLimit.Period != "none" && req.Key.RateLimit.Limit > 0 {
			mu.Lock()
			l, ok := limiters[int64(req.Key.ID)]
			if !ok {
				var r rate.Limit
				switch req.Key.RateLimit.Period {
				case "second":
					r = rate.Limit(req.Key.RateLimit.Limit)
				case "minute":
					r = rate.Every(time.Minute / time.Duration(req.Key.RateLimit.Limit))
				default:
					r = rate.Inf
				}
				l = rate.NewLimiter(r, req.Key.RateLimit.Limit)
				limiters[int64(req.Key.ID)] = l
			}
			mu.Unlock()

			if !l.Allow() {
				return nil, fmt.Errorf("rate limit exceeded")
			}
		}
		return next.Handle(req)
	})
}
