package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pouch-ai/backend/api"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/infra/engine"
	"pouch-ai/backend/plugins/providers"
	"pouch-ai/backend/service"

	"github.com/labstack/echo/v4"
)

func TestProxy_PassThrough(t *testing.T) {
	// 1. Setup Mock Upstream Server
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "Hello world"}}]}`))
	}))
	defer mockUpstream.Close()

	// 2. Setup Dependencies
	// Pricing & TokenCounter (Mock or Real)
	pricing, err := providers.NewOpenAIPricing() // Using real implementation with embedded json
	if err != nil {
		t.Fatalf("Failed to create pricing: %v", err)
	}
	tokenCounter := providers.NewTiktokenCounter()

	// Registry
	registry := domain.NewProviderRegistry()
	provider := providers.NewOpenAIProvider("test-key", mockUpstream.URL, pricing, tokenCounter)
	registry.Register(provider.Name(), provider)

	// Service
	executionHandler := engine.NewExecutionHandler(nil)
	mwRegistry := domain.NewMiddlewareRegistry()
	keyService := service.NewKeyService(&MockRepository{}, registry, mwRegistry)
	proxyService := service.NewProxyService(executionHandler, mwRegistry, keyService)

	// Handler
	handler := api.NewProxyHandler(proxyService, registry)

	// 3. Setup Echo
	e := echo.New()
	reqBody := `{"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock App Key
	appKey := &domain.Key{
		Configuration: &domain.KeyConfiguration{
			Provider: domain.PluginConfig{ID: "openai"},
		},
	}
	c.Set("app_key", appKey)

	// 4. Execute
	if err := handler.Proxy(c); err != nil {
		t.Fatalf("handler.Proxy failed: %v", err)
	}

	// 5. Assertions
	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got '%s'", contentType)
	}

	customHeader := resp.Header.Get("X-Custom-Header")
	if customHeader != "custom-value" {
		t.Errorf("expected X-Custom-Header custom-value, got '%s'", customHeader)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	if !strings.Contains(bodyString, "Hello world") {
		t.Errorf("expected body to contain 'Hello world', got %s", bodyString)
	}
}
