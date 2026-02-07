package middleware

import (
	"fmt"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"
	"time"
)

func NewBudgetResetMiddleware(keyService *service.KeyService) func(map[string]string) domain.Middleware {
	return func(config map[string]string) domain.Middleware {
		period := config["period"]

		return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
			if req.Key != nil && period != "" && period != "none" {
				now := time.Now()
				var duration time.Duration
				switch period {
				case "weekly":
					duration = 7 * 24 * time.Hour
				case "monthly":
					duration = 30 * 24 * time.Hour
				case "daily":
					duration = 24 * time.Hour
				}

				if duration > 0 && now.After(req.Key.LastResetAt.Add(duration)) {
					if err := keyService.ResetKeyUsage(req.Context, req.Key); err != nil {
						return nil, fmt.Errorf("failed to reset budget usage: %w", err)
					}
				}
			}
			return next.Handle(req)
		})
	}
}
