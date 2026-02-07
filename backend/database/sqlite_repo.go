package database

import (
	"context"
	"database/sql"
	"encoding/json"
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var expiresAt *int64
	if k.ExpiresAt != nil {
		val := k.ExpiresAt.Unix()
		expiresAt = &val
	}

	// Marshal provider config to JSON
	var providerConfig string
	if k.Configuration != nil && k.Configuration.Provider.Config != nil {
		b, _ := json.Marshal(k.Configuration.Provider.Config)
		providerConfig = string(b)
	}

	providerID := "openai"
	budgetLimit := 0.0
	resetPeriod := 0
	if k.Configuration != nil {
		providerID = k.Configuration.Provider.ID
		budgetLimit = k.Configuration.BudgetLimit
		resetPeriod = k.Configuration.ResetPeriod
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO app_keys (name, key_hash, prefix, expires_at, budget_usage, last_reset_at, created_at, provider_id, provider_config, budget_limit, reset_period)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, k.Name, k.KeyHash, k.Prefix, expiresAt, k.BudgetUsage, k.LastResetAt.Unix(), k.CreatedAt.Unix(), providerID, providerConfig, budgetLimit, resetPeriod)

	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	k.ID = domain.ID(id)

	// Insert middlewares
	if k.Configuration != nil {
		for i, mw := range k.Configuration.Middlewares {
			var mwConfig string
			if mw.Config != nil {
				b, _ := json.Marshal(mw.Config)
				mwConfig = string(b)
			}
			_, err = tx.ExecContext(ctx, `
				INSERT INTO app_key_middlewares (app_key_id, middleware_id, config, priority)
				VALUES (?, ?, ?, ?)
			`, k.ID, mw.ID, mwConfig, i)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *SQLiteKeyRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Key, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, key_hash, prefix, expires_at, budget_usage, last_reset_at, created_at,
		       provider_id, provider_config, budget_limit, reset_period
		FROM app_keys WHERE id = ?
	`, id)

	k, err := r.scanKey(row)
	if err != nil || k == nil {
		return k, err
	}

	return r.loadMiddlewares(ctx, k)
}

func (r *SQLiteKeyRepository) GetByHash(ctx context.Context, hash string) (*domain.Key, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, key_hash, prefix, expires_at, budget_usage, last_reset_at, created_at,
		       provider_id, provider_config, budget_limit, reset_period
		FROM app_keys WHERE key_hash = ?
	`, hash)

	k, err := r.scanKey(row)
	if err != nil || k == nil {
		return k, err
	}

	return r.loadMiddlewares(ctx, k)
}

func (r *SQLiteKeyRepository) List(ctx context.Context) ([]*domain.Key, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, key_hash, prefix, expires_at, budget_usage, last_reset_at, created_at,
		       provider_id, provider_config, budget_limit, reset_period
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

	// Load middlewares for all keys
	for _, k := range keys {
		if _, err := r.loadMiddlewares(ctx, k); err != nil {
			return nil, err
		}
	}

	return keys, nil
}

func (r *SQLiteKeyRepository) Update(ctx context.Context, k *domain.Key) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var providerConfig string
	if k.Configuration != nil && k.Configuration.Provider.Config != nil {
		b, _ := json.Marshal(k.Configuration.Provider.Config)
		providerConfig = string(b)
	}

	providerID := "openai"
	budgetLimit := 0.0
	resetPeriod := 0
	if k.Configuration != nil {
		providerID = k.Configuration.Provider.ID
		budgetLimit = k.Configuration.BudgetLimit
		resetPeriod = k.Configuration.ResetPeriod
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE app_keys 
		SET name = ?, provider_id = ?, provider_config = ?, budget_limit = ?, reset_period = ?
		WHERE id = ?
	`, k.Name, providerID, providerConfig, budgetLimit, resetPeriod, k.ID)
	if err != nil {
		return err
	}

	// Delete old middlewares and insert new ones
	_, err = tx.ExecContext(ctx, `DELETE FROM app_key_middlewares WHERE app_key_id = ?`, k.ID)
	if err != nil {
		return err
	}

	if k.Configuration != nil {
		for i, mw := range k.Configuration.Middlewares {
			var mwConfig string
			if mw.Config != nil {
				b, _ := json.Marshal(mw.Config)
				mwConfig = string(b)
			}
			_, err = tx.ExecContext(ctx, `
				INSERT INTO app_key_middlewares (app_key_id, middleware_id, config, priority)
				VALUES (?, ?, ?, ?)
			`, k.ID, mw.ID, mwConfig, i)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *SQLiteKeyRepository) Delete(ctx context.Context, id domain.ID) error {
	// CASCADE will handle middleware deletion
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
	var providerID string
	var providerConfig sql.NullString
	var budgetLimit float64
	var resetPeriod int

	err := sc.Scan(
		&k.ID, &k.Name, &k.KeyHash, &k.Prefix, &expiresAt,
		&k.BudgetUsage, &lastResetAt, &createdAt,
		&providerID, &providerConfig, &budgetLimit, &resetPeriod,
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

	// Build configuration from columns
	k.Configuration = &domain.KeyConfiguration{
		Provider: domain.PluginConfig{
			ID: providerID,
		},
		BudgetLimit: budgetLimit,
		ResetPeriod: resetPeriod,
	}

	if providerConfig.Valid && providerConfig.String != "" {
		var cfg map[string]any
		if err := json.Unmarshal([]byte(providerConfig.String), &cfg); err == nil {
			k.Configuration.Provider.Config = cfg
		}
	}

	return &k, nil
}

func (r *SQLiteKeyRepository) scanRows(rows *sql.Rows) (*domain.Key, error) {
	return r.scanKey(rows)
}

func (r *SQLiteKeyRepository) loadMiddlewares(ctx context.Context, k *domain.Key) (*domain.Key, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT middleware_id, config FROM app_key_middlewares 
		WHERE app_key_id = ? ORDER BY priority ASC
	`, k.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var middlewares []domain.PluginConfig
	for rows.Next() {
		var mwID string
		var mwConfig sql.NullString
		if err := rows.Scan(&mwID, &mwConfig); err != nil {
			return nil, err
		}

		mw := domain.PluginConfig{ID: mwID}
		if mwConfig.Valid && mwConfig.String != "" {
			var cfg map[string]any
			if err := json.Unmarshal([]byte(mwConfig.String), &cfg); err == nil {
				mw.Config = cfg
			}
		}
		middlewares = append(middlewares, mw)
	}

	if k.Configuration != nil {
		k.Configuration.Middlewares = middlewares
	}

	return k, nil
}
