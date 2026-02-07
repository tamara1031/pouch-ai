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
}

func NewKeyService(repo domain.Repository, registry domain.Registry) *KeyService {
	return &KeyService{repo: repo, registry: registry}
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

	return k, nil
}

func (s *KeyService) ListKeys(ctx context.Context) ([]*domain.Key, error) {
	return s.repo.List(ctx)
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

	return s.repo.Update(ctx, k)
}

func (s *KeyService) DeleteKey(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, domain.ID(id))
}

func (s *KeyService) ResetKeyUsage(ctx context.Context, k *domain.Key) error {
	k.ResetUsage()
	return s.repo.ResetUsage(ctx, k.ID, k.LastResetAt)
}

func (s *KeyService) IncrementUsage(ctx context.Context, id domain.ID, amount float64) error {
	return s.repo.IncrementUsage(ctx, id, amount)
}

func (s *KeyService) ReserveUsage(ctx context.Context, id domain.ID, amount float64) error {
	return s.repo.IncrementUsageWithLimit(ctx, id, amount)
}

func (s *KeyService) CompleteUsage(ctx context.Context, id domain.ID, realCost float64, reservedCost float64) error {
	// The difference can be negative (if realCost < reservedCost), which means we refund the difference.
	diff := realCost - reservedCost
	// If diff is 0, we don't need to do anything
	if diff == 0 {
		return nil
	}
	return s.repo.IncrementUsage(ctx, id, diff)
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
