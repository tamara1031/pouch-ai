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

	dsn := fmt.Sprintf("%s?_pragma=journal_mode=WAL&_pragma=busy_timeout=5000", dbPath)
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
	// Simple schema migration
	schema := `
	CREATE TABLE IF NOT EXISTS credentials (
		provider TEXT PRIMARY KEY,
		encrypted_key TEXT NOT NULL,
		salt TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS audit_logs (
		request_id TEXT PRIMARY KEY,
		timestamp INTEGER NOT NULL,
		model TEXT NOT NULL,
		input_tokens INTEGER,
		output_tokens INTEGER,
		total_cost REAL
	);

	CREATE TABLE IF NOT EXISTS system_config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	
	-- Index for seeking logs by time
	CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);

	CREATE TABLE IF NOT EXISTS app_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL UNIQUE,
		prefix TEXT NOT NULL,
		expires_at INTEGER,
		budget_limit REAL DEFAULT 0,
		budget_usage REAL DEFAULT 0,
		created_at INTEGER NOT NULL
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
