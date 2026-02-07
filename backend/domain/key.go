package domain

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

var keyNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)

const MaxKeyNameLength = 50

type ID int64

type PluginConfig struct {
	ID     string            `json:"id"`
	Config map[string]string `json:"config,omitempty"`
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

func (k *Key) IsBudgetExceeded() bool {
	if k.Configuration == nil {
		return false
	}
	// Find budget middleware or provider limit
	for _, m := range k.Configuration.Middlewares {
		if m.ID == "budget" {
			limitStr := m.Config["limit"]
			var limit float64
			fmt.Sscanf(limitStr, "%f", &limit)
			if limit <= 0 {
				return false
			}
			return k.BudgetUsage >= limit
		}
	}
	return false
}

func (k *Key) NeedsReset() bool {
	if k.Configuration == nil {
		return false
	}

	var period string
	for _, m := range k.Configuration.Middlewares {
		if m.ID == "budget" {
			period = m.Config["period"]
			break
		}
	}

	if period == "" || period == "none" {
		return false
	}

	now := time.Now()
	var duration time.Duration
	switch period {
	case "weekly":
		duration = 7 * 24 * time.Hour
	case "monthly":
		duration = 30 * 24 * time.Hour
	default:
		return false
	}

	return now.After(k.LastResetAt.Add(duration))
}

func (k *Key) ResetUsage() {
	k.BudgetUsage = 0
	k.LastResetAt = time.Now()
}

func (k *Key) Validate() error {
	if k.Name == "" {
		return fmt.Errorf("key name is required")
	}
	if len(k.Name) > MaxKeyNameLength {
		return fmt.Errorf("key name is too long (max %d characters)", MaxKeyNameLength)
	}
	if !keyNameRegex.MatchString(k.Name) {
		return fmt.Errorf("key name contains invalid characters")
	}
	if k.Configuration == nil || k.Configuration.Provider.ID == "" {
		return fmt.Errorf("provider is required")
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
