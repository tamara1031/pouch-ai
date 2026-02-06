package middleware

import (
	"fmt"
	"pouch-ai/internal/app"
	"pouch-ai/internal/domain/proxy"
)

func NewBudgetResetMiddleware(keyService *app.KeyService) proxy.Middleware {
	return proxy.MiddlewareFunc(func(req *proxy.Request, next proxy.Handler) (*proxy.Response, error) {
		if req.Key != nil && req.Key.NeedsReset() {
			if err := keyService.ResetKeyUsage(req.Context, req.Key); err != nil {
				return nil, fmt.Errorf("failed to reset budget usage: %w", err)
			}
		}
		return next.Handle(req)
	})
}
