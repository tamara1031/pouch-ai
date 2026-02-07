package infra

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"pouch-ai/backend/domain"
	"strings"
	"testing"
)

// TestMockProvider implements domain.Provider for testing
type TestMockProvider struct{}

func (m *TestMockProvider) Name() string { return "mock" }
func (m *TestMockProvider) Configure(config map[string]string) (domain.Provider, error) {
	return m, nil
}
func (m *TestMockProvider) GetPricing(model domain.Model) (domain.Pricing, error) {
	return domain.Pricing{}, nil
}
func (m *TestMockProvider) CountTokens(model domain.Model, text string) (int, error) {
	return len(text) / 4, nil // Rough approximation
}
func (m *TestMockProvider) PrepareHTTPRequest(ctx context.Context, model domain.Model, body []byte) (*http.Request, error) {
	return nil, nil
}
func (m *TestMockProvider) EstimateUsage(model domain.Model, body []byte) (*domain.Usage, error) {
	return nil, nil
}
func (m *TestMockProvider) ParseOutputUsage(model domain.Model, responseBody []byte, isStream bool) (int, error) {
	return 0, nil
}
func (m *TestMockProvider) ProcessStreamChunk(chunk []byte) (string, error) {
	// Simple mock implementation
	s := string(chunk)
	if strings.Contains(s, "word ") {
		return "word ", nil
	}
	return "", nil
}
func (m *TestMockProvider) ParseRequest(body []byte) (domain.Model, bool, error) {
	return "", false, nil
}
func (m *TestMockProvider) GetUsage(ctx context.Context) (float64, error) {
	return 0, nil
}

func BenchmarkStreamingReader(b *testing.B) {
	// Create a large SSE stream
	var buf bytes.Buffer
	chunk := `data: {"choices":[{"delta":{"content":"word "}}]}` + "\n"

	// 10000 chunks
	iterations := 10000
	for i := 0; i < iterations; i++ {
		buf.WriteString(chunk)
	}
	buf.WriteString("data: [DONE]\n\n")

	rawData := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := NewStreamingReader(
			io.NopCloser(bytes.NewReader(rawData)),
			&TestMockProvider{},
			"test-model",
		)

		p := make([]byte, 1024)
		for {
			_, err := reader.Read(p)
			if err == io.EOF {
				break
			}
		}
		reader.Close()
	}
}

func TestStreamingReader(t *testing.T) {
	chunk := `data: {"choices":[{"delta":{"content":"word "}}]}` + "\n"
	stream := chunk + chunk + "data: [DONE]\n"

	reader := NewStreamingReader(
		io.NopCloser(strings.NewReader(stream)),
		&TestMockProvider{},
		"test-model",
	)

	p := make([]byte, 10) // Small buffer to force multiple reads
	var output bytes.Buffer
	for {
		n, err := reader.Read(p)
		if n > 0 {
			output.Write(p[:n])
		}
		if err == io.EOF {
			break
		}
	}

	if output.String() != stream {
		t.Errorf("Expected output to match input stream. Got len %d, expected len %d", output.Len(), len(stream))
	}

	// Verify internal accumulation (white-box testing)
	if reader.accumulatedText.String() != "word word " {
		t.Errorf("Expected accumulated text 'word word ', got '%s'", reader.accumulatedText.String())
	}

	reader.Close()
	// MockProvider returns len/4 tokens. "word word " is 10 chars. 10/4 = 2.
	if reader.Usage() != 2 {
		t.Errorf("Expected usage 2, got %d", reader.Usage())
	}
}
