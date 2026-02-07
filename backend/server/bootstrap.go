package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"pouch-ai/backend/api"
	"pouch-ai/backend/config"
	"pouch-ai/backend/database"
	"pouch-ai/backend/domain"
"pouch-ai/backend/infra/engine"
"pouch-ai/backend/plugins"
	"pouch-ai/backend/plugins/middlewares"
	"pouch-ai/backend/plugins/providers"
	"pouch-ai/backend/service"
)

type Server struct {
	echo *echo.Echo
	Port int
}

func New(cfg *config.Config, assets fs.FS) (*Server, error) {
	// 1. Init Database
	if err := database.InitDB(cfg.DataDir); err != nil {
		return nil, err
	}

	// 2. Initialize Repositories and Infrastructure
	keyRepo := database.NewSQLiteKeyRepository(database.DB)

	pricing, err := providers.NewOpenAIPricing()
	if err != nil {
		return nil, fmt.Errorf("failed to load pricing: %w", err)
	}

	tokenCounter := providers.NewTiktokenCounter()

	registry := domain.NewProviderRegistry()

	// Register OpenAI Provider
	// Use config for OpenAI Key if available, or fallback to Env (though config loader handles env)
	// Actually config loader handles env, so we should use cfg.OpenAIKey
	if cfg.OpenAIKey != "" {
		openaiProv := providers.NewOpenAIProvider(cfg.OpenAIKey, cfg.TargetURL, pricing, tokenCounter)
		registry.Register(openaiProv)
	} else {
		fmt.Println("WARN: OpenAI API Key not found. 'openai' provider will be unavailable.")
	}

	// Register Mock Provider
	mockProv := providers.NewMockProvider()
	registry.Register(mockProv)

	// TODO: Add more providers (Anthropic, Gemini, etc.) here

	// 3. Initialize Application Services
	mwRegistry := domain.NewMiddlewareRegistry()
	keyService := service.NewKeyService(keyRepo, registry, mwRegistry)

	executionHandler := engine.NewExecutionHandler(keyRepo)
	mwRegistry.Register(middlewares.GetKeyValidationInfo(), middlewares.NewKeyValidationMiddleware)
	mwRegistry.Register(middlewares.GetUsageTrackingInfo(keyService), middlewares.NewUsageTrackingMiddleware(keyService))
	mwRegistry.Register(middlewares.GetInfo(), middlewares.NewRateLimitMiddleware)
	mwRegistry.Register(middlewares.GetBudgetEnforcementInfo(), middlewares.NewBudgetEnforcementMiddleware)
	mwRegistry.Register(middlewares.GetBudgetResetInfo(keyService), middlewares.NewBudgetResetMiddleware(keyService))

	// Load external plugins
	pluginManager := plugins.NewPluginManager(mwRegistry, "./backend/plugins/middlewares")
	if err := pluginManager.LoadPlugins(); err != nil {
		fmt.Printf("WARN: Failed to load external plugins: %v\n", err)
	}

	proxyService := service.NewProxyService(executionHandler, mwRegistry)

	// 4. Initialize Handlers
	keyHandler := api.NewKeyHandler(keyService)
	proxyHandler := api.NewProxyHandler(proxyService, registry)

	// 5. Echo Setup
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	corsConfig := middleware.DefaultCORSConfig
	if len(cfg.AllowedOrigins) > 0 {
		corsConfig.AllowOrigins = cfg.AllowedOrigins
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
	apiGroup.GET("/config/providers", keyHandler.ListProviders)
	apiGroup.GET("/config/providers/usage", keyHandler.GetProviderUsage)

	// UI
	e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(assets))))

	return &Server{echo: e, Port: cfg.Port}, nil
}

func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.Port))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
