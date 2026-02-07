package infra

import (
	"context"
	"io"
	"pouch-ai/internal/domain"
)

type CountingReader struct {
	inner    io.ReadCloser
	provider domain.Provider
	model    domain.Model
	repo     domain.Repository
	keyID    domain.ID
	ctx      context.Context
	buffer   []byte
}

func NewCountingReader(inner io.ReadCloser, provider domain.Provider, model domain.Model, repo domain.Repository, keyID domain.ID, ctx context.Context) *CountingReader {
	return &CountingReader{
		inner:    inner,
		provider: provider,
		model:    model,
		repo:     repo,
		keyID:    keyID,
		ctx:      ctx,
	}
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.inner.Read(p)
	if n > 0 {
		r.buffer = append(r.buffer, p[:n]...)
	}
	return n, err
}

func (r *CountingReader) Close() error {
	// Attempt to calculate usage before closing
	// We do this in a function to ensure we don't return early and skip Close()
	func() {
		// isStream is true because we are in the streaming path
		outputTokens, err := r.provider.ParseOutputUsage(r.model, r.buffer, true)
		if err != nil {
			return
		}

		pricing, err := r.provider.GetPricing(r.model)
		if err != nil {
			return
		}

		outputCost := float64(outputTokens) / 1000.0 * pricing.Output

		if r.repo != nil && outputCost > 0 {
			// Update usage
			_ = r.repo.IncrementUsage(r.ctx, r.keyID, outputCost)
		}
	}()

	return r.inner.Close()
}
