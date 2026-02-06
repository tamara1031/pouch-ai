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
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Prefix      string  `json:"prefix"`
	ExpiresAt   *int64  `json:"expires_at"` // Unix timestamp
	BudgetLimit float64 `json:"budget_limit"`
	BudgetUsage float64 `json:"budget_usage"`
	CreatedAt   int64   `json:"created_at"`
}

func NewKeyManager(db *sql.DB) *KeyManager {
	return &KeyManager{db: db}
}

// GenerateKey creates a new API key, stores its hash, and returns the plain key.
func (km *KeyManager) GenerateKey(name string, expiresAt *int64, budgetLimit float64) (string, error) {
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

	_, err := km.db.Exec(`
		INSERT INTO app_keys (name, key_hash, prefix, expires_at, budget_limit, budget_usage, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, name, hashStr, prefix, expiresAt, budgetLimit, 0, createdAt)

	if err != nil {
		return "", fmt.Errorf("failed to insert key: %w", err)
	}

	return key, nil
}

// VerifyKey checks if a key is valid and within budget/expiration.
func (km *KeyManager) VerifyKey(key string) (*KeyInfo, error) {
	hash := sha256.Sum256([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	var info KeyInfo
	var expiresAt sql.NullInt64

	err := km.db.QueryRow(`
		SELECT id, name, prefix, expires_at, budget_limit, budget_usage, created_at
		FROM app_keys WHERE key_hash = ?
	`, hashStr).Scan(&info.ID, &info.Name, &info.Prefix, &expiresAt, &info.BudgetLimit, &info.BudgetUsage, &info.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Invalid key
		}
		return nil, err
	}

	if expiresAt.Valid {
		val := expiresAt.Int64
		info.ExpiresAt = &val
		if time.Now().Unix() > val {
			return nil, fmt.Errorf("key expired")
		}
	}

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
	rows, err := km.db.Query("SELECT id, name, prefix, expires_at, budget_limit, budget_usage, created_at FROM app_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []KeyInfo
	for rows.Next() {
		var k KeyInfo
		var expiresAt sql.NullInt64
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &expiresAt, &k.BudgetLimit, &k.BudgetUsage, &k.CreatedAt); err != nil {
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

func (km *KeyManager) RevokeKey(id int) error {
	_, err := km.db.Exec("DELETE FROM app_keys WHERE id = ?", id)
	return err
}
