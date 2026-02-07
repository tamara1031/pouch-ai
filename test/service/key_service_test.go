package service_test

import (
	"context"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/service"
	"testing"
)

type mockRepo struct {
	domain.Repository
	keys map[domain.ID]*domain.Key
}

func (m *mockRepo) Save(ctx context.Context, k *domain.Key) error {
	m.keys[k.ID] = k
	return nil
}

func (m *mockRepo) GetByHash(ctx context.Context, hash string) (*domain.Key, error) {
	for _, k := range m.keys {
		if k.KeyHash == hash {
			return k, nil
		}
	}
	return nil, nil
}

type mockRegistry struct {
	domain.Registry
}

func (m *mockRegistry) Get(name string) (domain.Provider, error) {
	return nil, nil // Simple mock
}

func TestKeyService_CreateKey(t *testing.T) {
	repo := &mockRepo{keys: make(map[domain.ID]*domain.Key)}
	reg := &mockRegistry{}
	svc := service.NewKeyService(repo, reg)

	input := service.CreateKeyInput{
		Name:         "test-key",
		Provider:     "openai",
		BudgetLimit:  10.0,
		BudgetPeriod: "monthly",
		RateLimit:    10,
		RatePeriod:   "minute",
	}

	raw, key, err := svc.CreateKey(context.Background(), input)
	if err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	if raw == "" {
		t.Error("Raw key should not be empty")
	}
	if key.Name != "test-key" {
		t.Errorf("Expected name test-key, got %s", key.Name)
	}
}

func TestKeyService_VerifyKey(t *testing.T) {
	repo := &mockRepo{keys: make(map[domain.ID]*domain.Key)}
	reg := &mockRegistry{}
	svc := service.NewKeyService(repo, reg)

	input := service.CreateKeyInput{
		Name:         "test-key",
		Provider:     "openai",
		BudgetLimit:  10.0,
		BudgetPeriod: "monthly",
		RateLimit:    10,
		RatePeriod:   "minute",
	}

	raw, _, _ := svc.CreateKey(context.Background(), input)

	key, err := svc.VerifyKey(context.Background(), raw)
	if err != nil {
		t.Fatalf("Failed to verify key: %v", err)
	}
	if key.Name != "test-key" {
		t.Errorf("Expected name test-key, got %s", key.Name)
	}
}
