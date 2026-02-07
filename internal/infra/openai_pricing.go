package infra

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

//go:embed openai_pricing.json
var pricingJSON []byte

type OpenAIModelPrice struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type pricingEntry struct {
	prefix string
	price  OpenAIModelPrice
}

type OpenAIPricing struct {
	prices        map[string]OpenAIModelPrice
	sortedEntries []pricingEntry
	mu            sync.RWMutex
}

func NewOpenAIPricing() (*OpenAIPricing, error) {
	var prices map[string]OpenAIModelPrice
	if err := json.Unmarshal(pricingJSON, &prices); err != nil {
		return nil, fmt.Errorf("failed to parse pricing.json: %w", err)
	}

	var entries []pricingEntry
	for k, v := range prices {
		entries = append(entries, pricingEntry{prefix: k, price: v})
	}

	// Sort by length descending, then lexicographically for stability
	sort.Slice(entries, func(i, j int) bool {
		if len(entries[i].prefix) != len(entries[j].prefix) {
			return len(entries[i].prefix) > len(entries[j].prefix)
		}
		return entries[i].prefix < entries[j].prefix
	})

	return &OpenAIPricing{
		prices:        prices,
		sortedEntries: entries,
	}, nil
}

func (p *OpenAIPricing) GetPrice(model string) (OpenAIModelPrice, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if price, ok := p.prices[model]; ok {
		return price, nil
	}

	for _, entry := range p.sortedEntries {
		if strings.HasPrefix(model, entry.prefix) {
			return entry.price, nil
		}
	}

	return OpenAIModelPrice{}, fmt.Errorf("price not found for model: %s", model)
}
