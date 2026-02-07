package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pouch-ai/internal/domain"
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
	url := p.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
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
		prefixData := []byte("data: ")
		suffixDone := []byte("[DONE]")

		lines := bytes.Split(responseBody, []byte("\n"))
		for _, line := range lines {
			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, prefixData) || bytes.HasSuffix(line, suffixDone) {
				continue
			}
			data := bytes.TrimPrefix(line, prefixData)
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal(data, &chunk); err == nil {
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
	return len(responseBody) / 4, nil
}

func (p *OpenAIProvider) ProcessStreamChunk(chunk []byte) (string, error) {
	chunk = bytes.TrimSpace(chunk)
	if !bytes.HasPrefix(chunk, []byte("data: ")) || bytes.HasSuffix(chunk, []byte("[DONE]")) {
		return "", nil
	}
	dataBytes := bytes.TrimPrefix(chunk, []byte("data: "))
	var streamChunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(dataBytes, &streamChunk); err != nil {
		return "", err
	}
	if len(streamChunk.Choices) > 0 {
		return streamChunk.Choices[0].Delta.Content, nil
	}
	return "", nil
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
