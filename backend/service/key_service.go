package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"pouch-ai/backend/domain"
	"sync"
	"time"
)

type cachedKey struct {
	key       *domain.Key
	expiresAt time.Time
}

// KeyService manages the lifecycle, validation, and usage tracking of API keys.
// It implements a write-through cache strategy for high-performance key verification.
type KeyService struct {
	repo       domain.Repository
	registry   domain.ProviderRegistry
	mwRegistry domain.MiddlewareRegistry
	cache      map[string]cachedKey
	cacheMu    sync.RWMutex
}

// NewKeyService creates a new instance of KeyService.
func NewKeyService(repo domain.Repository, registry domain.ProviderRegistry, mwRegistry domain.MiddlewareRegistry) *KeyService {
	return &KeyService{
		repo:       repo,
		registry:   registry,
		mwRegistry: mwRegistry,
		cache:      make(map[string]cachedKey),
	}
}

// CreateKeyInput defines the parameters for creating a new API key.
type CreateKeyInput struct {
	Name        string
	Provider    domain.PluginConfig
	ExpiresAt   *int64
	Middlewares []domain.PluginConfig
	BudgetLimit float64
	ResetPeriod int
	AutoRenew   bool
}

// CreateKey generates a new API key with the specified configuration and persists it.
// It returns the raw key string (which should be shown to the user once) and the Key entity.
func (s *KeyService) CreateKey(ctx context.Context, input CreateKeyInput) (string, *domain.Key, error) {
	if input.Provider.ID != "" {
		if _, err := s.registry.Get(input.Provider.ID); err != nil {
			return "", nil, err
		}
	}

	rawKey, err := s.generateRandomKey()
	if err != nil {
		return "", nil, err
	}

	hash := s.hashKey(rawKey)
	prefix := rawKey[:8]

	k := &domain.Key{
		ID:        0,
		Name:      input.Name,
		KeyHash:   hash,
		Prefix:    prefix,
		AutoRenew: input.AutoRenew,
		Configuration: &domain.KeyConfiguration{
			Provider:    input.Provider,
			Middlewares: input.Middlewares,
			BudgetLimit: input.BudgetLimit,
			ResetPeriod: input.ResetPeriod,
		},
		BudgetUsage: 0,
		LastResetAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	if input.ExpiresAt != nil {
		t := time.Unix(*input.ExpiresAt, 0)
		k.ExpiresAt = &t
	}

	if err := k.Validate(); err != nil {
		return "", nil, err
	}

	if err := s.repo.Save(ctx, k); err != nil {
		return "", nil, err
	}

	return rawKey, k, nil
}

// VerifyKey authenticates a raw API key string.
// It checks the local cache first, then the database. If successful, it caches the key.
func (s *KeyService) VerifyKey(ctx context.Context, rawKey string) (*domain.Key, error) {
	hash := s.hashKey(rawKey)

	s.cacheMu.RLock()
	if entry, ok := s.cache[hash]; ok {
		if time.Now().Before(entry.expiresAt) {
			copied := s.copyKey(entry.key)
			s.cacheMu.RUnlock()
			return copied, nil
		}
	}
	s.cacheMu.RUnlock()

	k, err := s.repo.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, fmt.Errorf("invalid API key")
	}

	s.cacheMu.Lock()
	s.cache[hash] = cachedKey{
		key:       s.copyKey(k),
		expiresAt: time.Now().Add(1 * time.Minute),
	}
	s.cacheMu.Unlock()

	return k, nil
}

// ListKeys retrieves all API keys from the repository.
func (s *KeyService) ListKeys(ctx context.Context) ([]*domain.Key, error) {
	return s.repo.List(ctx)
}

// UpdateKeyInput defines the parameters for updating an existing API key.
type UpdateKeyInput struct {
	ID          int64
	Name        string
	Provider    domain.PluginConfig
	ExpiresAt   *int64
	Middlewares []domain.PluginConfig
	BudgetLimit float64
	ResetPeriod int
	AutoRenew   bool
}

// UpdateKey modifies an existing API key's configuration.
// It invalidates the cache entry for the key.
func (s *KeyService) UpdateKey(ctx context.Context, input UpdateKeyInput) error {
	k, err := s.repo.GetByID(ctx, domain.ID(input.ID))
	if err != nil {
		return err
	}
	if k == nil {
		return fmt.Errorf("key not found")
	}

	if input.Provider.ID != "" {
		if _, err := s.registry.Get(input.Provider.ID); err != nil {
			return err
		}
	}

	k.Name = input.Name
	k.AutoRenew = input.AutoRenew
	k.Configuration = &domain.KeyConfiguration{
		Provider:    input.Provider,
		Middlewares: input.Middlewares,
		BudgetLimit: input.BudgetLimit,
		ResetPeriod: input.ResetPeriod,
	}

	k.ExpiresAt = nil
	if input.ExpiresAt != nil {
		t := time.Unix(*input.ExpiresAt, 0)
		k.ExpiresAt = &t
	}

	if err := k.Validate(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, k); err != nil {
		return err
	}

	s.cacheMu.Lock()
	delete(s.cache, k.KeyHash)
	s.cacheMu.Unlock()

	return nil
}

// DeleteKey removes an API key from the system and invalidates the cache.
func (s *KeyService) DeleteKey(ctx context.Context, id int64) error {
	k, _ := s.repo.GetByID(ctx, domain.ID(id))
	if err := s.repo.Delete(ctx, domain.ID(id)); err != nil {
		return err
	}
	if k != nil {
		s.cacheMu.Lock()
		delete(s.cache, k.KeyHash)
		s.cacheMu.Unlock()
	}
	return nil
}

// ResetKeyUsage resets the budget usage counter for a key.
func (s *KeyService) ResetKeyUsage(ctx context.Context, k *domain.Key) error {
	k.BudgetUsage = 0
	k.LastResetAt = time.Now()
	if err := s.repo.ResetUsage(ctx, k.ID, k.LastResetAt); err != nil {
		return err
	}
	s.cacheMu.Lock()
	if entry, ok := s.cache[k.KeyHash]; ok {
		entry.key.BudgetUsage = 0
		entry.key.LastResetAt = k.LastResetAt
	}
	s.cacheMu.Unlock()
	return nil
}

// RenewKey extends the expiration of an auto-renewable key and resets its usage.
func (s *KeyService) RenewKey(ctx context.Context, k *domain.Key) error {
	k.BudgetUsage = 0
	k.LastResetAt = time.Now()

	// Extend expiration if it exists
	if k.ExpiresAt != nil {
		period := 30 * 24 * time.Hour // Default 30 days
		if k.Configuration != nil && k.Configuration.ResetPeriod > 0 {
			period = time.Duration(k.Configuration.ResetPeriod) * time.Second
		}
		newExpiry := time.Now().Add(period)
		k.ExpiresAt = &newExpiry
	}

	if err := s.repo.Update(ctx, k); err != nil {
		return err
	}

	s.cacheMu.Lock()
	delete(s.cache, k.KeyHash)
	s.cacheMu.Unlock()

	return nil
}

// IncrementUsage adds to the accumulated cost of a key.
func (s *KeyService) IncrementUsage(ctx context.Context, key *domain.Key, amount float64) error {
	if err := s.repo.IncrementUsage(ctx, key.ID, amount); err != nil {
		return err
	}

	s.cacheMu.Lock()
	if entry, ok := s.cache[key.KeyHash]; ok {
		entry.key.BudgetUsage += amount
	}
	s.cacheMu.Unlock()

	return nil
}

// GetProviderUsage aggregates usage statistics from all registered providers.
func (s *KeyService) GetProviderUsage(ctx context.Context) (map[string]float64, error) {
	providers := s.registry.List()
	usage := make(map[string]float64, len(providers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, p := range providers {
		wg.Add(1)
		go func(p domain.Provider) {
			defer wg.Done()
			u, err := p.GetUsage(ctx)
			if err != nil {
				// Log error but continue with other providers
				fmt.Printf("Error fetching usage for %s: %v\n", p.Name(), err)
				return
			}
			mu.Lock()
			usage[p.Name()] = u
			mu.Unlock()
		}(p)
	}
	wg.Wait()
	return usage, nil
}

// ListProviders returns information about all available providers.
func (s *KeyService) ListProviders(ctx context.Context) ([]domain.ProviderInfo, error) {
	return s.registry.ListInfo(), nil
}

// ListMiddlewares returns information about all available middlewares.
func (s *KeyService) ListMiddlewares(ctx context.Context) ([]domain.MiddlewareInfo, error) {
	return s.mwRegistry.List(), nil
}

// Helpers

func (s *KeyService) generateRandomKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pa-" + hex.EncodeToString(b), nil
}

func (s *KeyService) hashKey(key string) string {
	h := sha256.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *KeyService) copyKey(k *domain.Key) *domain.Key {
	if k == nil {
		return nil
	}
	copy := *k
	if k.ExpiresAt != nil {
		t := *k.ExpiresAt
		copy.ExpiresAt = &t
	}
	if k.Configuration != nil {
		cfg := domain.KeyConfiguration{
			Provider: domain.PluginConfig{
				ID: k.Configuration.Provider.ID,
			},
			Middlewares: make([]domain.PluginConfig, len(k.Configuration.Middlewares)),
			BudgetLimit: k.Configuration.BudgetLimit,
			ResetPeriod: k.Configuration.ResetPeriod,
		}
		if k.Configuration.Provider.Config != nil {
			cfg.Provider.Config = make(map[string]any)
			for k, v := range k.Configuration.Provider.Config {
				cfg.Provider.Config[k] = v
			}
		}
		for i, mw := range k.Configuration.Middlewares {
			mwCopy := domain.PluginConfig{
				ID: mw.ID,
			}
			if mw.Config != nil {
				mwCopy.Config = make(map[string]any)
				for mk, mv := range mw.Config {
					mwCopy.Config[mk] = mv
				}
			}
			cfg.Middlewares[i] = mwCopy
		}
		copy.Configuration = &cfg
	}
	return &copy
}
