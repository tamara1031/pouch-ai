package middleware

import (
	"bytes"
	"io"
	"pouch-ai/internal/domain"
)

func NewMockMiddleware() domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && req.Key.IsMock {
			return &domain.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(req.Key.MockConfig)),
			}, nil
		}
		return next.Handle(req)
	})
}
