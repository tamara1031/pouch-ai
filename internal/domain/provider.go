package domain

import (
	"context"
	"fmt"
	"net/http"
)

type Model string

func (m Model) String() string {
	return string(m)
}

type Pricing struct {
	Input  float64
	Output float64
}

type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalCost    float64
}

type Provider interface {
	Name() string
	GetPricing(model Model) (Pricing, error)
	CountTokens(model Model, text string) (int, error)
	PrepareHTTPRequest(ctx context.Context, model Model, body []byte) (*http.Request, error)

	// DDD: The provider is responsible for knowing how to estimate its own cost
	EstimateUsage(model Model, requestBody []byte) (*Usage, error)
	// Output tokens often come from the response body (JSON usage or stream parsing)
	ParseOutputUsage(model Model, responseBody []byte, isStream bool) (int, error)
	// ParseRequest extracts generic info from provider-specific request body
	ParseRequest(body []byte) (Model, bool, error)

	// GetUsage returns the total usage cost from the provider side (e.g. billing)
	GetUsage(ctx context.Context) (float64, error)
}

type Registry interface {
	Register(p Provider)
	Get(name string) (Provider, error)
	List() []Provider
}

type DefaultRegistry struct {
	providers map[string]Provider
}

func NewRegistry() Registry {
	return &DefaultRegistry{providers: make(map[string]Provider)}
}

func (r *DefaultRegistry) Register(p Provider) {
	r.providers[p.Name()] = p
}

func (r *DefaultRegistry) Get(name string) (Provider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return p, nil
}

func (r *DefaultRegistry) List() []Provider {
	var providers []Provider
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}
