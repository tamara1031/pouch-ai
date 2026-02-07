package service

import (
	"fmt"
	"pouch-ai/backend/domain"
	"time"
)

// ProxyService orchestrates the request processing flow.
// It handles key validation, budget enforcement, middleware execution, and usage tracking.
type ProxyService struct {
	finalHandler domain.Handler
	mwRegistry   domain.MiddlewareRegistry
	keyService   *KeyService
}

// NewProxyService creates a new instance of ProxyService.
func NewProxyService(finalHandler domain.Handler, mwRegistry domain.MiddlewareRegistry, keyService *KeyService) *ProxyService {
	return &ProxyService{
		finalHandler: finalHandler,
		mwRegistry:   mwRegistry,
		keyService:   keyService,
	}
}

// Execute processes the incoming request through the following steps:
// 1. Key Validation & Auto-Renewal
// 2. Budget Reset (if period passed)
// 3. Budget Limit Enforcement
// 4. Dynamic Middleware Chain Execution (including the final handler)
// 5. Usage Tracking (post-execution)
func (s *ProxyService) Execute(req *domain.Request) (*domain.Response, error) {
	if req.Key == nil {
		return nil, fmt.Errorf("no application key provided")
	}

	// 1. Core Validation
	if req.Key.IsExpired() {
		if req.Key.AutoRenew {
			if err := s.keyService.RenewKey(req.Context, req.Key); err != nil {
				fmt.Printf("WARN: failed to auto-renew key: %v\n", err)
				return nil, fmt.Errorf("key has expired and auto-renew failed")
			}
			fmt.Printf("INFO: key %s auto-renewed\n", req.Key.Prefix)
		} else {
			return nil, fmt.Errorf("key has expired")
		}
	}

	config := req.Key.Configuration
	if config == nil {
		return s.finalHandler.Handle(req)
	}

	// 2. Core Budget Reset Logic
	if config.ResetPeriod > 0 {
		now := time.Now()
		duration := time.Duration(config.ResetPeriod) * time.Second
		if now.After(req.Key.LastResetAt.Add(duration)) {
			if err := s.keyService.ResetKeyUsage(req.Context, req.Key); err != nil {
				fmt.Printf("WARN: failed to reset budget usage: %v\n", err)
			}
		}
	}

	// 3. Core Budget Enforcement
	if config.BudgetLimit > 0 {
		if req.Key.BudgetUsage >= config.BudgetLimit {
			return nil, fmt.Errorf("budget limit exceeded (limit: $%.2f, used: $%.2f)", config.BudgetLimit, req.Key.BudgetUsage)
		}
	}

	// 4. Plugin Middlewares
	var mws []domain.Middleware
	for _, mwConfig := range config.Middlewares {
		mw, err := s.mwRegistry.Get(mwConfig.ID, mwConfig.Config)
		if err != nil {
			fmt.Printf("WARN: middleware %s not found or failed to initialize: %v\n", mwConfig.ID, err)
			continue
		}
		mws = append(mws, mw)
	}

	chain := domain.NewChain(s.finalHandler, mws...)
	resp, err := chain.Handle(req)

	// 5. Core Usage Tracking
	if err == nil && resp != nil && resp.TotalCost > 0 {
		_ = s.keyService.IncrementUsage(req.Context, req.Key, resp.TotalCost)
	}

	return resp, err
}
