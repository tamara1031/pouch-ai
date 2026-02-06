package budget

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
)

// Manager handles budget reservations and refunds.
type Manager struct {
	db *sql.DB
	mu sync.Mutex
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// GetBalance returns the current global budget balance.
func (m *Manager) GetBalance() (float64, error) {
	var valStr string
	err := m.db.QueryRow("SELECT value FROM system_config WHERE key = 'budget_usd'").Scan(&valStr)
	if err == sql.ErrNoRows {
		return 0.0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(valStr, 64)
}

// SetBalance sets the global budget balance.
func (m *Manager) SetBalance(amount float64) error {
	_, err := m.db.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES ('budget_usd', ?)", fmt.Sprintf("%f", amount))
	return err
}

// Reserve attempts to reserve 'amount' from the budget.
// Returns error if insufficient funds.
func (m *Manager) Reserve(amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var valStr string
	err = tx.QueryRow("SELECT value FROM system_config WHERE key = 'budget_usd'").Scan(&valStr)
	current := 0.0
	if err == nil {
		current, _ = strconv.ParseFloat(valStr, 64)
	}

	if current < amount {
		return fmt.Errorf("insufficient funds: budget %.4f, required %.4f", current, amount)
	}

	newBalance := current - amount
	_, err = tx.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES ('budget_usd', ?)", fmt.Sprintf("%f", newBalance))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Refund adds 'amount' back to the budget.
func (m *Manager) Refund(amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var valStr string
	err = tx.QueryRow("SELECT value FROM system_config WHERE key = 'budget_usd'").Scan(&valStr)
	current := 0.0
	if err == nil {
		current, _ = strconv.ParseFloat(valStr, 64)
	}

	newBalance := current + amount
	_, err = tx.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES ('budget_usd', ?)", fmt.Sprintf("%f", newBalance))
	if err != nil {
		return err
	}

	return tx.Commit()
}
