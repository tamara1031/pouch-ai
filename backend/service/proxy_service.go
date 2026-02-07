package service

import (
	"fmt"
	"pouch-ai/backend/domain"
)

type ProxyService struct {
	finalHandler domain.Handler
	mwRegistry   domain.MiddlewareRegistry
}

func NewProxyService(finalHandler domain.Handler, mwRegistry domain.MiddlewareRegistry) *ProxyService {
	return &ProxyService{
		finalHandler: finalHandler,
		mwRegistry:   mwRegistry,
	}
}

func (s *ProxyService) Execute(req *domain.Request) (*domain.Response, error) {
	if req.Key == nil || req.Key.Configuration == nil {
		return s.finalHandler.Handle(req)
	}

	var mws []domain.Middleware
	for _, mwConfig := range req.Key.Configuration.Middlewares {
		mw, err := s.mwRegistry.Get(mwConfig.ID, mwConfig.Config)
		if err != nil {
			fmt.Printf("WARN: middleware %s not found or failed to initialize: %v\n", mwConfig.ID, err)
			continue
		}
		mws = append(mws, mw)
	}

	chain := domain.NewChain(s.finalHandler, mws...)
	return chain.Handle(req)
}
