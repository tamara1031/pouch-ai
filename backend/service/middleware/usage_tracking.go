package middleware

import (
	"pouch-ai/backend/service"
	"pouch-ai/backend/domain"
)

func NewUsageTrackingMiddleware(keyService *service.KeyService) domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		resp, err := next.Handle(req)
		if err == nil && resp != nil && req.Key != nil && resp.TotalCost > 0 {
			_ = keyService.IncrementUsage(req.Context, req.Key, resp.TotalCost)
		}
		return resp, err
	})
}
