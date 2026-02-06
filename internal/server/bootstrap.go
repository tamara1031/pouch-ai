package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"pouch-ai/internal/api/http/handler"
	pouch_mw "pouch-ai/internal/api/http/middleware"
	"pouch-ai/internal/app"
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
		app.NewRateLimitMiddleware(),
		app.NewMockMiddleware(),
		app.NewUsageTrackingMiddleware(keyService),
	)

	// 4. Initialize Handlers
	keyHandler := handler.NewKeyHandler(keyService)
	proxyHandler := handler.NewProxyHandler(proxyService, registry)

	// 5. Echo Setup
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 6. Global Rate Limiter (Panic Guard)
	throttleRate := 100
	if envRate := os.Getenv("THROTTLE_RATE"); envRate != "" {
		if r, err := strconv.Atoi(envRate); err == nil && r > 0 {
			throttleRate = r
		}
	}
	limiter := rate.NewLimiter(rate.Limit(float64(throttleRate)/60.0), 10)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !limiter.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Panic Guard: Rate limit exceeded")
			}
			return next(c)
		}
	})

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
