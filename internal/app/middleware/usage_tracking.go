package middleware

import (
	"pouch-ai/internal/app"
	"pouch-ai/internal/domain/proxy"
)

func NewUsageTrackingMiddleware(keyService *app.KeyService) proxy.Middleware {
	return proxy.MiddlewareFunc(func(req *proxy.Request, next proxy.Handler) (*proxy.Response, error) {
		resp, err := next.Handle(req)
		if err == nil && resp != nil && req.Key != nil && resp.TotalCost > 0 {
			_ = keyService.IncrementUsage(req.Context, req.Key.ID, resp.TotalCost)
		}
		return resp, err
	})
}
