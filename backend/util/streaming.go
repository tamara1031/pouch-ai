package util

import (
	"bytes"
	"io"
	"pouch-ai/backend/domain"
	"strings"
)

type streamingReader struct {
	inner           io.ReadCloser
	provider        domain.Provider
	model           domain.Model
	pending         []byte
	accumulatedText strings.Builder
	UsageTokenCount int
}

func NewStreamingReader(inner io.ReadCloser, provider domain.Provider, model domain.Model) *streamingReader {
	return &streamingReader{
		inner:    inner,
		provider: provider,
		model:    model,
		pending:  make([]byte, 0, 4096),
	}
}

func (r *streamingReader) Usage() int {
	return r.UsageTokenCount
}

func (r *streamingReader) Read(p []byte) (n int, err error) {
	n, err = r.inner.Read(p)
	if n > 0 {
		r.pending = append(r.pending, p[:n]...)

		// Process complete lines
		for {
			idx := bytes.IndexByte(r.pending, '\n')
			if idx == -1 {
				break
			}

			line := r.pending[:idx+1]
			content, _ := r.provider.ProcessStreamChunk(line)
			if content != "" {
				r.accumulatedText.WriteString(content)
			}

			// Slice off the processed line.
			// Note: This does not free memory immediately but append() will handle reallocation eventually.
			// To be more memory efficient we could copy to start if it gets too big and empty,
			// but for streaming chunks it's usually fine.
			r.pending = r.pending[idx+1:]
		}
	}
	return n, err
}

func (r *streamingReader) Close() error {
	// When the stream is closed, we calculate the token count.
	// We no longer have the huge buffer, only the accumulated text.
	if r.accumulatedText.Len() > 0 {
		count, _ := r.provider.CountTokens(r.model, r.accumulatedText.String())
		r.UsageTokenCount = count
	}
	return r.inner.Close()
}
