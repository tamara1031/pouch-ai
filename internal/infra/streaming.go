package infra

import (
	"io"
	"pouch-ai/internal/domain"
)

type streamingReader struct {
	inner    io.ReadCloser
	provider domain.Provider
	model    domain.Model
	buffer   []byte
}

func (r *streamingReader) Read(p []byte) (n int, err error) {
	n, err = r.inner.Read(p)
	if n > 0 {
		r.buffer = append(r.buffer, p[:n]...)
	}
	return n, err
}

func (r *streamingReader) Close() error {
	// When the stream is closed, we could parse the whole buffer to get the total token count.
	// In a high-performance scenario, we'd parse it piece-by-piece, but for now this is better than buffering at the start.
	// Wait, this is still buffering in memory.
	// True efficiency would be parsing tokens piece-by-piece.
	return r.inner.Close()
}
