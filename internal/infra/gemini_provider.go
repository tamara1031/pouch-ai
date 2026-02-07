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

type GeminiProvider struct {
	pricing      *GeminiPricing
	tokenCounter TokenCounter
	apiKey       string
	baseURL      string
}

func NewGeminiProvider(apiKey string, pricing *GeminiPricing, counter TokenCounter) *GeminiProvider {
	return &GeminiProvider{
		apiKey:       apiKey,
		baseURL:      "https://generativelanguage.googleapis.com/v1beta/models",
		pricing:      pricing,
		tokenCounter: counter,
	}
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) GetPricing(model domain.Model) (domain.Pricing, error) {
	mp, err := p.pricing.GetPrice(string(model))
	if err != nil {
		return domain.Pricing{}, err
	}
	return domain.Pricing{
		Input:  mp.Input,
		Output: mp.Output,
	}, nil
}

func (p *GeminiProvider) CountTokens(model domain.Model, text string) (int, error) {
	return p.tokenCounter.Count(string(model), text)
}

func (p *GeminiProvider) ParseRequest(body []byte) (domain.Model, bool, error) {
	var req struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return "", false, err
	}
	return domain.Model(req.Model), req.Stream, nil
}

func (p *GeminiProvider) PrepareHTTPRequest(ctx context.Context, model domain.Model, body []byte) (*http.Request, error) {
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

	// Gemini Request Structure
	type Part struct {
		Text string `json:"text,omitempty"`
	}
	type Content struct {
		Role  string `json:"role"`
		Parts []Part `json:"parts"`
	}
	type GeminiRequest struct {
		Contents          []Content `json:"contents"`
		SystemInstruction *Content  `json:"system_instruction,omitempty"`
		GenerationConfig  struct {
			MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
			Temperature     *float64 `json:"temperature,omitempty"`
		} `json:"generationConfig,omitempty"`
	}

	var gReq GeminiRequest
	gReq.GenerationConfig.MaxOutputTokens = oReq.MaxTokens
	gReq.GenerationConfig.Temperature = oReq.Temperature

	for _, m := range oReq.Messages {
		contentStr := ""
		switch v := m.Content.(type) {
		case string:
			contentStr = v
		default:
			contentStr = fmt.Sprintf("%v", v)
		}

		if m.Role == "system" {
			gReq.SystemInstruction = &Content{
				Role:  "user", // System instructions in Gemini don't strictly have a role, or use 'user' context usually
				Parts: []Part{{Text: contentStr}},
			}
		} else {
			role := "user"
			if m.Role == "assistant" {
				role = "model"
			}
			gReq.Contents = append(gReq.Contents, Content{
				Role:  role,
				Parts: []Part{{Text: contentStr}},
			})
		}
	}

	gBody, err := json.Marshal(gReq)
	if err != nil {
		return nil, err
	}

	method := "generateContent"
	if oReq.Stream {
		method = "streamGenerateContent"
	}
	url := fmt.Sprintf("%s/%s:%s?key=%s", p.baseURL, model, method, p.apiKey)
	if oReq.Stream {
		url += "&alt=sse" // Gemini supports SSE with alt=sse! This simplifies things immensely.
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(gBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (p *GeminiProvider) EstimateUsage(model domain.Model, body []byte) (*domain.Usage, error) {
	// Re-parse body to extract text
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

func (p *GeminiProvider) ParseOutputUsage(model domain.Model, responseBody []byte, isStream bool) (int, error) {
	if !isStream {
		var resp struct {
			UsageMetadata struct {
				CandidatesTokenCount int `json:"candidatesTokenCount"`
			} `json:"usageMetadata"`
		}
		if err := json.Unmarshal(responseBody, &resp); err == nil {
			return resp.UsageMetadata.CandidatesTokenCount, nil
		}
		return 0, nil
	}

	// Streaming logic
	maxTokens := 0
	lines := bytes.Split(responseBody, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}
		data := bytes.TrimPrefix(line, []byte("data: "))

		if !bytes.Contains(data, []byte("usageMetadata")) {
			continue
		}

		var evt struct {
			UsageMetadata struct {
				CandidatesTokenCount int `json:"candidatesTokenCount"`
			} `json:"usageMetadata"`
		}
		if err := json.Unmarshal(data, &evt); err == nil {
			if evt.UsageMetadata.CandidatesTokenCount > maxTokens {
				maxTokens = evt.UsageMetadata.CandidatesTokenCount
			}
		}
	}

	if maxTokens > 0 {
		return maxTokens, nil
	}

	return 0, nil
}

func (p *GeminiProvider) ProcessStreamChunk(chunk []byte) (string, error) {
	chunk = bytes.TrimSpace(chunk)
	if !bytes.HasPrefix(chunk, []byte("data: ")) {
		return "", nil
	}
	data := bytes.TrimPrefix(chunk, []byte("data: "))
	var gChunk struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &gChunk); err == nil {
		if len(gChunk.Candidates) > 0 {
			text := ""
			for _, part := range gChunk.Candidates[0].Content.Parts {
				text += part.Text
			}
			return text, nil
		}
	}
	return "", nil
}

func (p *GeminiProvider) TransformResponse(body io.Reader, isStream bool) (io.Reader, error) {
	if !isStream {
		var gResp struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
				FinishReason string `json:"finishReason"`
			} `json:"candidates"`
		}
		if err := json.NewDecoder(body).Decode(&gResp); err != nil {
			return nil, err
		}

		fullText := ""
		if len(gResp.Candidates) > 0 {
			for _, part := range gResp.Candidates[0].Content.Parts {
				fullText += part.Text
			}
		}

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

	// With alt=sse, Gemini sends:
	// data: {"candidates": [...]}
	transformer := func(evt ServerSentEvent) ([]ServerSentEvent, error) {
		if evt.Data == "" {
			return nil, nil
		}

		var gChunk struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(evt.Data), &gChunk); err != nil {
			return nil, err
		}

		if len(gChunk.Candidates) > 0 {
			text := ""
			for _, part := range gChunk.Candidates[0].Content.Parts {
				text += part.Text
			}

			// Only send if there is text (avoid empty deltas from metadata updates)
			if text != "" {
				oData := map[string]any{
					"choices": []map[string]any{
						{
							"delta": map[string]any{
								"content": text,
							},
						},
					},
				}
				b, _ := json.Marshal(oData)
				return []ServerSentEvent{{Data: string(b)}}, nil
			}
		}

		return nil, nil
	}

	return NewSSETransformer(body, transformer), nil
}

func (p *GeminiProvider) GetUsage(ctx context.Context) (float64, error) {
	return 0, nil
}
