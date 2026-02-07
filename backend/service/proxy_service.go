package service

import (
	"fmt"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/util/logger"
	"time"
)

type ProxyService struct {
	finalHandler domain.Handler
	mwRegistry   domain.MiddlewareRegistry
	keyService   *KeyService
}

func NewProxyService(finalHandler domain.Handler, mwRegistry domain.MiddlewareRegistry, keyService *KeyService) *ProxyService {
	return &ProxyService{
		finalHandler: finalHandler,
		mwRegistry:   mwRegistry,
		keyService:   keyService,
	}
}

func (s *ProxyService) Execute(req *domain.Request) (*domain.Response, error) {
	if req.Key == nil {
		return nil, fmt.Errorf("no application key provided")
	}

	// 1. Validation & Auto-renewal
	if err := s.validateKey(req); err != nil {
		return nil, err
	}

	// 2. Budget Management (Reset & Reservation)
	if err := s.manageBudget(req); err != nil {
		return nil, err
	}

	// 3. Build & Execute Middleware Chain
	chain := s.buildChain(req.Key.Configuration)
	return chain.Handle(req)
}

func (s *ProxyService) validateKey(req *domain.Request) error {
	if !req.Key.IsExpired() {
		return nil
	}

	if !req.Key.AutoRenew {
		return domain.ErrKeyExpired
	}

	if err := s.keyService.RenewKey(req.Context, req.Key); err != nil {
		logger.L.Warn("failed to auto-renew key", "prefix", req.Key.Prefix, "error", err)
		return fmt.Errorf("key has expired and auto-renew failed: %w", err)
	}

	logger.L.Info("key auto-renewed", "prefix", req.Key.Prefix)
	return nil
}

func (s *ProxyService) manageBudget(req *domain.Request) error {
	config := req.Key.Configuration
	if config == nil {
		return nil
	}

	// 1. Budget Reset Logic
	if config.ResetPeriod > 0 {
		now := time.Now()
		duration := time.Duration(config.ResetPeriod) * time.Second
		if now.After(req.Key.LastResetAt.Add(duration)) {
			if err := s.keyService.ResetKeyUsage(req.Context, req.Key); err != nil {
				logger.L.Warn("failed to reset budget usage", "prefix", req.Key.Prefix, "error", err)
			} else {
				logger.L.Info("budget reset", "prefix", req.Key.Prefix)
			}
		}
	}

	// 2. Budget Enforcement (Atomic Reservation)
	estimatedUsage, _ := req.Provider.EstimateUsage(req.Model, req.RawBody)
	reservedCost := 0.0
	if estimatedUsage != nil {
		reservedCost = estimatedUsage.TotalCost
	}

	if err := s.keyService.ReserveUsage(req.Context, req.Key.ID, reservedCost); err != nil {
		return err
	}

	req.ReservedCost = reservedCost
	req.Committer = s.keyService
	return nil
}

func (s *ProxyService) buildChain(config *domain.KeyConfiguration) domain.Handler {
	if config == nil || len(config.Middlewares) == 0 {
		return s.finalHandler
	}

	var mws []domain.Middleware
	for _, mwConfig := range config.Middlewares {
		mw, err := s.mwRegistry.Get(mwConfig.ID, mwConfig.Config)
		if err != nil {
			logger.L.Warn("middleware not found or failed to initialize", "id", mwConfig.ID, "error", err)
			continue
		}
		mws = append(mws, mw)
	}

	return domain.NewChain(s.finalHandler, mws...)
}
