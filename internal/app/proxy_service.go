package app

import (
	"fmt"
	"pouch-ai/internal/domain/proxy"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ProxyService struct {
	chain *proxy.Chain
}

func NewProxyService(finalHandler proxy.Handler, middlewares ...proxy.Middleware) *ProxyService {
	return &ProxyService{
		chain: proxy.NewChain(finalHandler, middlewares...),
	}
}

func (s *ProxyService) Execute(req *proxy.Request) (*proxy.Response, error) {
	return s.chain.Handle(req)
}

// Middlewares

func NewRateLimitMiddleware() proxy.Middleware {
	var (
		limiters = make(map[int64]*rate.Limiter)
		mu       sync.Mutex
	)

	return proxy.MiddlewareFunc(func(req *proxy.Request, next proxy.Handler) (*proxy.Response, error) {
		if req.Key != nil && req.Key.RateLimit.Period != "none" && req.Key.RateLimit.Limit > 0 {
			mu.Lock()
			l, ok := limiters[int64(req.Key.ID)]
			if !ok {
				var r rate.Limit
				switch req.Key.RateLimit.Period {
				case "second":
					r = rate.Limit(req.Key.RateLimit.Limit)
				case "minute":
					r = rate.Every(time.Minute / time.Duration(req.Key.RateLimit.Limit))
				default:
					r = rate.Inf
				}
				l = rate.NewLimiter(r, req.Key.RateLimit.Limit)
				limiters[int64(req.Key.ID)] = l
			}
			mu.Unlock()

			if !l.Allow() {
				return nil, fmt.Errorf("rate limit exceeded")
			}
		}
		return next.Handle(req)
	})
}

func NewUsageTrackingMiddleware(keyService *KeyService) proxy.Middleware {
	return proxy.MiddlewareFunc(func(req *proxy.Request, next proxy.Handler) (*proxy.Response, error) {
		resp, err := next.Handle(req)
		if err == nil && resp != nil && req.Key != nil && resp.TotalCost > 0 {
			_ = keyService.IncrementUsage(req.Context, req.Key.ID, resp.TotalCost)
		}
		return resp, err
	})
}

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
