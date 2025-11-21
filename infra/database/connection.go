package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // SQLite driver
)

// Config holds database configuration
type Config struct {
	DatabasePath string
}

// NewConnection creates a new SQLite database connection
// It creates the database file and directory if they don't exist
func NewConnection(config Config) (*sql.DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(config.DatabasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys (important for referential integrity)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency (allows reads during writes)
	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout to 5 seconds (retry instead of immediate error)
	_, err = db.Exec("PRAGMA busy_timeout = 5000")
	if err != nil {
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Configure connection pool for SQLite (single writer optimization)
	db.SetMaxOpenConns(1)    // SQLite only supports 1 concurrent writer
	db.SetMaxIdleConns(1)    // Keep connection alive
	db.SetConnMaxLifetime(0) // Reuse connections indefinitely

	return db, nil
}

// Close closes the database connection
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
