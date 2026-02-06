package api_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"

	"pouch-ai/internal/api"
	"pouch-ai/internal/database"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/infra"
	"pouch-ai/internal/service"
	service_mw "pouch-ai/internal/service/middleware"
)

func TestBudgetRaceCondition(t *testing.T) {
	// 1. Setup In-Memory DB
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	database.DB = db

	// Initialize schema manually
	schema := `
	CREATE TABLE IF NOT EXISTS app_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		provider TEXT NOT NULL DEFAULT 'openai',
		key_hash TEXT NOT NULL UNIQUE,
		prefix TEXT NOT NULL,
		expires_at INTEGER,
		budget_limit REAL DEFAULT 0,
		budget_usage REAL DEFAULT 0,
		budget_period TEXT,
		last_reset_at INTEGER,
		is_mock BOOLEAN DEFAULT 0,
		mock_config TEXT,
		rate_limit INTEGER DEFAULT 10,
		rate_period TEXT DEFAULT 'minute',
		created_at INTEGER NOT NULL
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to init schema: %v", err)
	}

	// 2. Setup Dependencies
	keyRepo := infra.NewSQLiteKeyRepository(db)

	pricing, err := infra.NewOpenAIPricing()
	if err != nil {
		t.Fatalf("Failed to create pricing: %v", err)
	}
	tokenCounter := infra.NewTiktokenCounter()

	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Widen race window
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "Hello world"}}]}`))
	}))
	defer mockUpstream.Close()

	registry := domain.NewRegistry()
	provider := infra.NewOpenAIProvider("test-key", mockUpstream.URL, pricing, tokenCounter)
	registry.Register(provider)

	keyService := service.NewKeyService(keyRepo, registry)
	executionHandler := infra.NewExecutionHandler()

	proxyService := service.NewProxyService(
		executionHandler,
		service_mw.NewUsageTrackingMiddleware(keyService),
	)

	handler := api.NewProxyHandler(proxyService, registry)

	// 3. Create Key with budget 0.05
    ctx := context.Background()
	// CreateKey: budgetLimit=0.05
	_, k, err := keyService.CreateKey(ctx, "test-key", "openai", nil, 0.05, "monthly", false, "", 1000, "minute")
    if err != nil {
        t.Fatalf("failed to create key: %v", err)
    }

	// 4. Run Concurrent Requests
    var wg sync.WaitGroup
    successCount := 0
    failCount := 0
    var mu sync.Mutex

    // Use unknown model to trigger fallback reservation of 0.01
    reqBodyUnknown := `{"model": "gpt-unknown-expensive", "messages": [{"role": "user", "content": "Hello"}]}`

    totalReqs := 10
    for i := 0; i < totalReqs; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            e := echo.New()
            req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(reqBodyUnknown))
            req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            c.Set("app_key", k)

            err := handler.Proxy(c)

            mu.Lock()
            defer mu.Unlock()

            if err == nil && rec.Code == 200 {
                successCount++
            } else {
                failCount++
            }
        }()
    }

    wg.Wait()

    t.Logf("Success: %d, Failed: %d", successCount, failCount)

    // With budget 0.05 and reservation 0.01, max 5 concurrent requests should pass.
    if successCount > 5 {
        t.Errorf("Expected at most 5 successes, got %d", successCount)
    }

    if failCount < (totalReqs - 5) {
        t.Errorf("Expected at least %d failures, got %d", totalReqs-5, failCount)
    }
}
