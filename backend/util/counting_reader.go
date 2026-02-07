package util

import (
	"bytes"
	"context"
	"io"
	"pouch-ai/backend/domain"
)

type CountingReader struct {
	inner       io.ReadCloser
	provider    domain.Provider
	model       domain.Model
	committer   domain.UsageCommitter
	keyID       domain.ID
	ctx         context.Context
	pending     []byte
	totalTokens int
	finalUsage  *domain.Usage
	reserved    float64
}

func NewCountingReader(inner io.ReadCloser, provider domain.Provider, model domain.Model, committer domain.UsageCommitter, keyID domain.ID, reserved float64, ctx context.Context) io.ReadCloser {
	return &CountingReader{
		inner:     inner,
		provider:  provider,
		model:     model,
		committer: committer,
		keyID:     keyID,
		reserved:  reserved,
		ctx:       ctx,
	}
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.inner.Read(p)
	if n > 0 {
		r.pending = append(r.pending, p[:n]...)
		for {
			idx := bytes.IndexByte(r.pending, '\n')
			if idx == -1 {
				break
			}
			line := r.pending[:idx+1]
			_, tokens, usage, _ := r.provider.ParseStreamChunk(r.model, line)
			if usage != nil {
				r.finalUsage = usage
			}
			r.totalTokens += tokens
			r.pending = r.pending[idx+1:]
		}
	}
	return n, err
}

func (r *CountingReader) Close() error {
	defer r.inner.Close()

	var actual float64
	if r.finalUsage != nil {
		actual = r.finalUsage.TotalCost
	} else {
		pricing, err := r.provider.GetPricing(r.model)
		if err == nil {
			actual = float64(r.totalTokens) / 1000.0 * pricing.Output
		}
	}

	if r.committer != nil {
		_ = r.committer.CommitUsage(r.ctx, r.keyID, r.reserved, actual)
	}

	return nil
}
