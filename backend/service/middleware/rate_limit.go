package middleware

import (
	"fmt"
	"pouch-ai/backend/domain"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func NewRateLimitMiddleware(config map[string]string) domain.Middleware {
	var (
		limiters = make(map[int64]*rate.Limiter)
		mu       sync.Mutex

		limit, _ = strconv.Atoi(config["limit"])
		period   = config["period"]
	)

	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && period != "" && period != "none" && limit > 0 {
			mu.Lock()
			l, ok := limiters[int64(req.Key.ID)]
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
		return next.Handle(req)
	})
}
