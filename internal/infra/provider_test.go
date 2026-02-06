package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"pouch-ai/internal/domain"
)

type mockCounter struct{}

func (m *mockCounter) Count(model string, text string) (int, error) {
	return len(text), nil
}

func TestAnthropicProvider_PrepareHTTPRequest(t *testing.T) {
	p := NewAnthropicProvider("key", &AnthropicPricing{}, &mockCounter{})

	oBody := `{"model":"claude-3-opus","messages":[{"role":"user","content":"Hello"}],"stream":false}`
	req, err := p.PrepareHTTPRequest(context.Background(), domain.Model("claude-3-opus"), []byte(oBody))
	if err != nil {
		t.Fatalf("PrepareHTTPRequest failed: %v", err)
	}

	if req.Header.Get("x-api-key") != "key" {
		t.Errorf("expected x-api-key header")
	}

	body, _ := io.ReadAll(req.Body)
	var aReq struct {
		Model string `json:"model"`
		Messages []struct {
			Role string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		MaxTokens int `json:"max_tokens"`
	}
	if err := json.Unmarshal(body, &aReq); err != nil {
		t.Fatalf("Failed to parse request body: %v", err)
	}

	if aReq.Model != "claude-3-opus" {
		t.Errorf("expected model claude-3-opus, got %s", aReq.Model)
	}
	if aReq.MaxTokens != 4096 {
		t.Errorf("expected default max_tokens 4096, got %d", aReq.MaxTokens)
	}
	if len(aReq.Messages) != 1 {
		t.Errorf("expected 1 message")
	}
}

func TestAnthropicProvider_TransformResponse(t *testing.T) {
	p := NewAnthropicProvider("key", &AnthropicPricing{}, &mockCounter{})

	aResp := `{"content":[{"text":"Response"}],"usage":{"input_tokens":10,"output_tokens":5}}`

	reader, err := p.TransformResponse(bytes.NewBufferString(aResp), false)
	if err != nil {
		t.Fatalf("TransformResponse failed: %v", err)
	}

	out, _ := io.ReadAll(reader)
	var oResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(out, &oResp); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if oResp.Choices[0].Message.Content != "Response" {
		t.Errorf("expected content 'Response', got %s", oResp.Choices[0].Message.Content)
	}
}

func TestGeminiProvider_PrepareHTTPRequest(t *testing.T) {
	p := NewGeminiProvider("key", &GeminiPricing{}, &mockCounter{})

	oBody := `{"model":"gemini-1.5-pro","messages":[{"role":"user","content":"Hello"}],"stream":false}`
	req, err := p.PrepareHTTPRequest(context.Background(), domain.Model("gemini-1.5-pro"), []byte(oBody))
	if err != nil {
		t.Fatalf("PrepareHTTPRequest failed: %v", err)
	}

	if !bytes.Contains([]byte(req.URL.String()), []byte("key=key")) {
		t.Errorf("expected key in url")
	}

	body, _ := io.ReadAll(req.Body)
	var gReq struct {
		Contents []struct {
			Role string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &gReq); err != nil {
		t.Fatalf("Failed to parse request body: %v", err)
	}

	if len(gReq.Contents) != 1 {
		t.Errorf("expected 1 content")
	}
	if gReq.Contents[0].Role != "user" {
		t.Errorf("expected role user, got %s", gReq.Contents[0].Role)
	}
}

func TestGeminiProvider_TransformResponse(t *testing.T) {
	p := NewGeminiProvider("key", &GeminiPricing{}, &mockCounter{})

	gResp := `{"candidates":[{"content":{"parts":[{"text":"Response"}]},"finishReason":"STOP"}]}`

	reader, err := p.TransformResponse(bytes.NewBufferString(gResp), false)
	if err != nil {
		t.Fatalf("TransformResponse failed: %v", err)
	}

	out, _ := io.ReadAll(reader)
	var oResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(out, &oResp); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	if oResp.Choices[0].Message.Content != "Response" {
		t.Errorf("expected content 'Response', got %s", oResp.Choices[0].Message.Content)
	}
}
