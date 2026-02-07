package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"pouch-ai/internal/api"
	"pouch-ai/internal/database"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/infra"
	"pouch-ai/internal/service"
	service_mw "pouch-ai/internal/service/middleware"
)

type Server struct {
	echo *echo.Echo
	Port int
}

func New(dataDir string, port int, targetURL string, assets fs.FS, allowedOrigins []string) (*Server, error) {
	// 1. Init Database
	if err := database.InitDB(dataDir); err != nil {
		return nil, err
	}

	// 2. Initialize Repositories and Infrastructure
	keyRepo := infra.NewSQLiteKeyRepository(database.DB)

	pricing, err := infra.NewOpenAIPricing()
	if err != nil {
		return nil, fmt.Errorf("failed to load pricing: %w", err)
	}

	tokenCounter := infra.NewTiktokenCounter()

	registry := domain.NewRegistry()

	// Register OpenAI Provider
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey != "" {
		openaiProv := infra.NewOpenAIProvider(openaiKey, targetURL, pricing, tokenCounter)
		registry.Register(openaiProv)
	}

	// Register Anthropic Provider
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey != "" {
		aPricing, err := infra.NewAnthropicPricing()
		if err != nil {
			return nil, fmt.Errorf("failed to load anthropic pricing: %w", err)
		}
		anthropicProv := infra.NewAnthropicProvider(anthropicKey, aPricing, tokenCounter)
		registry.Register(anthropicProv)
	}

	// Register Gemini Provider
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey != "" {
		gPricing, err := infra.NewGeminiPricing()
		if err != nil {
			return nil, fmt.Errorf("failed to load gemini pricing: %w", err)
		}
		geminiProv := infra.NewGeminiProvider(geminiKey, gPricing, tokenCounter)
		registry.Register(geminiProv)
	}

	// 3. Initialize Application Services
	keyService := service.NewKeyService(keyRepo, registry)

	executionHandler := infra.NewExecutionHandler(keyRepo)
	proxyService := service.NewProxyService(
		executionHandler,
		service_mw.NewRateLimitMiddleware(),               // Shut out request if rate limit is exceeded
		service_mw.NewBudgetResetMiddleware(keyService),   // Reset budget if needed
		service_mw.NewKeyValidationMiddleware(),           // check if key is expired
		service_mw.NewUsageTrackingMiddleware(keyService), // add usage and check if budget will be exceeded
		service_mw.NewMockMiddleware(),                    // Mock requests if needed
	)

	// 4. Initialize Handlers
	keyHandler := api.NewKeyHandler(keyService)
	proxyHandler := api.NewProxyHandler(proxyService, registry)

	// 5. Echo Setup
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	corsConfig := middleware.DefaultCORSConfig
	if len(allowedOrigins) > 0 {
		corsConfig.AllowOrigins = allowedOrigins
	}
	e.Use(middleware.CORSWithConfig(corsConfig))

	// 7. Routes
	apiGroup := e.Group("/v1")

	// Proxy Route
	apiGroup.POST("/chat/completions", proxyHandler.Proxy, api.AuthMiddleware(keyService))

	// Config Routes
	apiGroup.GET("/config/app-keys", keyHandler.ListKeys)
	apiGroup.POST("/config/app-keys", keyHandler.CreateKey)
	apiGroup.PUT("/config/app-keys/:id", keyHandler.UpdateKey)
	apiGroup.DELETE("/config/app-keys/:id", keyHandler.DeleteKey)
	apiGroup.GET("/config/providers/usage", keyHandler.GetProviderUsage)

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
