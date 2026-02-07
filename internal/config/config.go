package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           int
	TargetURL      string
	DataDir        string
	AllowedOrigins []string
	OpenAIKey      string
}

func Load() (*Config, error) {
	// Defaults
	cfg := &Config{
		Port:           8080,
		TargetURL:      "https://api.openai.com",
		DataDir:        "./data",
		AllowedOrigins: []string{"*"},
	}

	// Override with Env vars (12-factor app)
	if val := os.Getenv("PORT"); val != "" {
		port, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = port
	}

	if val := os.Getenv("TARGET_URL"); val != "" {
		cfg.TargetURL = val
	}

	if val := os.Getenv("DATA_DIR"); val != "" {
		cfg.DataDir = val
	}

	if val := os.Getenv("CORS_ORIGINS"); val != "" {
		origins := strings.Split(val, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		cfg.AllowedOrigins = origins
	}

	if val := os.Getenv("OPENAI_API_KEY"); val != "" {
		cfg.OpenAIKey = val
	}

	return cfg, nil
}
