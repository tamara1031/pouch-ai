package proxy

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"

	"pouch-ai/internal/security"
)

type CredentialsManager struct {
	db *sql.DB
}

func NewCredentialsManager(db *sql.DB) *CredentialsManager {
	return &CredentialsManager{db: db}
}

// GetAPIKey retrieves the API key for a provider.
// Priority: 1) Environment variable (OPENAI_API_KEY), 2) Database.
func (cm *CredentialsManager) GetAPIKey(provider string, password string) (string, error) {
	// Check environment variable first (for container deployments)
	if provider == "openai" {
		if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
			return envKey, nil
		}
	}

	// Fallback to database
	var encryptedKey, saltStr string
	err := cm.db.QueryRow("SELECT encrypted_key, salt FROM credentials WHERE provider = ?", provider).Scan(&encryptedKey, &saltStr)
	if err != nil {
		return "", err
	}

	masterPassword := "pouch-default-insecure-master-password"

	salt, err := base64Decode(saltStr)
	if err != nil {
		return "", fmt.Errorf("invalid salt: %w", err)
	}

	key := security.DeriveKey(masterPassword, salt)
	return security.Decrypt(key, encryptedKey)
}

// SetAPIKey encrypts and stores the API key.
func (cm *CredentialsManager) SetAPIKey(provider string, apiKey string) error {
	masterPassword := "pouch-default-insecure-master-password"

	salt, err := security.GenerateSalt()
	if err != nil {
		return err
	}

	key := security.DeriveKey(masterPassword, salt)
	encrypted, err := security.Encrypt(key, apiKey)
	if err != nil {
		return err
	}

	saltStr := base64Encode(salt)

	_, err = cm.db.Exec("INSERT OR REPLACE INTO credentials (provider, encrypted_key, salt) VALUES (?, ?, ?)",
		provider, encrypted, saltStr)
	return err
}

func base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
