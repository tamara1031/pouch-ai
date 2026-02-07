package service

import (
	"pouch-ai/backend/domain"
)

type ProxyService struct {
	chain *domain.Chain
}

func NewProxyService(finalHandler domain.Handler, middlewares ...domain.Middleware) *ProxyService {
	return &ProxyService{
		chain: domain.NewChain(finalHandler, middlewares...),
	}
}

func (s *ProxyService) Execute(req *domain.Request) (*domain.Response, error) {
	return s.chain.Handle(req)
}
