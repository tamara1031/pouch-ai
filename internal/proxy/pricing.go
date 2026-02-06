package proxy

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

	// Direct match
	if price, ok := p.prices[model]; ok {
		return price, nil
	}

	// Prefix match for snapshots (e.g. gpt-4-0613 -> gpt-4)
	// This is a naive implementation, can be improved.
    // For now, let's just return error or fallback.
    // Let's try to match keys that `model` starts with.
    // Actually, usually pricing is specific. Let's just return error for strictness or default to something safe.
    // Or better, check if we have a "gpt-4" key and the requested model is "gpt-4-0314".
    
    // Quick fallback loop
    for k, v := range p.prices {
        if strings.HasPrefix(model, k) {
             return v, nil
        }
    }

	return ModelPrice{}, fmt.Errorf("price not found for model: %s", model)
}
