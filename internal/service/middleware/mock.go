package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"pouch-ai/internal/domain"
	"strings"
)

func NewMockMiddleware() domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		if req.Key != nil && req.Key.IsMock {
			if !json.Valid([]byte(req.Key.MockConfig)) {
				header := make(http.Header)
				header.Set("Content-Type", "application/json")
				header.Set("X-Content-Type-Options", "nosniff")
				return &domain.Response{
					StatusCode: 500,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader(`{"error": "Invalid mock configuration"}`)),
				}, nil
			}

			header := make(http.Header)
			header.Set("Content-Type", "application/json")
			header.Set("X-Content-Type-Options", "nosniff")
			return &domain.Response{
				StatusCode: 200,
				Header:     header,
				Body:       io.NopCloser(bytes.NewBufferString(req.Key.MockConfig)),
			}, nil
		}
		return next.Handle(req)
	})
}
