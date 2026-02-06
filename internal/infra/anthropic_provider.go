package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pouch-ai/internal/domain"
	"strings"
)

type AnthropicProvider struct {
	pricing      *AnthropicPricing
	tokenCounter TokenCounter
	apiKey       string
	baseURL      string
}

func NewAnthropicProvider(apiKey string, pricing *AnthropicPricing, counter TokenCounter) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey:       apiKey,
		baseURL:      "https://api.anthropic.com/v1",
		pricing:      pricing,
		tokenCounter: counter,
	}
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

func (p *AnthropicProvider) GetPricing(model domain.Model) (domain.Pricing, error) {
	mp, err := p.pricing.GetPrice(string(model))
	if err != nil {
		return domain.Pricing{}, err
	}
	return domain.Pricing{
		Input:  mp.Input,
		Output: mp.Output,
	}, nil
}

func (p *AnthropicProvider) CountTokens(model domain.Model, text string) (int, error) {
	return p.tokenCounter.Count(string(model), text)
}

func (p *AnthropicProvider) ParseRequest(body []byte) (domain.Model, bool, error) {
	var req struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return "", false, err
	}
	return domain.Model(req.Model), req.Stream, nil
}

func (p *AnthropicProvider) PrepareHTTPRequest(ctx context.Context, model domain.Model, body []byte) (*http.Request, error) {
	var oReq struct {
		Messages []struct {
			Role    string `json:"role"`
			Content any    `json:"content"`
		} `json:"messages"`
		MaxTokens   int      `json:"max_tokens,omitempty"`
		Stream      bool     `json:"stream"`
		Temperature *float64 `json:"temperature,omitempty"`
	}
	if err := json.Unmarshal(body, &oReq); err != nil {
		return nil, err
	}

	var aReq struct {
		Model       string `json:"model"`
		Messages    []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		System      string   `json:"system,omitempty"`
		MaxTokens   int      `json:"max_tokens"`
		Stream      bool     `json:"stream,omitempty"`
		Temperature *float64 `json:"temperature,omitempty"`
	}

	aReq.Model = string(model)
	aReq.Stream = oReq.Stream
	aReq.Temperature = oReq.Temperature
	aReq.MaxTokens = oReq.MaxTokens
	if aReq.MaxTokens == 0 {
		aReq.MaxTokens = 4096 // Default for Anthropic
	}

	for _, m := range oReq.Messages {
		contentStr := ""
		switch v := m.Content.(type) {
		case string:
			contentStr = v
		default:
			// Fallback for array/objects (not supported yet)
			contentStr = fmt.Sprintf("%v", v)
		}

		if m.Role == "system" {
			if aReq.System != "" {
				aReq.System += "\n"
			}
			aReq.System += contentStr
		} else {
			aReq.Messages = append(aReq.Messages, struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				Role:    m.Role,
				Content: contentStr,
			})
		}
	}

	aBody, err := json.Marshal(aReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewBuffer(aBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	return req, nil
}

func (p *AnthropicProvider) EstimateUsage(model domain.Model, body []byte) (*domain.Usage, error) {
	// Re-parse body to extract text. Note: body passed here is original OpenAI body.
	var req struct {
		Messages []struct {
			Content any `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	var inputBuilder strings.Builder
	for _, m := range req.Messages {
		switch v := m.Content.(type) {
		case string:
			inputBuilder.WriteString(v)
		}
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

func (p *AnthropicProvider) ParseOutputUsage(model domain.Model, responseBody []byte, isStream bool) (int, error) {
	// IMPORTANT: ExecutionHandler calls this with the RAW Provider Response (if not streaming).
	// But if we TransformResponse, ExecutionHandler reads that.
	// Wait, in execution.go:
	// 3. For non-streaming...
	//    body, err := io.ReadAll(resp.Body)
	//    ... ParseOutputUsage(..., body, false)
	//    ... TransformResponse(bytes.NewBuffer(body))
	// So ParseOutputUsage receives the ANTHROPIC response body.

	if isStream {
		// Not used by ExecutionHandler currently for streaming.
		return 0, nil
	}

	var resp struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(responseBody, &resp); err == nil {
		return resp.Usage.OutputTokens, nil
	}
	return 0, nil
}

func (p *AnthropicProvider) TransformResponse(body io.Reader, isStream bool) (io.Reader, error) {
	if !isStream {
		// Parse Anthropic JSON
		var aResp struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.NewDecoder(body).Decode(&aResp); err != nil {
			return nil, err
		}

		fullText := ""
		for _, c := range aResp.Content {
			fullText += c.Text
		}

		// Create OpenAI JSON
		oResp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": fullText,
					},
					"finish_reason": "stop",
				},
			},
		}
		b, _ := json.Marshal(oResp)
		return bytes.NewBuffer(b), nil
	}

	// Streaming
	transformer := func(evt ServerSentEvent) ([]ServerSentEvent, error) {
		if evt.Event == "message_stop" {
			return []ServerSentEvent{{Data: "[DONE]"}}, nil
		}

		if evt.Event == "content_block_delta" {
			var data struct {
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(evt.Data), &data); err != nil {
				return nil, err
			}

			oData := map[string]any{
				"choices": []map[string]any{
					{
						"delta": map[string]any{
							"content": data.Delta.Text,
						},
					},
				},
			}
			b, _ := json.Marshal(oData)
			return []ServerSentEvent{{Data: string(b)}}, nil
		}
		// Ignore other events
		return nil, nil
	}

	return NewSSETransformer(body, transformer), nil
}

func (p *AnthropicProvider) GetUsage(ctx context.Context) (float64, error) {
	// Anthropic doesn't have a simple usage API like OpenAI
	return 0, nil
}
