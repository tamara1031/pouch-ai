package middleware

import (
	"context"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/service"
)

func NewUsageTrackingMiddleware(keyService *service.KeyService) domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		resp, err := next.Handle(req)
		if err == nil && resp != nil && req.Key != nil && resp.TotalCost > 0 {
			// Run usage increment in background to avoid blocking the response
			go func(ctx context.Context, keyID domain.ID, cost float64) {
				_ = keyService.IncrementUsage(ctx, keyID, cost)
			}(context.WithoutCancel(req.Context), req.Key.ID, resp.TotalCost)
		}
		return resp, err
	})
}
