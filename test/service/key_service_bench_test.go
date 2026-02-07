package service_test

import (
	"context"
	"net/http"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/service"
	"testing"
	"time"
)

type MockProvider struct {
	name  string
	delay time.Duration
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) GetPricing(model domain.Model) (domain.Pricing, error) {
	return domain.Pricing{}, nil
}

func (m *MockProvider) CountTokens(model domain.Model, text string) (int, error) {
	return 0, nil
}

func (m *MockProvider) PrepareHTTPRequest(ctx context.Context, model domain.Model, body []byte) (*http.Request, error) {
	return nil, nil
}

func (m *MockProvider) EstimateUsage(model domain.Model, requestBody []byte) (*domain.Usage, error) {
	return nil, nil
}

func (m *MockProvider) ParseOutputUsage(model domain.Model, responseBody []byte, isStream bool) (int, error) {
	return 0, nil
}

func (m *MockProvider) ParseRequest(body []byte) (domain.Model, bool, error) {
	return "", false, nil
}

func (m *MockProvider) GetUsage(ctx context.Context) (float64, error) {
	time.Sleep(m.delay)
	return 10.0, nil
}

func (m *MockProvider) ProcessStreamChunk(chunk []byte) (string, error) {
	return "", nil
}

type BenchRegistry struct {
	providers []domain.Provider
}

func (r *BenchRegistry) Register(p domain.Provider) {
	r.providers = append(r.providers, p)
}

func (r *BenchRegistry) Get(name string) (domain.Provider, error) {
	for _, p := range r.providers {
		if p.Name() == name {
			return p, nil
		}
	}
	return nil, nil
}

func (r *BenchRegistry) List() []domain.Provider {
	return r.providers
}

func BenchmarkGetProviderUsage(b *testing.B) {
	registry := &BenchRegistry{}
	// Register 10 providers with 50ms delay each
	for i := 0; i < 10; i++ {
		registry.Register(&MockProvider{
			name:  string(rune('A' + i)),
			delay: 50 * time.Millisecond,
		})
	}

	repo := &mockRepo{keys: make(map[domain.ID]*domain.Key)}
	svc := service.NewKeyService(repo, registry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetProviderUsage(context.Background())
	}
}
