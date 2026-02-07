package middleware

import (
	"pouch-ai/internal/service"
	"pouch-ai/internal/domain"
)

func NewUsageTrackingMiddleware(keyService *service.KeyService) domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		resp, err := next.Handle(req)
		if err == nil && resp != nil && req.Key != nil && resp.TotalCost > 0 {
			// Increment usage in-memory (synchronously) to ensure strict budget enforcement
			// for subsequent requests. The database update happens asynchronously within the service.
			_ = keyService.IncrementUsage(req.Context, req.Key.ID, resp.TotalCost)
		}
		return resp, err
	})
}
