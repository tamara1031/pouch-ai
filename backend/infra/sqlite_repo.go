package infra

import (
	"context"
	"database/sql"
	"pouch-ai/backend/domain"
	"time"
)

type SQLiteKeyRepository struct {
	db *sql.DB
}

func NewSQLiteKeyRepository(db *sql.DB) *SQLiteKeyRepository {
	return &SQLiteKeyRepository{db: db}
}

func (r *SQLiteKeyRepository) Save(ctx context.Context, k *domain.Key) error {
	var expiresAt *int64
	if k.ExpiresAt != nil {
		val := k.ExpiresAt.Unix()
		expiresAt = &val
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO app_keys (name, provider, key_hash, prefix, expires_at, budget_limit, budget_usage, budget_period, last_reset_at, is_mock, mock_config, rate_limit, rate_period, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, k.Name, k.Provider, k.KeyHash, k.Prefix, expiresAt, k.Budget.Limit, k.Budget.Usage, k.Budget.Period, k.LastResetAt.Unix(), k.IsMock, k.MockConfig, k.RateLimit.Limit, k.RateLimit.Period, k.CreatedAt.Unix())

	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		k.ID = domain.ID(id)
	}
	return nil
}

func (r *SQLiteKeyRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Key, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, provider, key_hash, prefix, expires_at, budget_limit, budget_usage, budget_period, last_reset_at, is_mock, mock_config, rate_limit, rate_period, created_at
		FROM app_keys WHERE id = ?
	`, id)
	return r.scanKey(row)
}

func (r *SQLiteKeyRepository) GetByHash(ctx context.Context, hash string) (*domain.Key, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, provider, key_hash, prefix, expires_at, budget_limit, budget_usage, budget_period, last_reset_at, is_mock, mock_config, rate_limit, rate_period, created_at
		FROM app_keys WHERE key_hash = ?
	`, hash)
	return r.scanKey(row)
}

func (r *SQLiteKeyRepository) List(ctx context.Context) ([]*domain.Key, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, provider, key_hash, prefix, expires_at, budget_limit, budget_usage, budget_period, last_reset_at, is_mock, mock_config, rate_limit, rate_period, created_at
		FROM app_keys ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.Key
	for rows.Next() {
		k, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *SQLiteKeyRepository) Update(ctx context.Context, k *domain.Key) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE app_keys 
		SET name = ?, provider = ?, budget_limit = ?, is_mock = ?, mock_config = ?, rate_limit = ?, rate_period = ?
		WHERE id = ?
	`, k.Name, k.Provider, k.Budget.Limit, k.IsMock, k.MockConfig, k.RateLimit.Limit, k.RateLimit.Period, k.ID)
	return err
}

func (r *SQLiteKeyRepository) Delete(ctx context.Context, id domain.ID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM app_keys WHERE id = ?", id)
	return err
}

func (r *SQLiteKeyRepository) IncrementUsage(ctx context.Context, id domain.ID, amount float64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE app_keys SET budget_usage = budget_usage + ? WHERE id = ?", amount, id)
	return err
}

func (r *SQLiteKeyRepository) ResetUsage(ctx context.Context, id domain.ID, lastResetAt time.Time) error {
	_, err := r.db.ExecContext(ctx, "UPDATE app_keys SET budget_usage = 0, last_reset_at = ? WHERE id = ?", lastResetAt.Unix(), id)
	return err
}

// Helpers

func (r *SQLiteKeyRepository) scanKey(sc interface {
	Scan(dest ...interface{}) error
}) (*domain.Key, error) {
	var k domain.Key
	var expiresAt sql.NullInt64
	var lastResetAt, createdAt int64

	err := sc.Scan(
		&k.ID, &k.Name, &k.Provider, &k.KeyHash, &k.Prefix, &expiresAt,
		&k.Budget.Limit, &k.Budget.Usage, &k.Budget.Period,
		&lastResetAt, &k.IsMock, &k.MockConfig,
		&k.RateLimit.Limit, &k.RateLimit.Period,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if expiresAt.Valid {
		t := time.Unix(expiresAt.Int64, 0)
		k.ExpiresAt = &t
	}
	k.LastResetAt = time.Unix(lastResetAt, 0)
	k.CreatedAt = time.Unix(createdAt, 0)

	return &k, nil
}

func (r *SQLiteKeyRepository) scanRows(rows *sql.Rows) (*domain.Key, error) {
	return r.scanKey(rows)
}
