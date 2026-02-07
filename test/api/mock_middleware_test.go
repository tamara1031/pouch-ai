package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"pouch-ai/internal/api"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/infra"
	"pouch-ai/internal/service"
	"pouch-ai/internal/service/middleware"
)

func TestMockMiddleware_SecurityAndValidation(t *testing.T) {
	// Setup
	mockMiddleware := middleware.NewMockMiddleware()
	executionHandler := infra.NewExecutionHandler()
	proxyService := service.NewProxyService(executionHandler, mockMiddleware)

	registry := domain.NewRegistry()
	pricing, _ := infra.NewOpenAIPricing()
	tokenCounter := infra.NewTiktokenCounter()
	provider := infra.NewOpenAIProvider("test-key", "http://example.com", pricing, tokenCounter)
	registry.Register(provider)

	handler := api.NewProxyHandler(proxyService, registry)

	tests := []struct {
		name           string
		mockConfig     string
		expectedStatus int
		expectedBody   string // partial match
		expectJSON     bool
	}{
		{
			name:           "Valid JSON Config",
			mockConfig:     `{"choices": [{"message": {"content": "Mocked response"}}]}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `"Mocked response"`,
			expectJSON:     true,
		},
		{
			name:           "Invalid JSON Config (XSS Payload)",
			mockConfig:     `<html><script>alert('XSS')</script></html>`,
			expectedStatus: http.StatusInternalServerError, // We expect this to fail validation now
			expectedBody:   `Invalid mock configuration`,
			expectJSON:     true,
		},
		{
			name:           "Empty Config",
			mockConfig:     ``,
			expectedStatus: http.StatusInternalServerError, // Empty string is not valid JSON
			expectedBody:   `Invalid mock configuration`,
			expectJSON:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			reqBody := `{"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Hello"}]}`
			req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			appKey := &domain.Key{
				Provider:   "openai",
				IsMock:     true,
				MockConfig: tt.mockConfig,
			}
			c.Set("app_key", appKey)

			if err := handler.Proxy(c); err != nil {
				// ProxyHandler may return error which Echo handles.
				// But here we are calling handler.Proxy directly.
				// If it returns an echo.HTTPError, we should check that.
				if he, ok := err.(*echo.HTTPError); ok {
					if he.Code != tt.expectedStatus {
						t.Errorf("expected status %d, got %d", tt.expectedStatus, he.Code)
					}
					// Verify error message if possible
					if msg, ok := he.Message.(string); ok && !strings.Contains(msg, tt.expectedBody) {
						t.Errorf("expected error message to contain '%s', got '%s'", tt.expectedBody, msg)
					}
					return
				} else {
					// Other errors
					t.Fatalf("handler.Proxy returned unexpected error: %v", err)
				}
			}

			// If no error returned, check recorder
			resp := rec.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectJSON {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "application/json") {
					t.Errorf("expected Content-Type application/json, got '%s'", ct)
				}
			}

			// Check nosniff header (we will add this requirement)
			if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("expected X-Content-Type-Options: nosniff, got '%s'", resp.Header.Get("X-Content-Type-Options"))
			}

			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			if !strings.Contains(bodyString, tt.expectedBody) {
				t.Errorf("expected body to contain '%s', got '%s'", tt.expectedBody, bodyString)
			}
		})
	}
}
