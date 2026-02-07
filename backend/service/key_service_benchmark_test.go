package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"pouch-ai/internal/domain"
	"testing"
	"time"
)

// MockRepository simulates a database with latency
type MockRepository struct {
	keys map[string]*domain.Key
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		keys: make(map[string]*domain.Key),
	}
}

func (m *MockRepository) Save(ctx context.Context, k *domain.Key) error {
	m.keys[k.KeyHash] = k
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Key, error) {
	for _, k := range m.keys {
		if k.ID == id {
			return k, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) GetByHash(ctx context.Context, hash string) (*domain.Key, error) {
	// Simulate DB latency for VerifyKey
	time.Sleep(100 * time.Microsecond) // Small latency to make cache visible
	if k, ok := m.keys[hash]; ok {
		return k, nil
	}
	return nil, nil
}

func (m *MockRepository) List(ctx context.Context) ([]*domain.Key, error) {
	var keys []*domain.Key
	for _, k := range m.keys {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *MockRepository) Update(ctx context.Context, k *domain.Key) error {
	m.keys[k.KeyHash] = k
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id domain.ID) error {
	for hash, k := range m.keys {
		if k.ID == id {
			delete(m.keys, hash)
			return nil
		}
	}
	return nil
}

func (m *MockRepository) IncrementUsage(ctx context.Context, id domain.ID, amount float64) error {
	for _, k := range m.keys {
		if k.ID == id {
			k.Budget.Usage += amount
			return nil
		}
	}
	return nil
}

func (m *MockRepository) ResetUsage(ctx context.Context, id domain.ID, lastResetAt time.Time) error {
	for _, k := range m.keys {
		if k.ID == id {
			k.Budget.Usage = 0
			k.LastResetAt = lastResetAt
			return nil
		}
	}
	return nil
}

// MockRegistry
type MockRegistry struct{}

func (m *MockRegistry) Register(p domain.Provider) {}
func (m *MockRegistry) Get(name string) (domain.Provider, error) {
	return &DummyProvider{}, nil
}
func (m *MockRegistry) List() []domain.Provider { return nil }

// DummyProvider
type DummyProvider struct{}

func (d *DummyProvider) Name() string { return "dummy" }
func (d *DummyProvider) GetPricing(model domain.Model) (domain.Pricing, error) {
	return domain.Pricing{}, nil
}
func (d *DummyProvider) CountTokens(model domain.Model, text string) (int, error) { return 0, nil }
func (d *DummyProvider) PrepareHTTPRequest(ctx context.Context, model domain.Model, body []byte) (*http.Request, error) {
	return nil, nil
}
func (d *DummyProvider) EstimateUsage(model domain.Model, requestBody []byte) (*domain.Usage, error) {
	return nil, nil
}
func (d *DummyProvider) ParseOutputUsage(model domain.Model, responseBody []byte, isStream bool) (int, error) {
	return 0, nil
}
func (d *DummyProvider) ParseRequest(body []byte) (domain.Model, bool, error) {
	return "model", false, nil
}
func (d *DummyProvider) GetUsage(ctx context.Context) (float64, error) { return 0, nil }

func (d *DummyProvider) ProcessStreamChunk(chunk []byte) (string, error) {
	return "", nil
}

func BenchmarkVerifyKey(b *testing.B) {
	repo := NewMockRepository()
	registry := &MockRegistry{}
	svc := NewKeyService(repo, registry)

	// Create a key manually to avoid needing CreateKey internals
	rawKey := "pa-benchmarkkey1234567890"
	hash := sha256.Sum256([]byte(rawKey))
	hashStr := hex.EncodeToString(hash[:])

	k := &domain.Key{
		ID:       1,
		Name:     "Bench Key",
		Provider: "dummy",
		KeyHash:  hashStr,
	}
	repo.Save(context.Background(), k)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.VerifyKey(ctx, rawKey)
		if err != nil {
			b.Fatalf("VerifyKey failed: %v", err)
		}
	}
}
