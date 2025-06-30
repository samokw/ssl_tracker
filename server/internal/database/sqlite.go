package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// InitSQLite initializes the SQLite database connection
func InitSQLite(dbPath string) (*sql.DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	domainsTable := `
	CREATE TABLE IF NOT EXISTS domains (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		domain_name TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		expiry_date DATETIME,
		last_checked DATETIME,
		last_error TEXT,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		UNIQUE(user_id, domain_name)
	);`

	if _, err := db.Exec(domainsTable); err != nil {
		return fmt.Errorf("failed to create domains table: %w", err)
	}

	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	defaultUser := `INSERT OR IGNORE INTO users (id, username) VALUES (1, 'default');`
	if _, err := db.Exec(defaultUser); err != nil {
		return fmt.Errorf("failed to insert default user: %w", err)
	}

	return nil
}

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "sslcerttop"), nil
}

func GetDefaultDBPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sslcerttop.db"), nil
}
