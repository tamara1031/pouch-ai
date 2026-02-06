package openai

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed pricing.json
var pricingJSON []byte

type ModelPrice struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type Pricing struct {
	prices map[string]ModelPrice
	mu     sync.RWMutex
}

func NewPricing() (*Pricing, error) {
	var prices map[string]ModelPrice
	if err := json.Unmarshal(pricingJSON, &prices); err != nil {
		return nil, fmt.Errorf("failed to parse pricing.json: %w", err)
	}
	return &Pricing{prices: prices}, nil
}

func (p *Pricing) GetPrice(model string) (ModelPrice, error) {
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

	return ModelPrice{}, fmt.Errorf("price not found for model: %s", model)
}
