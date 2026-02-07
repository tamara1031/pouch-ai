package main

import (
	"context"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"pouch-ai/backend/config"
	"pouch-ai/backend/server"
	ui "pouch-ai/frontend"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Allow flags to override env vars (optional, but good for local dev)
	port := flag.Int("port", cfg.Port, "Port to listen on")
	target := flag.String("target", cfg.TargetURL, "Target OpenAI API Base URL")
	dataDir := flag.String("data", cfg.DataDir, "Directory to store data")
	corsOrigins := flag.String("cors-origins", strings.Join(cfg.AllowedOrigins, ","), "Comma-separated list of allowed CORS origins")
	flag.Parse()

	// Update config from flags
	cfg.Port = *port
	cfg.TargetURL = *target
	cfg.DataDir = *dataDir
	if *corsOrigins != "" {
		cfg.AllowedOrigins = strings.Split(*corsOrigins, ",")
		for i := range cfg.AllowedOrigins {
			cfg.AllowedOrigins[i] = strings.TrimSpace(cfg.AllowedOrigins[i])
		}
	}

	// Ensure absolute path for data integrity
	absDataDir, err := filepath.Abs(cfg.DataDir)
	if err != nil {
		log.Fatalf("Invalid data path: %v", err)
	}
	cfg.DataDir = absDataDir

	log.Printf("Starting pouch-ai on port %d...", cfg.Port)
	log.Printf("Data Directory: %s", cfg.DataDir)
	log.Printf("Target Proxy: %s", cfg.TargetURL)

	// Sub-filesystem for ui/dist
	distFS, err := fs.Sub(ui.Assets, "dist")
	if err != nil {
		log.Fatalf("Failed to create sub FS for UI: %v", err)
	}

	srv, err := server.New(cfg, distFS)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}
	log.Println("Server exited")
}
