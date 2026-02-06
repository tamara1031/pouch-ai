package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"pouch-ai/internal/domain/provider"
	"strings"
)

type TokenCounter interface {
	Count(model string, text string) (int, error)
}

type OpenAIProvider struct {
	pricing      *Pricing
	tokenCounter TokenCounter
	apiKey       string
	baseURL      string
}

func NewOpenAIProvider(apiKey string, baseURL string, pricing *Pricing, counter TokenCounter) *OpenAIProvider {
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

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) GetPricing(model provider.Model) (provider.Pricing, error) {
	mp, err := p.pricing.GetPrice(string(model))
	if err != nil {
		return provider.Pricing{}, err
	}
	return provider.Pricing{
		Input:  mp.Input,
		Output: mp.Output,
	}, nil
}

func (p *OpenAIProvider) CountTokens(model provider.Model, text string) (int, error) {
	return p.tokenCounter.Count(string(model), text)
}

func (p *OpenAIProvider) PrepareHTTPRequest(ctx context.Context, model provider.Model, body []byte) (*http.Request, error) {
	url := p.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	return req, nil
}

func (p *OpenAIProvider) EstimateUsage(model provider.Model, body []byte) (*provider.Usage, error) {
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

	return &provider.Usage{
		InputTokens: inputTokens,
		TotalCost:   float64(inputTokens) / 1000.0 * pricing.Input,
	}, nil
}

func (p *OpenAIProvider) ParseOutputUsage(model provider.Model, responseBody []byte, isStream bool) (int, error) {
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
		var fullContent strings.Builder
		lines := strings.Split(respStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") || strings.HasSuffix(line, "[DONE]") {
				continue
			}
			dataStr := strings.TrimPrefix(line, "data: ")
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(dataStr), &chunk); err == nil {
				if len(chunk.Choices) > 0 {
					fullContent.WriteString(chunk.Choices[0].Delta.Content)
				}
			}
		}
		finalText := fullContent.String()
		if len(finalText) > 0 {
			return p.CountTokens(model, finalText)
		}
	}

	// Fallback
	return len(respStr) / 4, nil
}

func (p *OpenAIProvider) ParseRequest(body []byte) (provider.Model, bool, error) {
	var req struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return "", false, err
	}
	return provider.Model(req.Model), req.Stream, nil
}
