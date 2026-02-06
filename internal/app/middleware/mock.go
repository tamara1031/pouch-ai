package middleware

import (
	"pouch-ai/internal/domain/proxy"
)

func NewMockMiddleware() proxy.Middleware {
	return proxy.MiddlewareFunc(func(req *proxy.Request, next proxy.Handler) (*proxy.Response, error) {
		if req.Key != nil && req.Key.IsMock {
			return &proxy.Response{
				StatusCode: 200,
				Body:       []byte(req.Key.MockConfig),
			}, nil
		}
		return next.Handle(req)
	})
}
