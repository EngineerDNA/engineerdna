package config

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Set WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
