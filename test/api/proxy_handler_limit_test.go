package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"pouch-ai/backend/api"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/infra"
	"pouch-ai/backend/service"
)

func TestProxy_LargeBody(t *testing.T) {
	// 1. Setup Mock Dependencies
	registry := domain.NewRegistry()

	// 2. Setup Service
	executionHandler := infra.NewExecutionHandler(nil)
	proxyService := service.NewProxyService(executionHandler)
	handler := api.NewProxyHandler(proxyService, registry)

	// 3. Setup Echo
	e := echo.New()

	// Create a large body > 10MB (e.g. 11MB)
	size := 11 * 1024 * 1024
	largeBody := bytes.Repeat([]byte("a"), size)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(largeBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 4. Execute
	err := handler.Proxy(c)

	// 5. Assertions
	// We expect the handler to return an error, but verify the response code recorded.
	// Note: handler.Proxy may return nil if it calls c.JSON/c.String successfully.

	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code != http.StatusRequestEntityTooLarge {
				t.Errorf("expected status 413, got %d", he.Code)
			}
			return
		}
		// If it is another type of error, we check the recorded response
	}

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", rec.Code)
	}
}
