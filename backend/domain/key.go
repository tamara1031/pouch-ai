package domain

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

var keyNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

const MaxKeyNameLength = 50

type ID int64

type PluginConfig struct {
	ID     string         `json:"id"`
	Config map[string]any `json:"config,omitempty"`
}

type KeyConfiguration struct {
	Provider    PluginConfig   `json:"provider"`
	Middlewares []PluginConfig `json:"middlewares"`
}

type Key struct {
	ID            ID                `json:"id"`
	Name          string            `json:"name"`
	KeyHash       string            `json:"key_hash"`
	Prefix        string            `json:"prefix"`
	ExpiresAt     *time.Time        `json:"expires_at"`
	BudgetUsage   float64           `json:"budget_usage"`
	LastResetAt   time.Time         `json:"last_reset_at"`
	CreatedAt     time.Time         `json:"created_at"`
	Configuration *KeyConfiguration `json:"configuration"`
}

func (k *Key) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

func (k *Key) Validate() error {
	if k.Name == "" {
		return &ValidationError{"key name is required"}
	}
	if len(k.Name) > MaxKeyNameLength {
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
