package main

import (
	"flag"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"pouch-ai/internal/server"
	"pouch-ai/ui"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	target := flag.String("target", "https://api.openai.com", "Target OpenAI API Base URL")
	dataDir := flag.String("data", "./data", "Directory to store data")
	corsOrigins := flag.String("cors-origins", "*", "Comma-separated list of allowed CORS origins")
	flag.Parse()

	// Ensure absolute path for data integrity
	absDataDir, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Fatalf("Invalid data path: %v", err)
	}

	log.Printf("Starting pouch-ai on port %d...", *port)
	log.Printf("Data Directory: %s", absDataDir)
	log.Printf("Target Proxy: %s", *target)

	// Sub-filesystem for ui/dist
	distFS, err := fs.Sub(ui.Assets, "dist")
	if err != nil {
		log.Fatalf("Failed to create sub FS for UI: %v", err)
	}

	var allowedOrigins []string
	if *corsOrigins != "" {
		for _, o := range strings.Split(*corsOrigins, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	srv, err := server.New(absDataDir, *port, *target, distFS, allowedOrigins)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
