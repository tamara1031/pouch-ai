package middlewares

import (
	"fmt"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"
	"strconv"
	"time"
)

func GetBudgetResetInfo(keyService *service.KeyService) domain.MiddlewareInfo {
	return domain.MiddlewareInfo{
		ID: "budget_reset",
		Schema: domain.MiddlewareSchema{
			"period": {Type: domain.FieldTypeNumber, DisplayName: "Reset Period (seconds)", Default: 2592000, Description: "Reset interval in seconds (default 30 days)", Role: domain.FieldRolePeriod},
		},
	}
}

func NewBudgetResetMiddleware(keyService *service.KeyService) func(map[string]any) domain.Middleware {
	return func(config map[string]any) domain.Middleware {
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
			if req.Key != nil && periodSeconds > 0 {
				now := time.Now()
				duration := time.Duration(periodSeconds * float64(time.Second))

				if now.After(req.Key.LastResetAt.Add(duration)) {
					if err := keyService.ResetKeyUsage(req.Context, req.Key); err != nil {
						return nil, fmt.Errorf("failed to reset budget usage: %w", err)
					}
				}
			}
			return next.Handle(req)
		})
	}
}
