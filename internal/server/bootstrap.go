package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"pouch-ai/internal/api/http/handler"
	pouch_mw "pouch-ai/internal/api/http/middleware"
	"pouch-ai/internal/app"
	app_mw "pouch-ai/internal/app/middleware"
	"pouch-ai/internal/database"
	"pouch-ai/internal/domain/provider"
	"pouch-ai/internal/infra/db"
	"pouch-ai/internal/infra/provider/openai"
	infra_proxy "pouch-ai/internal/infra/proxy"
)

type Server struct {
	echo *echo.Echo
	Port int
}

func New(dataDir string, port int, targetURL string, assets fs.FS) (*Server, error) {
	// 1. Init Database
	if err := database.InitDB(dataDir); err != nil {
		return nil, err
	}

	// 2. Initialize Repositories and Infrastructure
	keyRepo := db.NewSQLiteKeyRepository(database.DB)

	pricing, err := openai.NewPricing()
	if err != nil {
		return nil, fmt.Errorf("failed to load pricing: %w", err)
	}

	tokenCounter := openai.NewTiktokenCounter()

	openaiProv := openai.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), targetURL, pricing, tokenCounter)

	registry := provider.NewRegistry()
	registry.Register(openaiProv)

	// 3. Initialize Application Services
	keyService := app.NewKeyService(keyRepo)

	executionHandler := infra_proxy.NewExecutionHandler()
	proxyService := app.NewProxyService(
		executionHandler,
		app_mw.NewRateLimitMiddleware(),
		app_mw.NewBudgetResetMiddleware(keyService),
		app_mw.NewKeyValidationMiddleware(),
		app_mw.NewUsageTrackingMiddleware(keyService),
		app_mw.NewMockMiddleware(),
	)

	// 4. Initialize Handlers
	keyHandler := handler.NewKeyHandler(keyService)
	proxyHandler := handler.NewProxyHandler(proxyService, registry)

	// 5. Echo Setup
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 7. Routes
	api := e.Group("/v1")

	// Proxy Route
	api.POST("/chat/completions", proxyHandler.Proxy, pouch_mw.AuthMiddleware(keyService))

	// Config Routes
	api.GET("/config/app-keys", keyHandler.ListKeys)
	api.POST("/config/app-keys", keyHandler.CreateKey)
	api.PUT("/config/app-keys/:id", keyHandler.UpdateKey)
	api.DELETE("/config/app-keys/:id", keyHandler.DeleteKey)

	// UI
	e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(assets))))

	return &Server{echo: e, Port: port}, nil
}

func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.Port))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
