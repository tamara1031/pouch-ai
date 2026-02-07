package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIProvider_GetUsage_UsesBaseURL(t *testing.T) {
	var serverHit bool

	// Setup mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHit = true
		// Check if request path contains the expected endpoint
		if !strings.Contains(r.URL.Path, "/dashboard/billing/usage") {
			t.Errorf("Expected request path to contain /dashboard/billing/usage, got %s", r.URL.Path)
		}

		// Return valid JSON response to avoid unmarshal errors
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"total_usage": 1234.56}`)
	}))
	defer mockServer.Close()

	// Initialize provider with mock server URL
	// We pass nil for dependencies that are not used in GetUsage
	provider := NewOpenAIProvider("test-key", mockServer.URL, nil, nil)

	// Call GetUsage
	// We expect this to fail initially (or return error from OpenAI), so we don't strictly check err yet,
	// or we check if serverHit is true.
	usage, err := provider.GetUsage(context.Background())

	if !serverHit {
		t.Error("Expected mock server to be hit, but it was not. Code likely ignoring baseURL.")
	}

	// If server was hit (after fix), we expect success
	if serverHit {
		if err != nil {
			t.Errorf("GetUsage failed: %v", err)
		}
		expectedUsage := 12.3456
		if usage != expectedUsage {
			t.Errorf("Expected usage %f, got %f", expectedUsage, usage)
		}
	}
}
