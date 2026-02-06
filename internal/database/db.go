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
	CREATE TABLE IF NOT EXISTS app_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		provider TEXT NOT NULL DEFAULT 'openai',
		key_hash TEXT NOT NULL UNIQUE,
		prefix TEXT NOT NULL,
		expires_at INTEGER,
		budget_limit REAL DEFAULT 0,
		budget_usage REAL DEFAULT 0,
		budget_period TEXT,
		last_reset_at INTEGER,
		is_mock BOOLEAN DEFAULT 0,
		mock_config TEXT,
		rate_limit INTEGER DEFAULT 10,
		rate_period TEXT DEFAULT 'minute',
		created_at INTEGER NOT NULL
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Add columns if they don't exist (for existing DBs)
	addColumn(db, "app_keys", "provider", "TEXT NOT NULL DEFAULT 'openai'")
	addColumn(db, "app_keys", "rate_limit", "INTEGER DEFAULT 10")
	addColumn(db, "app_keys", "rate_period", "TEXT DEFAULT 'minute'")

	return nil
}

func addColumn(db *sql.DB, table, column, definition string) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", table)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Failed to check table info for %s: %v", table, err)
		return
	}
	defer rows.Close()

	exists := false
	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err == nil {
			if name == column {
				exists = true
				break
			}
		}
	}

	if !exists {
		alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
		if _, err := db.Exec(alterQuery); err != nil {
			log.Printf("Failed to add column %s to %s: %v", column, table, err)
		} else {
			log.Printf("Added column %s to table %s", column, table)
		}
	}
}
