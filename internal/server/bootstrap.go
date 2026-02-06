package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"pouch-ai/internal/budget"
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
	budg := budget.NewManager(database.DB)
	tok := token.NewCounter()
	pric, err := proxy.NewPricing()
	if err != nil {
		return nil, fmt.Errorf("failed to load pricing: %w", err)
	}

	// 3. Proxy Handler
	prox, err := proxy.NewHandler(budg, tok, pric, targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init proxy: %w", err)
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
	api.POST("/chat/completions", prox.Handle)
    
    // Admin / System Routes
    e.GET("/stats/budget", func(c echo.Context) error {
        bal, err := budg.GetBalance()
        if err != nil {
            return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
        }
        return c.JSON(http.StatusOK, map[string]float64{"budget": bal})
    })

    e.POST("/stats/budget", func(c echo.Context) error {
        var req struct {
            Amount float64 `json:"amount"`
        }
        if err := c.Bind(&req); err != nil {
             return echo.NewHTTPError(http.StatusBadRequest, err.Error())
        }
        if err := budg.SetBalance(req.Amount); err != nil {
             return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
        }
        return c.JSON(http.StatusOK, "updated")
    })

	// Serve UI embedded
    // Let's assume we build the UI to `ui/dist` then `go build` embeds it.
    // For this step, I'll modify this file to accept `http.FileSystem` or similar.
    // But wait, `//go:embed` must be in the package.
    
    return &Server{echo: e, Port: port}, nil
}

func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.Port))
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.echo.Shutdown(ctx)
}
