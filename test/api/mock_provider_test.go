package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"pouch-ai/backend/api"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/infra"
	"pouch-ai/backend/service"
)

func TestMockProvider_Integration(t *testing.T) {
	// 1. Setup Dependencies
	registry := domain.NewRegistry()

	// Register Mock Provider
	mockProv := infra.NewMockProvider()
	registry.Register(mockProv)

	// Service
	executionHandler := infra.NewExecutionHandler(nil)

	proxyService := service.NewProxyService(
		executionHandler,
		// No middlewares for simplicity
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
		Provider: "mock",
		ID:       1, // Dummy ID
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
	registry := domain.NewRegistry()
	mockProv := infra.NewMockProvider()
	registry.Register(mockProv)

	executionHandler := infra.NewExecutionHandler(nil)
	proxyService := service.NewProxyService(executionHandler)
	handler := api.NewProxyHandler(proxyService, registry)

	// 2. Setup Echo
	e := echo.New()
	reqBody := `{"model": "mock-gpt-4", "stream": true, "messages": [{"role": "user", "content": "Hello Stream"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	appKey := &domain.Key{Provider: "mock", ID: 1}
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
