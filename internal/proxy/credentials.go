package proxy

import (
	"errors"
	"os"
)

// CredentialsManager provides API keys from environment variables.
type CredentialsManager struct{}

func NewCredentialsManager() *CredentialsManager {
	return &CredentialsManager{}
}

// GetAPIKey retrieves the API key for a provider from environment variables.
func (cm *CredentialsManager) GetAPIKey(provider string) (string, error) {
	switch provider {
	case "openai":
		if key := os.Getenv("OPENAI_API_KEY"); key != "" {
			return key, nil
		}
		return "", errors.New("OPENAI_API_KEY environment variable not set")
	default:
		return "", errors.New("unknown provider: " + provider)
	}
}
