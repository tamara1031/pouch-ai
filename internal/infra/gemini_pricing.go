package infra

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed gemini_pricing.json
var geminiPricingJSON []byte

type GeminiModelPrice struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type GeminiPricing struct {
	prices map[string]GeminiModelPrice
	mu     sync.RWMutex
}

func NewGeminiPricing() (*GeminiPricing, error) {
	var prices map[string]GeminiModelPrice
	if err := json.Unmarshal(geminiPricingJSON, &prices); err != nil {
		return nil, fmt.Errorf("failed to parse gemini_pricing.json: %w", err)
	}
	return &GeminiPricing{prices: prices}, nil
}

func (p *GeminiPricing) GetPrice(model string) (GeminiModelPrice, error) {
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

	return GeminiModelPrice{}, fmt.Errorf("price not found for model: %s", model)
}
