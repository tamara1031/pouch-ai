package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           int
	OpenAIURL      string
	DataDir        string
	AllowedOrigins []string
	OpenAIKey      string
}

func New() *Config {
	return &Config{
		Port:           8080,
		OpenAIURL:      "https://api.openai.com",
		DataDir:        "./data",
		AllowedOrigins: []string{"*"},
	}
}

func (cfg *Config) LoadEnv() error {
	// Override with Env vars (12-factor app)
	if val := os.Getenv("PORT"); val != "" {
		port, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = port
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

	return nil
}

func Load() (*Config, error) {
	cfg := New()
	if err := cfg.LoadEnv(); err != nil {
		return nil, err
	}
	return cfg, nil
}
