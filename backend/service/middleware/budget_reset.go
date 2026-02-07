package middleware

import (
	"fmt"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"
)

func NewBudgetResetMiddleware(keyService *service.KeyService) func(map[string]string) domain.Middleware {
	return func(_ map[string]string) domain.Middleware {
		return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
			if req.Key != nil && req.Key.NeedsReset() {
				if err := keyService.ResetKeyUsage(req.Context, req.Key); err != nil {
					return nil, fmt.Errorf("failed to reset budget usage: %w", err)
				}
			}
			return next.Handle(req)
		})
	}
}
