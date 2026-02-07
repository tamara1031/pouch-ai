package middlewares

import (
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"
)

func GetUsageTrackingInfo(keyService *service.KeyService) domain.MiddlewareInfo {
	return domain.MiddlewareInfo{
		ID:        "usage_tracking",
		Schema:    domain.MiddlewareSchema{},
		IsDefault: true,
	}
}

func NewUsageTrackingMiddleware(keyService *service.KeyService) func(map[string]any) domain.Middleware {
	return func(_ map[string]any) domain.Middleware {
		return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
			resp, err := next.Handle(req)
			if err == nil && resp != nil && req.Key != nil && resp.TotalCost > 0 {
				_ = keyService.IncrementUsage(req.Context, req.Key, resp.TotalCost)
			}
			return resp, err
		})
	}
}
