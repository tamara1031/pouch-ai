package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"pouch-ai/internal/domain"
	"sync"
	"time"
)

type KeyService struct {
	repo     domain.Repository
	registry domain.Registry
	// In-memory write-behind cache for usage counters
	usageCache map[domain.ID]float64
	cacheMu    sync.RWMutex
}

func NewKeyService(repo domain.Repository, registry domain.Registry) *KeyService {
	return &KeyService{
		repo:       repo,
		registry:   registry,
		usageCache: make(map[domain.ID]float64),
	}
}

func (s *KeyService) CreateKey(ctx context.Context, name string, provider string, expiresAt *int64, budgetLimit float64, budgetPeriod string, isMock bool, mockConfig string, rateLimit int, ratePeriod string) (string, *domain.Key, error) {
	if provider != "" {
		if _, err := s.registry.Get(provider); err != nil {
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
		Name:     name,
		Provider: provider,
		KeyHash:  hash,
		Prefix:   prefix,
		Budget: domain.Budget{
			Limit:  budgetLimit,
			Period: budgetPeriod,
		},
		RateLimit: domain.RateLimit{
			Limit:  rateLimit,
			Period: ratePeriod,
		},
		IsMock:      isMock,
		MockConfig:  mockConfig,
		LastResetAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	if expiresAt != nil {
		t := time.Unix(*expiresAt, 0)
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
	k, err := s.repo.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Apply cached usage if available (Write-Behind pattern)
	s.cacheMu.RLock()
	if cachedUsage, ok := s.usageCache[k.ID]; ok {
		// If cache has a value, it is likely more recent than DB (or at least includes pending writes)
		// Assuming cache stores TOTAL usage.
		// Wait, if cache stores TOTAL usage, we need to make sure we loaded the initial value correctly.
		// If cache is empty, we load from DB.
		// If cache is present, we use cache.
		// BUT: What if another instance updated DB? (We assume single instance).
		// What if cache was evicted? We don't implement eviction yet.
		// Strategy: If in cache, use cache. If not, use DB value.
		// When Incrementing, if not in cache, load from DB first? Or just add to DB value?
		// Better: VerifyKey loads from DB. Increment adds to DB. Cache should track *pending increments*?
		// No, simplest for single instance is Cache is Authoritative for Usage if present.
		// On startup/first load, cache is empty.

		// If we use cache as authoritative:
		// When VerifyKey loads from DB, should we populate cache?
		// Yes, otherwise IncrementUsage won't know the base value.
		k.Budget.Usage = cachedUsage
	} else {
		// Not in cache, so the DB value is current. Populate cache?
		// If we don't populate here, IncrementUsage needs to know base.
		// Let's populate cache on read to ensure it's hot for subsequent increments.
		s.cacheMu.RUnlock() // Drop read lock to acquire write lock
		s.cacheMu.Lock()
		// Double check
		if cachedUsage, ok := s.usageCache[k.ID]; ok {
			k.Budget.Usage = cachedUsage
		} else {
			s.usageCache[k.ID] = k.Budget.Usage
		}
		s.cacheMu.Unlock()
		return k, nil // Return early to avoid double unlock or complex logic
	}
	s.cacheMu.RUnlock()

	return k, nil
}

func (s *KeyService) ListKeys(ctx context.Context) ([]*domain.Key, error) {
	keys, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	for _, k := range keys {
		if val, ok := s.usageCache[k.ID]; ok {
			k.Budget.Usage = val
		}
	}
	return keys, nil
}

func (s *KeyService) UpdateKey(ctx context.Context, id int64, name string, provider string, budgetLimit float64, isMock bool, mockConfig string, rateLimit int, ratePeriod string, expiresAt *int64) error {
	k, err := s.repo.GetByID(ctx, domain.ID(id))
	if err != nil {
		return err
	}
	if k == nil {
		return fmt.Errorf("key not found")
	}

	if provider != "" && provider != k.Provider {
		if _, err := s.registry.Get(provider); err != nil {
			return err
		}
	}

	k.Name = name
	k.Provider = provider
	k.Budget.Limit = budgetLimit
	k.IsMock = isMock
	k.MockConfig = mockConfig
	k.RateLimit.Limit = rateLimit
	k.RateLimit.Period = ratePeriod

	if expiresAt != nil {
		t := time.Unix(*expiresAt, 0)
		k.ExpiresAt = &t
	}

	if err := k.Validate(); err != nil {
		return err
	}

	// Persist update
	if err := s.repo.Update(ctx, k); err != nil {
		return err
	}

	// Ensure cache reflects persistent state if necessary?
	// UpdateKey modifies Budget.Limit, not Usage.
	// But if it *did* modify Usage (it doesn't seem to expose that param), we'd need to update cache.
	return nil
}

func (s *KeyService) DeleteKey(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, domain.ID(id))
	if err == nil {
		s.cacheMu.Lock()
		delete(s.usageCache, domain.ID(id))
		s.cacheMu.Unlock()
	}
	return err
}

func (s *KeyService) ResetKeyUsage(ctx context.Context, k *domain.Key) error {
	k.ResetUsage()

	// Update cache
	s.cacheMu.Lock()
	s.usageCache[k.ID] = 0
	s.cacheMu.Unlock()

	return s.repo.ResetUsage(ctx, k.ID, k.LastResetAt)
}

func (s *KeyService) IncrementUsage(ctx context.Context, id domain.ID, amount float64) error {
	// 1. Update In-Memory Cache Synchronously (Fast)
	s.cacheMu.Lock()

	// We need the current value. If it's not in cache, we technically need to fetch it.
	// However, fetching from DB here would be slow (synchronous).
	// Ideally, VerifyKey has already populated the cache.
	// If it's missing (e.g. server restart and no verify called yet?), we fall back to DB load?
	// OR we optimistically assume 0 + amount? No, that's dangerous.
	// If missing, we MUST load from DB to be correct.
	// Since VerifyKey is called for every request, the cache SHOULD be populated.
	// But let's be safe.

	currentUsage, ok := s.usageCache[id]
	if !ok {
		// Fallback: Load from DB (Slow path, but necessary for correctness on cold cache)
		// We drop lock to avoid blocking readers during DB IO?
		// If we drop lock, another writer might come in.
		// Better to hold lock or use finer granularity.
		// Since we are optimizing for the common case (Hot Cache), blocking on cold miss is acceptable.
		k, err := s.repo.GetByID(ctx, id)
		if err != nil {
			s.cacheMu.Unlock()
			return err
		}
		if k == nil {
			s.cacheMu.Unlock()
			return fmt.Errorf("key not found")
		}
		currentUsage = k.Budget.Usage
	}

	newUsage := currentUsage + amount
	s.usageCache[id] = newUsage
	s.cacheMu.Unlock()

	// 2. Update Database Asynchronously (Slow)
	// We use WithoutCancel to ensure it completes even if request ctx is cancelled.
	go func(ctx context.Context, id domain.ID, amount float64) {
		_ = s.repo.IncrementUsage(ctx, id, amount)
	}(context.WithoutCancel(ctx), id, amount)

	return nil
}

func (s *KeyService) GetProviderUsage(ctx context.Context) (map[string]float64, error) {
	usage := make(map[string]float64)
	for _, p := range s.registry.List() {
		u, err := p.GetUsage(ctx)
		if err != nil {
			// Log error but continue with other providers
			fmt.Printf("Error fetching usage for %s: %v\n", p.Name(), err)
			continue
		}
		usage[p.Name()] = u
	}
	return usage, nil
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
