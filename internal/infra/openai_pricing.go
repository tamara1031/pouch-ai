package infra

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed openai_pricing.json
var pricingJSON []byte

type OpenAIModelPrice struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type OpenAIPricing struct {
	prices map[string]OpenAIModelPrice
	mu     sync.RWMutex
}

func NewOpenAIPricing() (*OpenAIPricing, error) {
	var prices map[string]OpenAIModelPrice
	if err := json.Unmarshal(pricingJSON, &prices); err != nil {
		return nil, fmt.Errorf("failed to parse pricing.json: %w", err)
	}
	return &OpenAIPricing{prices: prices}, nil
}

func (p *OpenAIPricing) GetPrice(model string) (OpenAIModelPrice, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if price, ok := p.prices[model]; ok {
		return price, nil
	}

	for k, v := range p.prices {
		if strings.HasPrefix(model, k) {
			return v, nil
		}
	}

	return OpenAIModelPrice{}, fmt.Errorf("price not found for model: %s", model)
}
