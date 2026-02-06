package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type KeyManager struct {
	db *sql.DB
}

type KeyInfo struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Prefix       string  `json:"prefix"`
	ExpiresAt    *int64  `json:"expires_at"`
	BudgetLimit  float64 `json:"budget_limit"`
	BudgetUsage  float64 `json:"budget_usage"`
	BudgetPeriod string  `json:"budget_period"`
	IsMock       bool    `json:"is_mock"`
	MockConfig   string  `json:"mock_config"`
	RateLimit    int     `json:"rate_limit"`  // requests per period (0 = unlimited)
	RatePeriod   string  `json:"rate_period"` // "second", "minute", "none"
	CreatedAt    int64   `json:"created_at"`
}

func NewKeyManager(db *sql.DB) *KeyManager {
	return &KeyManager{db: db}
}

// GenerateKey creates a new API key.
func (km *KeyManager) GenerateKey(name string, expiresAt *int64, budgetLimit float64, period string, isMock bool, mockConfig string, rateLimit int, ratePeriod string) (string, error) {
	// Generate random key
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	key := "pk-" + hex.EncodeToString(bytes)

	// Hash key
	hash := sha256.Sum256([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	// Store in DB
	prefix := key[:7] + "..."
	createdAt := time.Now().Unix()

	// Default rate limit if not specified
	if rateLimit <= 0 {
		rateLimit = 10
	}
	if ratePeriod == "" {
		ratePeriod = "minute"
	}

	_, err := km.db.Exec(`
		INSERT INTO app_keys (name, key_hash, prefix, expires_at, budget_limit, budget_usage, budget_period, last_reset_at, is_mock, mock_config, rate_limit, rate_period, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, name, hashStr, prefix, expiresAt, budgetLimit, 0, period, createdAt, isMock, mockConfig, rateLimit, ratePeriod, createdAt)

	if err != nil {
		return "", fmt.Errorf("failed to insert key: %w", err)
	}

	return key, nil
}

// VerifyKey checks if a key is valid, handles auto-renew, and returns info.
func (km *KeyManager) VerifyKey(key string) (*KeyInfo, error) {
	hash := sha256.Sum256([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	var info KeyInfo
	var expiresAt sql.NullInt64
	var lastResetAt int64

	err := km.db.QueryRow(`
		SELECT id, name, prefix, expires_at, budget_limit, budget_usage, budget_period, last_reset_at, is_mock, mock_config, rate_limit, rate_period, created_at
		FROM app_keys WHERE key_hash = ?
	`, hashStr).Scan(&info.ID, &info.Name, &info.Prefix, &expiresAt, &info.BudgetLimit, &info.BudgetUsage, &info.BudgetPeriod, &lastResetAt, &info.IsMock, &info.MockConfig, &info.RateLimit, &info.RatePeriod, &info.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Invalid key
		}
		return nil, err
	}

	now := time.Now().Unix()

	// Check Expiration
	if expiresAt.Valid {
		val := expiresAt.Int64
		info.ExpiresAt = &val
		if now > val {
			return nil, fmt.Errorf("key expired")
		}
	}

	// Handle Auto-Renew (Lazy Reset)
	if info.BudgetPeriod != "" && info.BudgetPeriod != "none" {
		resetNeeded := false
		switch info.BudgetPeriod {
		case "weekly":
			if now > lastResetAt+(7*24*3600) {
				resetNeeded = true
			}
		case "monthly":
			if now > lastResetAt+(30*24*3600) {
				resetNeeded = true
			}
		}

		if resetNeeded {
			// Reset usage
			_, err := km.db.Exec("UPDATE app_keys SET budget_usage = 0, last_reset_at = ? WHERE id = ?", now, info.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to reset budget: %w", err)
			}
			info.BudgetUsage = 0
			// We should technically update local 'lastResetAt' but we don't return it
		}
	}

	// Check Budget (Standard keys only? Or mocks too? Let's assume mocks obey limits too if set, or ignore? User said "dummy key", maybe unlimited? But "mock" is just a mode. If they set a budget, we obey it.)
	// Actually, if it's a mock key for testing, maybe we want to test "Budget Exceeded" scenario?
	// So we should enforce it.
	if info.BudgetLimit > 0 && info.BudgetUsage >= info.BudgetLimit {
		return nil, fmt.Errorf("budget limit exceeded")
	}

	return &info, nil
}

// IncrementUsage updates the budget usage for a key.
func (km *KeyManager) IncrementUsage(keyID int64, cost float64) error {
	_, err := km.db.Exec("UPDATE app_keys SET budget_usage = budget_usage + ? WHERE id = ?", cost, keyID)
	return err
}

func (km *KeyManager) ListKeys() ([]KeyInfo, error) {
	rows, err := km.db.Query("SELECT id, name, prefix, expires_at, budget_limit, budget_usage, budget_period, is_mock, mock_config, rate_limit, rate_period, created_at FROM app_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]KeyInfo, 0)
	for rows.Next() {
		var k KeyInfo
		var expiresAt sql.NullInt64
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &expiresAt, &k.BudgetLimit, &k.BudgetUsage, &k.BudgetPeriod, &k.IsMock, &k.MockConfig, &k.RateLimit, &k.RatePeriod, &k.CreatedAt); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			val := expiresAt.Int64
			k.ExpiresAt = &val
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// UpdateKey updates modifiable fields of a key.
func (km *KeyManager) UpdateKey(id int, name string, budgetLimit float64, isMock bool, mockConfig string, rateLimit int, ratePeriod string) error {
	_, err := km.db.Exec(`
		UPDATE app_keys 
		SET name = ?, budget_limit = ?, is_mock = ?, mock_config = ?, rate_limit = ?, rate_period = ?
		WHERE id = ?
	`, name, budgetLimit, isMock, mockConfig, rateLimit, ratePeriod, id)
	return err
}

func (km *KeyManager) RevokeKey(id int) error {
	_, err := km.db.Exec("DELETE FROM app_keys WHERE id = ?", id)
	return err
}
