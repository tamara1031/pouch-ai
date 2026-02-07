package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/util/logger"
	"pouch-ai/backend/util/registry"
	"sync"
	"time"
)

type cachedKey struct {
	key       *domain.Key
	expiresAt time.Time
}

type KeyService struct {
	repo       domain.Repository
	registry   domain.ProviderRegistry
	mwRegistry domain.MiddlewareRegistry
	cache      map[string]cachedKey
	cacheMu    sync.RWMutex
}

func NewKeyService(repo domain.Repository, registry domain.ProviderRegistry, mwRegistry domain.MiddlewareRegistry) *KeyService {
	return &KeyService{
		repo:       repo,
		registry:   registry,
		mwRegistry: mwRegistry,
		cache:      make(map[string]cachedKey),
	}
}

type CreateKeyInput struct {
	Name        string
	Provider    domain.PluginConfig
	ExpiresAt   *int64
	Middlewares []domain.PluginConfig
	BudgetLimit float64
	ResetPeriod int
	AutoRenew   bool
}

func (s *KeyService) CreateKey(ctx context.Context, input CreateKeyInput) (string, *domain.Key, error) {
	if input.Provider.ID != "" {
		if _, err := s.registry.Get(input.Provider.ID); err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				return "", nil, domain.ErrProviderNotFound
			}
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
		return nil, domain.ErrInvalidKey
	}

	s.cacheMu.Lock()
	s.cache[hash] = cachedKey{
		key:       s.copyKey(k),
		expiresAt: time.Now().Add(1 * time.Minute),
	}
	s.cacheMu.Unlock()

	return k, nil
}

func (s *KeyService) ListKeys(ctx context.Context) ([]*domain.Key, error) {
	return s.repo.List(ctx)
}

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

func (s *KeyService) UpdateKey(ctx context.Context, input UpdateKeyInput) error {
	k, err := s.repo.GetByID(ctx, domain.ID(input.ID))
	if err != nil {
		return err
	}
	if k == nil {
		return domain.ErrKeyNotFound
	}

	if input.Provider.ID != "" {
		if _, err := s.registry.Get(input.Provider.ID); err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				return domain.ErrProviderNotFound
			}
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

func (s *KeyService) ReserveUsage(ctx context.Context, keyID domain.ID, amount float64) error {
	k, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return err
	}
	if k == nil {
		return domain.ErrKeyNotFound
	}

	if k.Configuration != nil && k.Configuration.BudgetLimit > 0 {
		if k.BudgetUsage+amount > k.Configuration.BudgetLimit {
			return fmt.Errorf("%w (limit: $%.2f, current+reservation: $%.2f)", domain.ErrBudgetExceeded, k.Configuration.BudgetLimit, k.BudgetUsage+amount)
		}
	}

	if err := s.repo.IncrementUsage(ctx, keyID, amount); err != nil {
		return err
	}

	s.cacheMu.Lock()
	for _, entry := range s.cache {
		if entry.key.ID == keyID {
			entry.key.BudgetUsage += amount
			break
		}
	}
	s.cacheMu.Unlock()

	return nil
}

func (s *KeyService) CommitUsage(ctx context.Context, keyID domain.ID, reserved, actual float64) error {
	diff := actual - reserved
	if diff == 0 {
		return nil
	}

	if err := s.repo.IncrementUsage(ctx, keyID, diff); err != nil {
		return err
	}

	s.cacheMu.Lock()
	for _, entry := range s.cache {
		if entry.key.ID == keyID {
			entry.key.BudgetUsage += diff
			break
		}
	}
	s.cacheMu.Unlock()

	return nil
}

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
				logger.L.Error("failed to fetch usage", "provider", p.Name(), "error", err)
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

func (s *KeyService) ListProviders(ctx context.Context) ([]domain.PluginInfo, error) {
	providers := s.registry.List()
	infos := make([]domain.PluginInfo, 0, len(providers))
	for _, p := range providers {
		infos = append(infos, domain.PluginInfo{
			ID:     p.Name(),
			Schema: p.Schema(),
		})
	}
	return infos, nil
}

func (s *KeyService) ListMiddlewares(ctx context.Context) ([]domain.PluginInfo, error) {
	// For now, middlewares in registry are just factories.
	// We might need a separate way to get their Info if we want to list them with schemas.
	// But according to current registration, we only have names in registry.
	// Wait, the mwRegistry is Registry[func(map[string]any) Middleware].
	// Generic Registry.List() returns the items (factories).
	// Registry.ListKeys() returns the IDs.

	ids := s.mwRegistry.ListKeys()
	infos := make([]domain.PluginInfo, 0, len(ids))
	for _, id := range ids {
		infos = append(infos, domain.PluginInfo{
			ID: id,
			// Schema is missing here because we only registered the factory.
			// This is a known limitation that I should probably fix by registering a struct.
		})
	}
	return infos, nil
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
