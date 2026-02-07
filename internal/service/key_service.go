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

type cachedKey struct {
	key       *domain.Key
	expiresAt time.Time
}

type KeyService struct {
	repo     domain.Repository
	registry domain.Registry
	cache    map[string]cachedKey
	cacheMu  sync.RWMutex
}

func NewKeyService(repo domain.Repository, registry domain.Registry) *KeyService {
	return &KeyService{
		repo:     repo,
		registry: registry,
		cache:    make(map[string]cachedKey),
	}
}

type CreateKeyInput struct {
	Name         string
	Provider     string
	ExpiresAt    *int64
	BudgetLimit  float64
	BudgetPeriod string
	RateLimit    int
	RatePeriod   string
	IsMock       bool
	MockConfig   string
}

func (s *KeyService) CreateKey(ctx context.Context, input CreateKeyInput) (string, *domain.Key, error) {
	if input.Provider != "" {
		if _, err := s.registry.Get(input.Provider); err != nil {
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
		Name:     input.Name,
		Provider: input.Provider,
		KeyHash:  hash,
		Prefix:   prefix,
		Budget: domain.Budget{
			Limit:  input.BudgetLimit,
			Period: input.BudgetPeriod,
		},
		RateLimit: domain.RateLimit{
			Limit:  input.RateLimit,
			Period: input.RatePeriod,
		},
		IsMock:      input.IsMock,
		MockConfig:  input.MockConfig,
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

func (s *KeyService) ListKeys(ctx context.Context) ([]*domain.Key, error) {
	return s.repo.List(ctx)
}

type UpdateKeyInput struct {
	ID          int64
	Name        string
	Provider    string
	BudgetLimit float64
	RateLimit   int
	RatePeriod  string
	IsMock      bool
	MockConfig  string
	ExpiresAt   *int64
}

func (s *KeyService) UpdateKey(ctx context.Context, input UpdateKeyInput) error {
	k, err := s.repo.GetByID(ctx, domain.ID(input.ID))
	if err != nil {
		return err
	}
	if k == nil {
		return fmt.Errorf("key not found")
	}

	if input.Provider != "" && input.Provider != k.Provider {
		if _, err := s.registry.Get(input.Provider); err != nil {
			return err
		}
	}

	k.Name = input.Name
	k.Provider = input.Provider
	k.Budget.Limit = input.BudgetLimit
	k.IsMock = input.IsMock
	k.MockConfig = input.MockConfig
	k.RateLimit.Limit = input.RateLimit
	k.RateLimit.Period = input.RatePeriod

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
	k.ResetUsage()
	if err := s.repo.ResetUsage(ctx, k.ID, k.LastResetAt); err != nil {
		return err
	}
	s.cacheMu.Lock()
	if entry, ok := s.cache[k.KeyHash]; ok {
		entry.key.ResetUsage()
		entry.key.LastResetAt = k.LastResetAt
	}
	s.cacheMu.Unlock()
	return nil
}

func (s *KeyService) IncrementUsage(ctx context.Context, key *domain.Key, amount float64) error {
	if err := s.repo.IncrementUsage(ctx, key.ID, amount); err != nil {
		return err
	}

	s.cacheMu.Lock()
	if entry, ok := s.cache[key.KeyHash]; ok {
		entry.key.Budget.Usage += amount
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

func (s *KeyService) ListProviders(ctx context.Context) ([]string, error) {
	providers := s.registry.List()
	names := make([]string, 0, len(providers))
	for _, p := range providers {
		names = append(names, p.Name())
	}
	return names, nil
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
	return &copy
}
