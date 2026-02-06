package app

import (
	"pouch-ai/internal/domain/proxy"
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
