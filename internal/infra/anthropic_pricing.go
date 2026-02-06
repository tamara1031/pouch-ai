package infra

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed anthropic_pricing.json
var anthropicPricingJSON []byte

type AnthropicModelPrice struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type AnthropicPricing struct {
	prices map[string]AnthropicModelPrice
	mu     sync.RWMutex
}

func NewAnthropicPricing() (*AnthropicPricing, error) {
	var prices map[string]AnthropicModelPrice
	if err := json.Unmarshal(anthropicPricingJSON, &prices); err != nil {
		return nil, fmt.Errorf("failed to parse anthropic_pricing.json: %w", err)
	}
	return &AnthropicPricing{prices: prices}, nil
}

func (p *AnthropicPricing) GetPrice(model string) (AnthropicModelPrice, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Direct match
	if price, ok := p.prices[model]; ok {
		return price, nil
	}

	// Prefix match
	for k, v := range p.prices {
		if strings.HasPrefix(model, k) {
			return v, nil
		}
	}

	return AnthropicModelPrice{}, fmt.Errorf("price not found for model: %s", model)
}
