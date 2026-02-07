package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"pouch-ai/backend/config"
	"pouch-ai/backend/domain"
	"strings"
	"sync"
	"time"
)

type MockProvider struct {
	server *httptest.Server
	config map[string]string
	mu     *sync.RWMutex
}

type MockBuilder struct{}

func (b *MockBuilder) Build(ctx context.Context, cfg *config.Config) (domain.Provider, error) {
	return NewMockProvider(), nil
}

func NewMockProvider() *MockProvider {
	p := &MockProvider{
		config: make(map[string]string),
		mu:     &sync.RWMutex{},
	}
	p.server = httptest.NewServer(http.HandlerFunc(p.handleRequest))
	return p
}

func (p *MockProvider) Schema() domain.PluginSchema {
	return domain.PluginSchema{
		"mock_response": {
			Type:        domain.FieldTypeString,
			DisplayName: "Mock Response",
			Default:     "This is a mock response from the Pouch AI Mock Provider.",
			Description: "The response message to return for all requests.",
		},
	}
}

func (p *MockProvider) Configure(config map[string]any) (domain.Provider, error) {
	newP := *p
	// Share the server and lock, but copy the config
	newP.config = make(map[string]string)
	for k, v := range config {
		if s, ok := v.(string); ok {
			newP.config[k] = s
		}
	}
	return &newP, nil
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) GetPricing(model domain.Model) (domain.Pricing, error) {
	// Mock is free
	return domain.Pricing{Input: 0, Output: 0}, nil
}

func (p *MockProvider) CountTokens(model domain.Model, text string) (int, error) {
	// Simple approximation: 4 chars = 1 token
	return len(text) / 4, nil
}

func (p *MockProvider) PrepareHTTPRequest(ctx context.Context, model domain.Model, body []byte) (*http.Request, error) {
	// Direct the request to our internal mock server
	req, err := http.NewRequestWithContext(ctx, "POST", p.server.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (p *MockProvider) EstimateUsage(model domain.Model, requestBody []byte) (*domain.Usage, error) {
	var req struct {
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(requestBody, &req); err != nil {
		return nil, err
	}

	totalLen := 0
	for _, m := range req.Messages {
		totalLen += len(m.Content)
	}

	tokens := totalLen / 4
	return &domain.Usage{
		InputTokens: tokens,
		TotalCost:   0,
	}, nil
}

func (p *MockProvider) ParseOutputUsage(model domain.Model, responseBody []byte, isStream bool) (int, error) {
	if !isStream {
		var resp struct {
			Usage struct {
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(responseBody, &resp); err == nil {
			return resp.Usage.CompletionTokens, nil
		}
		return len(responseBody) / 4, nil
	}

	respStr := string(responseBody)
	totalTokens := 0
	lines := strings.Split(respStr, "\n")
	for _, line := range lines {
		_, tokens, usage, err := p.ParseStreamChunk(model, []byte(line))
		if err == nil {
			if usage != nil {
				return usage.OutputTokens, nil
			}
			totalTokens += tokens
		}
	}
	return totalTokens, nil
}

func (p *MockProvider) ParseStreamChunk(model domain.Model, chunk []byte) (string, int, *domain.Usage, error) {
	chunk = bytes.TrimSpace(chunk)
	if !bytes.HasPrefix(chunk, []byte("data: ")) || bytes.HasSuffix(chunk, []byte("[DONE]")) {
		return "", 0, nil, nil
	}
	dataBytes := bytes.TrimPrefix(chunk, []byte("data: "))
	var streamChunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(dataBytes, &streamChunk); err != nil {
		return "", 0, nil, err
	}

	content := ""
	if len(streamChunk.Choices) > 0 {
		content = streamChunk.Choices[0].Delta.Content
	}

	var usage *domain.Usage
	if streamChunk.Usage != nil {
		usage = &domain.Usage{
			InputTokens:  streamChunk.Usage.PromptTokens,
			OutputTokens: streamChunk.Usage.CompletionTokens,
			TotalCost:    0,
		}
	}

	return content, len(content) / 4, usage, nil
}

func (p *MockProvider) ParseRequest(body []byte) (domain.Model, bool, error) {
	var req struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return "", false, err
	}
	return domain.Model(req.Model), req.Stream, nil
}

func (p *MockProvider) GetUsage(ctx context.Context) (float64, error) {
	return 0, nil
}

// handleRequest mocks the OpenAI API response
func (p *MockProvider) handleRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model    string `json:"model"`
		Stream   bool   `json:"stream"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	responseContent := "This is a mock response from the Pouch AI Mock Provider."
	p.mu.RLock()
	if customResp, ok := p.config["mock_response"]; ok && customResp != "" {
		// If it's a valid JSON response from the user, we try to use it
		// For the simplest case, we just use it as the content of a successful response
		responseContent = customResp
	}
	p.mu.RUnlock()

	if len(req.Messages) > 0 && responseContent == "This is a mock response from the Pouch AI Mock Provider." {
		lastMsg := req.Messages[len(req.Messages)-1]
		responseContent = fmt.Sprintf("Mock response to: %q", lastMsg.Content)
	}

	if !req.Stream {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":      "mock-response-id",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   req.Model,
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": responseContent,
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     len(req.Messages) * 5, // Approximate
				"completion_tokens": len(responseContent) / 4,
				"total_tokens":      (len(req.Messages) * 5) + (len(responseContent) / 4),
			},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Streaming response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Split content into chunks to simulate streaming
	words := strings.Split(responseContent, " ")
	for i, word := range words {
		chunk := map[string]interface{}{
			"id":      "mock-response-id",
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   req.Model,
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"delta": map[string]interface{}{
						"content": word + " ",
					},
					"finish_reason": nil,
				},
			},
		}

		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		// Simulate latency
		time.Sleep(50 * time.Millisecond)

		// Last chunk validation logic not strictly needed for mock, but we can set finish_reason on a final empty chunk if we wanted.
		if i == len(words)-1 {
			// Send a finish chunk
			finishChunk := map[string]interface{}{
				"id":      "mock-response-id",
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   req.Model,
				"choices": []map[string]interface{}{
					{
						"index":         0,
						"delta":         map[string]interface{}{},
						"finish_reason": "stop",
					},
				},
			}
			data, _ = json.Marshal(finishChunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}
