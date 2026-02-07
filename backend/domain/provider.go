package domain

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"pouch-ai/backend/config"
)

// ErrProviderNotFound is returned when a requested provider is not registered.
var ErrProviderNotFound = errors.New("provider not found")

// Model represents the LLM model identifier (e.g., "gpt-4o").
type Model string

func (m Model) String() string {
	return string(m)
}

// Pricing defines the cost per 1M tokens (usually) or per unit.
// Note: Implementation details depend on the provider's unit.
type Pricing struct {
	Input  float64
	Output float64
}

// Usage captures the token usage and calculated cost for a request.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalCost    float64
}

// ProviderInfo contains metadata about a registered provider.
type ProviderInfo struct {
	ID     string       `json:"id"`
	Schema PluginSchema `json:"schema"`
}

// Provider defines the interface for LLM backends.
// It handles model-specific logic like token counting, request preparation, and response parsing.
type Provider interface {
	// Name returns the unique identifier of the provider (e.g., "openai").
	Name() string
	// Schema returns the configuration schema for this provider.
	Schema() PluginSchema
	// Configure returns a new instance of the provider with the given configuration.
	Configure(config map[string]any) (Provider, error)

	// GetPricing returns the pricing structure for the given model.
	GetPricing(model Model) (Pricing, error)
	// CountTokens estimates the number of tokens in the given text.
	CountTokens(model Model, text string) (int, error)

	// PrepareHTTPRequest creates a ready-to-send HTTP request for the provider's API.
	PrepareHTTPRequest(ctx context.Context, model Model, body []byte) (*http.Request, error)

	// EstimateUsage calculates the estimated input cost before execution.
	EstimateUsage(model Model, requestBody []byte) (*Usage, error)
	// ParseOutputUsage extracts usage statistics from the response body.
	ParseOutputUsage(model Model, responseBody []byte, isStream bool) (int, error)
	// ProcessStreamChunk extracts the content delta from a single stream event.
	ProcessStreamChunk(chunk []byte) (string, error)
	// ParseRequest extracts generic model and stream info from the provider-specific request body.
	ParseRequest(body []byte) (Model, bool, error)

	// GetUsage returns the total usage cost from the provider side (e.g. external billing API).
	GetUsage(ctx context.Context) (float64, error)
}

// ProviderRegistry manages the registration and retrieval of Providers.
type ProviderRegistry interface {
	Register(p Provider)
	Get(name string) (Provider, error)
	List() []Provider
	ListInfo() []ProviderInfo
}

// ProviderBuilder constructs a Provider instance from the global configuration.
type ProviderBuilder interface {
	Build(ctx context.Context, cfg *config.Config) (Provider, error)
}

// DefaultProviderRegistry is the standard implementation of ProviderRegistry.
type DefaultProviderRegistry struct {
	providers map[string]Provider
}

// NewProviderRegistry creates a new empty DefaultProviderRegistry.
func NewProviderRegistry() ProviderRegistry {
	return &DefaultProviderRegistry{providers: make(map[string]Provider)}
}

// Register adds a provider to the registry.
func (r *DefaultProviderRegistry) Register(p Provider) {
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name.
func (r *DefaultProviderRegistry) Get(name string) (Provider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}
	return p, nil
}

// List returns all registered providers.
func (r *DefaultProviderRegistry) List() []Provider {
	var providers []Provider
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}

// ListInfo returns metadata for all registered providers.
func (r *DefaultProviderRegistry) ListInfo() []ProviderInfo {
	var infos []ProviderInfo
	for _, p := range r.providers {
		infos = append(infos, ProviderInfo{
			ID:     p.Name(),
			Schema: p.Schema(),
		})
	}
	return infos
}
