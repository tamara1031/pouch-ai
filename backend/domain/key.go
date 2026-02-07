package domain

import (
	"context"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"
)

var keyNameRegex = regexp.MustCompile(`^[\pL\pN_\-\s]+$`)

// ValidationError represents an error during data validation.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// IsValidationError checks if the error is a ValidationError.
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// MaxKeyNameLength defines the maximum allowed characters for a key name.
const MaxKeyNameLength = 50

// ID represents a unique identifier for entities.
type ID int64

// PluginConfig holds configuration for a specific plugin (provider or middleware).
type PluginConfig struct {
	ID     string         `json:"id"`
	Config map[string]any `json:"config,omitempty"`
}

// KeyConfiguration stores the operational settings for an API key.
type KeyConfiguration struct {
	Provider    PluginConfig   `json:"provider"`
	Middlewares []PluginConfig `json:"middlewares"`
	BudgetLimit float64        `json:"budget_limit"`
	ResetPeriod int            `json:"reset_period"`
}

// Key represents an API key entity with its metadata and usage stats.
type Key struct {
	ID            ID                `json:"id"`
	Name          string            `json:"name"`
	KeyHash       string            `json:"key_hash"`
	Prefix        string            `json:"prefix"`
	ExpiresAt     *time.Time        `json:"expires_at"`
	AutoRenew     bool              `json:"auto_renew"`
	BudgetUsage   float64           `json:"budget_usage"`
	LastResetAt   time.Time         `json:"last_reset_at"`
	CreatedAt     time.Time         `json:"created_at"`
	Configuration *KeyConfiguration `json:"configuration"`
}

// IsExpired checks if the key has passed its expiration time.
func (k *Key) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// Validate ensures the key's properties meet the required constraints.
func (k *Key) Validate() error {
	if k.Name == "" {
		return &ValidationError{"key name is required"}
	}
	if utf8.RuneCountInString(k.Name) > MaxKeyNameLength {
		return &ValidationError{fmt.Sprintf("key name is too long (max %d characters)", MaxKeyNameLength)}
	}
	if !keyNameRegex.MatchString(k.Name) {
		return &ValidationError{"key name contains invalid characters"}
	}
	if k.Configuration == nil || k.Configuration.Provider.ID == "" {
		return &ValidationError{"provider is required"}
	}

	return nil
}

// Repository defines the persistence operations for Keys.
type Repository interface {
	Save(ctx context.Context, k *Key) error
	GetByID(ctx context.Context, id ID) (*Key, error)
	GetByHash(ctx context.Context, hash string) (*Key, error)
	List(ctx context.Context) ([]*Key, error)
	Update(ctx context.Context, k *Key) error
	Delete(ctx context.Context, id ID) error
	IncrementUsage(ctx context.Context, id ID, amount float64) error
	ResetUsage(ctx context.Context, id ID, lastResetAt time.Time) error
}
