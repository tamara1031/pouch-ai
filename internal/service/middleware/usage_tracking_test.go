package middleware

import (
	"context"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/service"
	"testing"
	"time"
)

type mockRepo struct{}

func (m *mockRepo) IncrementUsage(ctx context.Context, id domain.ID, amount float64) error {
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (m *mockRepo) Save(ctx context.Context, k *domain.Key) error { return nil }
func (m *mockRepo) GetByID(ctx context.Context, id domain.ID) (*domain.Key, error) { return nil, nil }
func (m *mockRepo) GetByHash(ctx context.Context, hash string) (*domain.Key, error) { return nil, nil }
func (m *mockRepo) List(ctx context.Context) ([]*domain.Key, error) { return nil, nil }
func (m *mockRepo) Update(ctx context.Context, k *domain.Key) error { return nil }
func (m *mockRepo) Delete(ctx context.Context, id domain.ID) error { return nil }
func (m *mockRepo) ResetUsage(ctx context.Context, id domain.ID, lastResetAt time.Time) error { return nil }

type mockRegistry struct{}

func (m *mockRegistry) Register(p domain.Provider) {}
func (m *mockRegistry) Get(name string) (domain.Provider, error) { return nil, nil }
func (m *mockRegistry) List() []domain.Provider { return nil }

func BenchmarkUsageTrackingMiddleware(b *testing.B) {
	repo := &mockRepo{}
	registry := &mockRegistry{}
	keyService := service.NewKeyService(repo, registry)
	mw := NewUsageTrackingMiddleware(keyService)

	handler := domain.HandlerFunc(func(req *domain.Request) (*domain.Response, error) {
		return &domain.Response{
			TotalCost: 0.01,
		}, nil
	})

	req := &domain.Request{
		Context: context.Background(),
		Key: &domain.Key{
			ID: 1,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mw.Execute(req, handler)
	}
}
