package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pouch-ai/backend/api"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"

	"github.com/labstack/echo/v4"
)

// MockRepository implements domain.Repository for testing
type MockRepository struct{}

func (m *MockRepository) Save(ctx context.Context, k *domain.Key) error { return nil }
func (m *MockRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Key, error) {
	return nil, nil
}
func (m *MockRepository) GetByHash(ctx context.Context, hash string) (*domain.Key, error) {
	return nil, nil
}
func (m *MockRepository) List(ctx context.Context) ([]*domain.Key, error) { return nil, nil }
func (m *MockRepository) Update(ctx context.Context, k *domain.Key) error { return nil }
func (m *MockRepository) Delete(ctx context.Context, id domain.ID) error  { return nil }
func (m *MockRepository) IncrementUsage(ctx context.Context, id domain.ID, amount float64) error {
	return nil
}
func (m *MockRepository) ResetUsage(ctx context.Context, id domain.ID, lastResetAt time.Time) error {
	return nil
}

// MockProvider implements domain.Provider for testing
type MockProvider struct{}

func (p *MockProvider) Name() string { return "test-provider" }
func (p *MockProvider) Configure(config map[string]string) (domain.Provider, error) {
	return p, nil
}
func (p *MockProvider) GetPricing(model domain.Model) (domain.Pricing, error) {
	return domain.Pricing{}, nil
}
func (p *MockProvider) CountTokens(model domain.Model, text string) (int, error) { return 0, nil }
func (p *MockProvider) PrepareHTTPRequest(ctx context.Context, model domain.Model, body []byte) (*http.Request, error) {
	return nil, nil
}
func (p *MockProvider) EstimateUsage(model domain.Model, requestBody []byte) (*domain.Usage, error) {
	return nil, nil
}
func (p *MockProvider) ParseOutputUsage(model domain.Model, responseBody []byte, isStream bool) (int, error) {
	return 0, nil
}
func (p *MockProvider) ProcessStreamChunk(chunk []byte) (string, error) {
	return "", nil
}
func (p *MockProvider) ParseRequest(body []byte) (domain.Model, bool, error) {
	return "", false, nil
}
func (p *MockProvider) GetUsage(ctx context.Context) (float64, error) { return 0, nil }

func TestKeyHandler_CreateKey_Validation(t *testing.T) {
	// Setup
	mockRepo := &MockRepository{}
	registry := domain.NewRegistry()
	registry.Register(&MockProvider{})

	mwReg := domain.NewMiddlewareRegistry()
	keyService := service.NewKeyService(mockRepo, registry, mwReg)
	handler := api.NewKeyHandler(keyService)
	e := echo.New()

	// Test Case 1: Very long name
	longName := strings.Repeat("a", 10000)
	reqBody := `{"name": "` + longName + `", "provider": "test-provider"}`
	req := httptest.NewRequest(http.MethodPost, "/keys", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.CreateKey(c); err != nil {
		// Handler might return error which Echo handles, or return nil if it wrote to response.
		// In our implementation, it returns error if BadRequest is called.
		// But we check rec.Code mostly.
		// Wait, if handler returns error, Echo usually processes it. But here we are calling handler directly.
		// If it returns error, we can check that error.
	}

	// In Echo, if handler returns error, we can assert on that if we want, but checking recorder state depends on if error was handled.
	// BadRequest(c, ...) writes to c.Response usually if using echo.Context default implementation?
	// Wait, api.BadRequest returns an error. It does NOT write to response automatically unless echo.Context is fully wired with ErrorHandler?
	// Let's check api.BadRequest again.
	/*
	   func NewAPIError(c echo.Context, status int, message string) error {
	       return c.JSON(status, APIError{...})
	   }
	*/
	// It calls c.JSON, which writes to response. So rec should have the status.

	if rec.Code == http.StatusBadRequest {
		// Pass
	} else {
		t.Errorf("Test Case 1 (Long Name): Expected status 400, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Test Case 2: XSS Payload / Invalid Chars
	xssName := `<script>alert(1)</script>`
	reqBody = `{"name": "` + xssName + `", "provider": "test-provider"}`
	req = httptest.NewRequest(http.MethodPost, "/keys", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if err := handler.CreateKey(c); err != nil {
		// Log error if needed
	}

	if rec.Code == http.StatusBadRequest {
		// Pass
	} else {
		t.Errorf("Test Case 2 (XSS/Invalid Chars): Expected status 400, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Test Case 3: Valid Name
	validName := "valid-key-name"
	reqBody = `{"name": "` + validName + `", "provider": "test-provider"}`
	req = httptest.NewRequest(http.MethodPost, "/keys", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	handler.CreateKey(c)

	if rec.Code == http.StatusCreated {
		// Pass
	} else {
		t.Errorf("Test Case 3 (Valid Name): Expected status 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}
