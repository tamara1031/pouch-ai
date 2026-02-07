package middleware

import (
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"
)

func NewUsageTrackingMiddleware(keyService *service.KeyService) func(map[string]string) domain.Middleware {
	return func(_ map[string]string) domain.Middleware {
		return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
			resp, err := next.Handle(req)
			if err == nil && resp != nil && req.Key != nil && resp.TotalCost > 0 {
				_ = keyService.IncrementUsage(req.Context, req.Key, resp.TotalCost)
			}
			return resp, err
		})
	}
}
