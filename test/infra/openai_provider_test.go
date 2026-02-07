package infra_test

import (
	"pouch-ai/backend/domain"
	"pouch-ai/plugins/providers"
	"testing"
)

type mockCounter struct{}

func (m *mockCounter) Count(model string, text string) (int, error) {
	return len(text) / 4, nil
}

func TestOpenAIProvider_ParseRequest(t *testing.T) {
	p := providers.NewOpenAIProvider("test-key", "", nil, &mockCounter{})

	reqBody := `{"model": "gpt-4", "stream": true}`
	model, isStream, err := p.ParseRequest([]byte(reqBody))
	if err != nil {
		t.Fatalf("Failed to parse request: %v", err)
	}

	if model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", model)
	}
	if !isStream {
		t.Error("Expected stream to be true")
	}
}

func TestOpenAIProvider_ParseOutputUsage_NonStream(t *testing.T) {
	p := providers.NewOpenAIProvider("test-key", "", nil, &mockCounter{})

	respBody := `{"usage": {"completion_tokens": 42}}`
	tokens, err := p.ParseOutputUsage(domain.Model("gpt-4"), []byte(respBody), false)
	if err != nil {
		t.Fatalf("Failed to parse output usage: %v", err)
	}

	if tokens != 42 {
		t.Errorf("Expected 42 tokens, got %d", tokens)
	}
}
