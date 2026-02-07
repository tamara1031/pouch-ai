package api_test

import (
	"encoding/json"
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

func TestMockProvider_Integration(t *testing.T) {
	// 1. Setup Dependencies
	registry := domain.NewProviderRegistry()
	mwRegistry := domain.NewMiddlewareRegistry()
	repo := &MockRepository{}
	keyService := service.NewKeyService(repo, registry, mwRegistry)

	// Register Mock Provider
	mockProv := providers.NewMockProvider()
	registry.Register(mockProv.Name(), mockProv)

	// Service
	executionHandler := engine.NewExecutionHandler(repo)

	proxyService := service.NewProxyService(
		executionHandler,
		mwRegistry,
		keyService,
	)

	// Handler
	handler := api.NewProxyHandler(proxyService, registry)

	// 2. Setup Echo
	e := echo.New()
	reqBody := `{"model": "mock-gpt-4", "messages": [{"role": "user", "content": "Hello Mock"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Inject Mock App Key
	appKey := &domain.Key{
		ID: 1, // Dummy ID
		Configuration: &domain.KeyConfiguration{
			Provider: domain.PluginConfig{ID: "mock"},
		},
	}
	c.Set("app_key", appKey)

	// 3. Execute
	if err := handler.Proxy(c); err != nil {
		t.Fatalf("handler.Proxy failed: %v", err)
	}

	// 4. Verify
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
		t.Logf("Body: %s", rec.Body.String())
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatalf("No choices in response")
	}

	content := resp.Choices[0].Message.Content
	if !strings.Contains(content, "Mock response") {
		t.Errorf("Expected mock response, got: %s", content)
	}
	if !strings.Contains(content, "Hello Mock") {
		t.Errorf("Expected echo of input, got: %s", content)
	}
}

func TestMockProvider_Streaming_Integration(t *testing.T) {
	// 1. Setup Dependencies
	registry := domain.NewProviderRegistry()
	mwRegistry := domain.NewMiddlewareRegistry()
	keyService := service.NewKeyService(&MockRepository{}, registry, mwRegistry)
	mockProv := providers.NewMockProvider()
	registry.Register(mockProv.Name(), mockProv)

	executionHandler := engine.NewExecutionHandler(&MockRepository{})
	proxyService := service.NewProxyService(executionHandler, mwRegistry, keyService)
	handler := api.NewProxyHandler(proxyService, registry)

	// 2. Setup Echo
	e := echo.New()
	reqBody := `{"model": "mock-gpt-4", "stream": true, "messages": [{"role": "user", "content": "Hello Stream"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	appKey := &domain.Key{
		ID: 1,
		Configuration: &domain.KeyConfiguration{
			Provider: domain.PluginConfig{ID: "mock"},
		},
	}
	c.Set("app_key", appKey)

	// 3. Execute
	if err := handler.Proxy(c); err != nil {
		t.Fatalf("handler.Proxy failed: %v", err)
	}

	// 4. Verify
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("Expected text/event-stream, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "data: ") {
		t.Errorf("Expected SSE data chunks, got: %s", body)
	}
	if !strings.Contains(body, "[DONE]") {
		t.Errorf("Expected [DONE] marker")
	}
}
