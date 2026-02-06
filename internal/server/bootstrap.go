package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"pouch-ai/internal/auth"
	"pouch-ai/internal/database"
	"pouch-ai/internal/proxy"
	"pouch-ai/internal/token"
)

type Server struct {
	echo *echo.Echo
	Port int
}

func New(dataDir string, port int, targetURL string, assets fs.FS) (*Server, error) {
	// 1. Init DB
	if err := database.InitDB(dataDir); err != nil {
		return nil, err
	}

	// 2. Dependencies
	tok := token.NewCounter()
	creds := proxy.NewCredentialsManager(database.DB)
	keyMgr := auth.NewKeyManager(database.DB)
	pric, err := proxy.NewPricing()
	if err != nil {
		return nil, fmt.Errorf("failed to load pricing: %w", err)
	}

	// 3. Proxy Handler
	prox, err := proxy.NewHandler(tok, pric, targetURL, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to init proxy: %w", err)
	}

	prox.UsageCallback = func(c echo.Context, cost float64) {
		keyID, ok := c.Get("app_key_id").(int64)
		if ok && keyID > 0 {
			if err := keyMgr.IncrementUsage(keyID, cost); err != nil {
				fmt.Printf("Failed to update key usage: %v\n", err)
			}
		}
	}

	// 4. Echo Setup
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 5. Rate Limiter (Panic Guard)
	// Allow 100 requests per minute burst. Generous for single user, tight for public.
	limiter := rate.NewLimiter(rate.Limit(100.0/60.0), 10)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !limiter.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Panic Guard: Rate limit exceeded")
			}
			return next(c)
		}
	})

	// 6. Routes
	api := e.Group("/v1")

	// OpenAI Chat Completions
	// OpenAI Chat Completions
	// OpenAI Chat Completions
	api.POST("/chat/completions", prox.Handle, AuthMiddleware(keyMgr))

	// Admin / System Routes

	// Config: API Keys (Provider)
	api.POST("/config/key", func(c echo.Context) error {
		var req struct {
			Provider string `json:"provider"`
			Key      string `json:"key"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if req.Provider == "" {
			req.Provider = "openai" // Default
		}

		if err := creds.SetAPIKey(req.Provider, req.Key); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to save key: %v", err))
		}
		return c.JSON(http.StatusOK, "key updated")
	})

	// Config: App Keys management
	api.GET("/config/app-keys", func(c echo.Context) error {
		keys, err := keyMgr.ListKeys()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, keys)
	})

	api.POST("/config/app-keys", func(c echo.Context) error {
		var req struct {
			Name         string  `json:"name"`
			ExpiresAt    *int64  `json:"expires_at"` // Nullable
			BudgetLimit  float64 `json:"budget_limit"`
			BudgetPeriod string  `json:"budget_period"`
			IsMock       bool    `json:"is_mock"`
			MockConfig   string  `json:"mock_config"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if req.Name == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
		}

		key, err := keyMgr.GenerateKey(req.Name, req.ExpiresAt, req.BudgetLimit, req.BudgetPeriod, req.IsMock, req.MockConfig)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, map[string]string{"key": key})
	})

	api.PUT("/config/app-keys/:id", func(c echo.Context) error {
		id := 0
		if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
		}
		var req struct {
			Name        string  `json:"name"`
			BudgetLimit float64 `json:"budget_limit"`
			IsMock      bool    `json:"is_mock"`
			MockConfig  string  `json:"mock_config"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if req.Name == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
		}

		if err := keyMgr.UpdateKey(id, req.Name, req.BudgetLimit, req.IsMock, req.MockConfig); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, "updated")
	})

	api.DELETE("/config/app-keys/:id", func(c echo.Context) error {
		id := 0
		if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
		}
		if err := keyMgr.RevokeKey(id); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, "deleted")
	})

	// Serve UI embedded
	// Serve static files from the root
	e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(assets))))

	return &Server{echo: e, Port: port}, nil
}

func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.Port))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
