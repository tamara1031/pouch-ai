package key

import (
	"context"
	"time"
)

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
