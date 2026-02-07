package middlewares

import (
	"fmt"
	"pouch-ai/backend/domain"
	"strconv"
)

func GetBudgetEnforcementInfo() domain.MiddlewareInfo {
	return domain.MiddlewareInfo{
		ID: "budget",
		Schema: domain.MiddlewareSchema{
			"limit": {Type: domain.FieldTypeNumber, DisplayName: "Budget Limit", Default: 5.00, Description: "Budget limit in USD", Role: domain.FieldRoleLimit},
		},
	}
}

func NewBudgetEnforcementMiddleware(config map[string]any) domain.Middleware {
	limit := 0.0
	if val, ok := config["limit"]; ok {
		switch v := val.(type) {
		case string:
			limit, _ = strconv.ParseFloat(v, 64)
		case float64:
			limit = v
		case int:
			limit = float64(v)
		}
	}

	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && limit > 0 {
			if req.Key.BudgetUsage >= limit {
				return nil, fmt.Errorf("budget limit exceeded (limit: $%.2f, used: $%.2f)", limit, req.Key.BudgetUsage)
			}
		}
		return next.Handle(req)
	})
}
