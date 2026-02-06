package middleware

import (
	"bytes"
	"io"
	"net/http"
	"pouch-ai/internal/domain"
)

func NewMockMiddleware() domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && req.Key.IsMock {
			header := make(http.Header)
			header.Set("Content-Type", "application/json")
			return &domain.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(bytes.NewBufferString(req.Key.MockConfig)),
			}, nil
		}
		return next.Handle(req)
	})
}
