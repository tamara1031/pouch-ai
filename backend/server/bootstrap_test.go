package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"pouch-ai/backend/config"
	"testing"
	"testing/fstest"
)

func TestNew_CORS(t *testing.T) {
	// Create a temp dir for DB
	tempDir, err := os.MkdirTemp("", "pouch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock assets
	assets := fstest.MapFS{
		"index.html": {Data: []byte("html")},
	}

	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectedOrigin string
	}{
		{
			name:           "Default (Nil)",
			allowedOrigins: nil,
			requestOrigin:  "http://example.com",
			expectedOrigin: "*", // Default behavior
		},
		{
			name:           "Specific Origin Allowed",
			allowedOrigins: []string{"http://trusted.com"},
			requestOrigin:  "http://trusted.com",
			expectedOrigin: "http://trusted.com",
		},
		{
			name:           "Specific Origin Blocked",
			allowedOrigins: []string{"http://trusted.com"},
			requestOrigin:  "http://malicious.com",
			expectedOrigin: "", // Should not return header
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Use a unique subdir for each test to avoid DB locking issues
			testDir, err := os.MkdirTemp(tempDir, "test-*")
			if err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}

			cfg := &config.Config{
				DataDir:        testDir,
				Port:           0,
				TargetURL:      "http://target.com",
				AllowedOrigins: tc.allowedOrigins,
			}
			srv, err := New(cfg, assets)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			req := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
			req.Header.Set("Origin", tc.requestOrigin)
			req.Header.Set("Access-Control-Request-Method", "POST")
			rec := httptest.NewRecorder()

			srv.echo.ServeHTTP(rec, req)

			resp := rec.Result()
			gotOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			if gotOrigin != tc.expectedOrigin {
				t.Errorf("expected Access-Control-Allow-Origin: %q, got %q", tc.expectedOrigin, gotOrigin)
			}
		})
	}
}
