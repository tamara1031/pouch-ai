package middleware

import (
	"fmt"
	"pouch-ai/internal/domain/proxy"
)

func NewKeyValidationMiddleware() proxy.Middleware {
	return proxy.MiddlewareFunc(func(req *proxy.Request, next proxy.Handler) (*proxy.Response, error) {
		if req.Key == nil {
			return nil, fmt.Errorf("no application key provided")
		}

		if req.Key.IsExpired() {
			return nil, fmt.Errorf("key has expired")
		}

		// Budget check only applies to non-mock requests
		if !req.Key.IsMock && req.Key.IsBudgetExceeded() {
			return nil, fmt.Errorf("budget limit exceeded")
		}

		return next.Handle(req)
	})
}
