package db

import (
	"database/sql"
	"fmt"
	"time"
)

type ConfigStore struct {
	db *sql.DB
}

func NewConfigStore(db *sql.DB) *ConfigStore {
	return &ConfigStore{db: db}
}

func (s *ConfigStore) Get(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`
		SELECT value FROM app_config WHERE key = ?
	`, key).Scan(&value)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get config value: %w", err)
	}

	return value, nil
}

func (s *ConfigStore) Set(key, value string) error {
	now := time.Now().UTC()

	_, err := s.db.Exec(`
		INSERT INTO app_config (key, value, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`, key, value, now, now)

	if err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	return nil
}

func (s *ConfigStore) Delete(key string) error {
	_, err := s.db.Exec(`DELETE FROM app_config WHERE key = ?`, key)
	if err != nil {
		return fmt.Errorf("failed to delete config value: %w", err)
	}
	return nil
}
