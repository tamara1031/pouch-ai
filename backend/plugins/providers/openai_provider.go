package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"pouch-ai/backend/config"
	"pouch-ai/backend/domain"
	"strings"
	"time"
)

type TokenCounter interface {
	Count(model string, text string) (int, error)
}

type OpenAIProvider struct {
	pricing      *OpenAIPricing
	tokenCounter TokenCounter
	apiKey       string
	baseURL      string
}

type OpenAIBuilder struct{}

func (b *OpenAIBuilder) Build(ctx context.Context, cfg *config.Config) (domain.Provider, error) {
	// Priority: Env > Flag > Default
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = cfg.OpenAIKey
	}

	apiURL := os.Getenv("OPENAI_URL")
	if apiURL == "" {
		apiURL = os.Getenv("OPENAI_API_URL")
	}
	if apiURL == "" {
		apiURL = cfg.OpenAIURL
	}

	if apiKey == "" {
		fmt.Println("WARN: OpenAI API Key not found. 'openai' provider will be unavailable.")
		return nil, nil
	}

	pricing, err := NewOpenAIPricing()
	if err != nil {
		return nil, fmt.Errorf("failed to load openai pricing: %w", err)
	}
	tokenCounter := NewTiktokenCounter()

	return NewOpenAIProvider(apiKey, apiURL, pricing, tokenCounter), nil
}

func NewOpenAIProvider(apiKey string, baseURL string, pricing *OpenAIPricing, counter TokenCounter) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	// Ensure no trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &OpenAIProvider{
		apiKey:       apiKey,
		baseURL:      baseURL,
		pricing:      pricing,
		tokenCounter: counter,
	}
}

func (p *OpenAIProvider) Schema() domain.PluginSchema {
	return domain.PluginSchema{
		"api_key": {
			Type:        domain.FieldTypeString,
			DisplayName: "API Key",
			Description: "Your OpenAI API Key",
		},
		"base_url": {
			Type:        domain.FieldTypeString,
			DisplayName: "Base URL",
			Default:     "https://api.openai.com/v1",
			Description: "OpenAI API Base URL",
		},
	}
}

func (p *OpenAIProvider) Configure(config map[string]any) (domain.Provider, error) {
	newP := *p
	if val, ok := config["api_key"]; ok {
		if s, ok := val.(string); ok {
			newP.apiKey = s
		}
	}
	if val, ok := config["base_url"]; ok {
		if s, ok := val.(string); ok {
			newP.baseURL = strings.TrimSuffix(s, "/")
		}
	}
	return &newP, nil
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) GetPricing(model domain.Model) (domain.Pricing, error) {
	mp, err := p.pricing.GetPrice(string(model))
	if err != nil {
		return domain.Pricing{}, err
	}
	return domain.Pricing{
		Input:  mp.Input,
		Output: mp.Output,
	}, nil
}

func (p *OpenAIProvider) CountTokens(model domain.Model, text string) (int, error) {
	return p.tokenCounter.Count(string(model), text)
}

func (p *OpenAIProvider) PrepareHTTPRequest(ctx context.Context, model domain.Model, body []byte) (*http.Request, error) {
	// Inject stream_options: {include_usage: true} if streaming
	var reqMap map[string]any
	if err := json.Unmarshal(body, &reqMap); err == nil {
		if stream, ok := reqMap["stream"].(bool); stream && ok {
			if _, ok := reqMap["stream_options"]; !ok {
				reqMap["stream_options"] = map[string]any{"include_usage": true}
				body, _ = json.Marshal(reqMap)
			}
		}
	}

	url := p.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	return req, nil
}

func (p *OpenAIProvider) EstimateUsage(model domain.Model, body []byte) (*domain.Usage, error) {
	var req struct {
		Model    string `json:"model"`
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	var inputBuilder strings.Builder
	for _, m := range req.Messages {
		inputBuilder.WriteString(m.Content)
	}

	inputTokens, err := p.CountTokens(model, inputBuilder.String())
	if err != nil {
		return nil, err
	}

	pricing, err := p.GetPricing(model)
	if err != nil {
		return nil, err
	}

	return &domain.Usage{
		InputTokens: inputTokens,
		TotalCost:   float64(inputTokens) / 1000.0 * pricing.Input,
	}, nil
}

func (p *OpenAIProvider) ParseOutputUsage(model domain.Model, responseBody []byte, isStream bool) (int, error) {
	respStr := string(responseBody)

	if !isStream {
		var resp struct {
			Usage struct {
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(responseBody, &resp); err == nil && resp.Usage.CompletionTokens > 0 {
			return resp.Usage.CompletionTokens, nil
		}
	} else {
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

	// Fallback
	return len(respStr) / 4, nil
}

func (p *OpenAIProvider) ParseStreamChunk(model domain.Model, chunk []byte) (string, int, *domain.Usage, error) {
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
			TotalTokens      int `json:"total_tokens"`
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
		pricing, _ := p.GetPricing(model)
		usage = &domain.Usage{
			InputTokens:  streamChunk.Usage.PromptTokens,
			OutputTokens: streamChunk.Usage.CompletionTokens,
			TotalCost:    (float64(streamChunk.Usage.PromptTokens) / 1000.0 * pricing.Input) + (float64(streamChunk.Usage.CompletionTokens) / 1000.0 * pricing.Output),
		}
	}

	tokens := 0
	if content != "" {
		tokens, _ = p.CountTokens(model, content)
	}

	return content, tokens, usage, nil
}

func (p *OpenAIProvider) ParseRequest(body []byte) (domain.Model, bool, error) {
	var req struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return "", false, err
	}
	return domain.Model(req.Model), req.Stream, nil
}

func (p *OpenAIProvider) GetUsage(ctx context.Context) (float64, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	end := now.Format("2006-01-02")

	url := fmt.Sprintf("%s/dashboard/billing/usage?start_date=%s&end_date=%s", p.baseURL, start, end)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("openai usage api returned status: %d", resp.StatusCode)
	}

	var data struct {
		TotalUsage float64 `json:"total_usage"` // in cents
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	return data.TotalUsage / 100.0, nil // Convert cents to dollars
}
