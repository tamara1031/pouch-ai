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

	// 3. Initialize Application Services
	mwRegistry := domain.NewMiddlewareRegistry()
	pRegistry := domain.NewProviderRegistry()

	// Initialize Plugin Manager and register built-ins
	pluginManager := plugins.NewPluginManager(mwRegistry, pRegistry, cfg, "./backend/plugins/middlewares")
	if err := pluginManager.InitializeBuiltins(); err != nil {
		return nil, fmt.Errorf("failed to initialize built-in plugins: %w", err)
	}

	// Load external plugins
	if err := pluginManager.LoadPlugins(); err != nil {
		fmt.Printf("WARN: Failed to load external plugins: %v\n", err)
	}

	keyService := service.NewKeyService(keyRepo, pRegistry, mwRegistry)
	executionHandler := engine.NewExecutionHandler(keyRepo)
	proxyService := service.NewProxyService(executionHandler, mwRegistry, keyService)

	// 4. Initialize Handlers
	keyHandler := api.NewKeyHandler(keyService)
	proxyHandler := api.NewProxyHandler(proxyService, pRegistry)

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
	apiGroup.GET("/config/middlewares", keyHandler.ListMiddlewares)

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
