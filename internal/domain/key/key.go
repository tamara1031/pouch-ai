package key

import (
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
	return nil
}
