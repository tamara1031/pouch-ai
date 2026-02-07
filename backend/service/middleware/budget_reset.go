package middleware

import (
	"fmt"
	"pouch-ai/internal/service"
	"pouch-ai/internal/domain"
)

func NewBudgetResetMiddleware(keyService *service.KeyService) domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && req.Key.NeedsReset() {
			if err := keyService.ResetKeyUsage(req.Context, req.Key); err != nil {
				return nil, fmt.Errorf("failed to reset budget usage: %w", err)
			}
		}
		return next.Handle(req)
	})
}
