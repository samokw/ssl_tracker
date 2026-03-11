package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// InitSQLite initializes the SQLite database connection
func InitSQLite(dbPath string) (*sql.DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
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

	if err := addDiscordWebhookColumn(db); err != nil {
		return err
	}

	notificationsTable := `
	CREATE TABLE IF NOT EXISTS notifications (
		id                INTEGER PRIMARY KEY AUTOINCREMENT,
		domain_id         INTEGER NOT NULL,
		notification_type TEXT NOT NULL,
		event_type        TEXT NOT NULL,
		days_before       INTEGER,
		condition_key     TEXT NOT NULL,
		sent_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(notificationsTable); err != nil {
		return fmt.Errorf("failed to create notifications table: %w", err)
	}

	notificationsIndex := `
	CREATE UNIQUE INDEX IF NOT EXISTS uq_notification_dedup
		ON notifications (domain_id, notification_type, event_type, days_before, condition_key);`

	if _, err := db.Exec(notificationsIndex); err != nil {
		return fmt.Errorf("failed to create notifications dedup index: %w", err)
	}

	return nil
}

// since sqlite does not support ADD COLUMN IF NOT EXISTS, we need to check the metadata to see if the column exists.
func addDiscordWebhookColumn(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info('users')`)
	if err != nil {
		return fmt.Errorf("failed to get query user table info: %w", err)
	}
	defer rows.Close()

	//
	for rows.Next() {
		var colID int
		var name, colType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&colID, &name, &colType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("failed to scan table info row: %w", err)
		}
		// if the row already exists
		if name == "discord_webhook_url" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate table info: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE users ADD COLUMN discord_webhook_url TEXT`); err != nil {
		return fmt.Errorf("failed to add discord_webhook_url column: %w", err)
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
