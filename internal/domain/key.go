package domain

import (
	"context"
	"fmt"
	"time"
)

type ID int64

type Budget struct {
	Limit  float64
	Usage  float64
	Period string // "monthly", "weekly", "none"
}

type RateLimit struct {
	Limit  int
	Period string // "second", "minute", "none"
}

type Key struct {
	ID          ID
	Name        string
	Provider    string // "openai", "anthropic", etc.
	KeyHash     string
	Prefix      string
	ExpiresAt   *time.Time
	Budget      Budget
	RateLimit   RateLimit
	IsMock      bool
	MockConfig  string
	LastResetAt time.Time
	CreatedAt   time.Time
}

func (k *Key) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

func (k *Key) IsBudgetExceeded() bool {
	if k.Budget.Limit <= 0 {
		return false
	}
	return k.Budget.Usage >= k.Budget.Limit
}

func (k *Key) NeedsReset() bool {
	if k.Budget.Period == "" || k.Budget.Period == "none" {
		return false
	}

	now := time.Now()
	var duration time.Duration
	switch k.Budget.Period {
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
	k.Budget.Usage = 0
	k.LastResetAt = time.Now()
}

func (k *Key) Validate() error {
	if k.Name == "" {
		return fmt.Errorf("key name is required")
	}
	if k.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	switch k.Budget.Period {
	case "monthly", "weekly", "none", "":
		// OK
	default:
		return fmt.Errorf("invalid budget period: %s", k.Budget.Period)
	}

	switch k.RateLimit.Period {
	case "second", "minute", "none", "":
		// OK
	default:
		return fmt.Errorf("invalid rate limit period: %s", k.RateLimit.Period)
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
