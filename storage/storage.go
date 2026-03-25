package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func LoadObserverStorage() error {
	dbPath, err := observerDbPath()
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open databse: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return err
	}

	DB = db

	return nil
}

func observerDbPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	appDir := filepath.Join(baseDir, "cf-observer")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}

	return filepath.Join(appDir, "cf-observer.db"), nil
}

func initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS hosts (
		name TEXT PRIMARY KEY,
		upstream TEXT NOT NULL UNIQUE,
		api_contract_file TEXT,
		resource_contract_file TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS findings (
		id TEXT PRIMARY KEY,
		source TEXT NOT NULL,
		stage TEXT NOT NULL,
		severity TEXT NOT NULL,
		code TEXT NOT NULL,
		message TEXT NOT NULL,
		request_id TEXT,
		host TEXT,
		path TEXT,
		method TEXT,
		body TEXT,
		operation_id TEXT,
		field TEXT,
		resource TEXT
	);
	`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
