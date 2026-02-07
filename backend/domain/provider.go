package domain

import (
	"context"
	"net/http"
	"pouch-ai/backend/config"
)

type ProviderBuilder interface {
	Build(ctx context.Context, cfg *config.Config) (Provider, error)
}

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
	Schema() PluginSchema
	Configure(config map[string]any) (Provider, error)
	GetPricing(model Model) (Pricing, error)
	CountTokens(model Model, text string) (int, error)
	PrepareHTTPRequest(ctx context.Context, model Model, body []byte) (*http.Request, error)

	// DDD: The provider is responsible for knowing how to estimate its own cost
	EstimateUsage(model Model, requestBody []byte) (*Usage, error)
	// Output tokens often come from the response body (JSON usage or stream parsing)
	ParseOutputUsage(model Model, responseBody []byte, isStream bool) (int, error)
	// ParseStreamChunk extracts content, token count, and usage from a single stream chunk
	ParseStreamChunk(model Model, chunk []byte) (string, int, *Usage, error)
	// ParseRequest extracts generic info from provider-specific request body
	ParseRequest(body []byte) (Model, bool, error)

	// GetUsage returns the total usage cost from the provider side (e.g. billing)
	GetUsage(ctx context.Context) (float64, error)
}
