package service_test

import (
	"context"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"
	"testing"
	"time"
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
	domain.ProviderRegistry
}

func (m *mockRegistry) Register(name string, item domain.Provider) {}

func (m *mockRegistry) Get(name string) (domain.Provider, error) {
	return nil, nil
}

func (m *mockRegistry) List() []domain.Provider {
	return nil
}

func (m *mockRegistry) ListKeys() []string {
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id domain.ID) (*domain.Key, error) {
	return m.keys[id], nil
}

func (m *mockRepo) List(ctx context.Context) ([]*domain.Key, error) {
	return nil, nil
}

func (m *mockRepo) Update(ctx context.Context, k *domain.Key) error {
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id domain.ID) error {
	return nil
}

func (m *mockRepo) IncrementUsage(ctx context.Context, id domain.ID, amount float64) error {
	return nil
}

func (m *mockRepo) ResetUsage(ctx context.Context, id domain.ID, lastResetAt time.Time) error {
	return nil
}

func TestKeyService_CreateKey(t *testing.T) {
	repo := &mockRepo{keys: make(map[domain.ID]*domain.Key)}
	reg := &mockRegistry{}
	mwReg := domain.NewMiddlewareRegistry()
	svc := service.NewKeyService(repo, reg, mwReg)

	input := service.CreateKeyInput{
		Name:     "test-key",
		Provider: domain.PluginConfig{ID: "openai"},
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
	mwReg := domain.NewMiddlewareRegistry()
	svc := service.NewKeyService(repo, reg, mwReg)

	input := service.CreateKeyInput{
		Name:     "test-key",
		Provider: domain.PluginConfig{ID: "openai"},
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
