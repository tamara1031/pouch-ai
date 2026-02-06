package proxy

import (
	"net/http/httptest"
	"testing"

	"pouch-ai/internal/token"
)

func TestCalculateTokens_JSON(t *testing.T) {
	tok := token.NewCounter()
	req := OpenAIRequest{Model: "gpt-3.5-turbo"}

	rw := &CostTrackingResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
		Mode:           "json",
		Req:            req,
		Token:          tok,
	}

	body := `{"id":"1","usage":{"completion_tokens":42}}`
	rw.Write([]byte(body))
	rw.CalculateTokens()

	if rw.OutputTokens != 42 {
		t.Errorf("Expected 42 tokens, got %d", rw.OutputTokens)
	}
}

func TestCalculateTokens_Stream(t *testing.T) {
	tok := token.NewCounter()
	req := OpenAIRequest{Model: "gpt-3.5-turbo"}

	rw := &CostTrackingResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
		Mode:           "stream",
		Req:            req,
		Token:          tok,
	}

	chunks := []string{
		`data: {"choices":[{"delta":{"content":"Hello"}}]}`,
		`data: {"choices":[{"delta":{"content":" world"}}]}`,
		`data: [DONE]`,
	}

	for _, chunk := range chunks {
		rw.Write([]byte(chunk + "\n"))
	}

	rw.CalculateTokens()

	// "Hello world" -> 2 tokens roughly. tiktoken might say 2.
	if rw.OutputTokens == 0 {
		t.Error("Expected > 0 tokens")
	}
	t.Logf("Tokens: %d", rw.OutputTokens)
}

func TestCalculateTokens_Stream_Empty(t *testing.T) {
	tok := token.NewCounter()
	req := OpenAIRequest{Model: "gpt-3.5-turbo"}

	rw := &CostTrackingResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
		Mode:           "stream",
		Req:            req,
		Token:          tok,
	}

	rw.CalculateTokens()
	if rw.OutputTokens != 0 {
		t.Errorf("Expected 0 tokens, got %d", rw.OutputTokens)
	}
}
