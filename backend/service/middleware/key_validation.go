package middleware

import (
	"fmt"
	"pouch-ai/backend/domain"
)

func NewKeyValidationMiddleware(_ map[string]string) domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key == nil {
			return nil, fmt.Errorf("no application key provided")
		}

		if req.Key.IsExpired() {
			return nil, fmt.Errorf("key has expired")
		}

		return next.Handle(req)
	})
}
