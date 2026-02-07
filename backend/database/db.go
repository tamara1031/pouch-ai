package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // CGO-free SQLite driver
)

var DB *sql.DB

// InitDB initializes the SQLite database with WAL mode enabled.
func InitDB(dataDir string) error {
	dbPath := filepath.Join(dataDir, "pouch.db")

	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Enable WAL mode, foreign keys, and busy timeout
	dsn := fmt.Sprintf("%s?_pragma=journal_mode=WAL&_pragma=busy_timeout=5000&_pragma=foreign_keys=ON", dbPath)
	var err error
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to SQLite database at %s (WAL mode enabled)", dbPath)

	return migrate(DB)
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS app_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL UNIQUE,
		prefix TEXT NOT NULL,
		expires_at INTEGER,
		budget_usage REAL DEFAULT 0,
		last_reset_at INTEGER,
		created_at INTEGER NOT NULL,
		-- Provider (1:1, embedded)
		provider_id TEXT NOT NULL DEFAULT 'openai',
		provider_config TEXT,
		-- Budget settings
		budget_limit REAL DEFAULT 0,
		reset_period INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS app_key_middlewares (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_key_id INTEGER NOT NULL REFERENCES app_keys(id) ON DELETE CASCADE,
		middleware_id TEXT NOT NULL,
		config TEXT,
		priority INTEGER NOT NULL DEFAULT 0,
		UNIQUE(app_key_id, middleware_id)
	);

	CREATE INDEX IF NOT EXISTS idx_middlewares_key ON app_key_middlewares(app_key_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
