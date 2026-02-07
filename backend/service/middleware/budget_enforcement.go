package middleware

import (
	"fmt"
	"pouch-ai/backend/domain"
	"strconv"
)

func NewBudgetEnforcementMiddleware(config map[string]string) domain.Middleware {
	limit, _ := strconv.ParseFloat(config["limit"], 64)

	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && limit > 0 {
			if req.Key.BudgetUsage >= limit {
				return nil, fmt.Errorf("budget limit exceeded (limit: $%.2f, used: $%.2f)", limit, req.Key.BudgetUsage)
			}
		}
		return next.Handle(req)
	})
}
