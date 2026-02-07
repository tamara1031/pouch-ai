package middlewares

import (
	"fmt"
	"pouch-ai/backend/domain"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func GetInfo() domain.PluginInfo {
	return domain.PluginInfo{
		ID: "rate_limit",
		Schema: domain.PluginSchema{
			"limit":  {Type: domain.FieldTypeNumber, DisplayName: "Request Limit", Default: 10, Description: "Requests per period", Role: domain.FieldRoleLimit},
			"period": {Type: domain.FieldTypeNumber, DisplayName: "Period (seconds)", Default: 60, Description: "Time window in seconds", Role: domain.FieldRolePeriod},
		},
		IsDefault: true,
	}
}

func NewRateLimitMiddleware(config map[string]any) domain.Middleware {
	var (
		limiters = make(map[int64]*rate.Limiter)
		mu       sync.Mutex
	)

	limit := 0
	if l, ok := config["limit"]; ok {
		switch v := l.(type) {
		case string:
			limit, _ = strconv.Atoi(v)
		case float64:
			limit = int(v)
		case int:
			limit = v
		}
	}

	periodSeconds := 0.0
	if p, ok := config["period"]; ok {
		switch v := p.(type) {
		case string:
			periodSeconds, _ = strconv.ParseFloat(v, 64)
		case float64:
			periodSeconds = v
		case int:
			periodSeconds = float64(v)
		}
	}

	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && periodSeconds > 0 && limit > 0 {
			mu.Lock()
			l, ok := limiters[int64(req.Key.ID)]
			if !ok {
				r := rate.Every(time.Duration(periodSeconds*float64(time.Second)) / time.Duration(limit))
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
