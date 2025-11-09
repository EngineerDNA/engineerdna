package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

type AnonymizationStore struct {
	db *sql.DB
}

func NewAnonymizationStore(db *sql.DB) *AnonymizationStore {
	return &AnonymizationStore{db: db}
}

// Mapping operations
func (s *AnonymizationStore) CreateMapping(mapping *models.AnonymizationMapping) error {
	if mapping.ID == "" {
		mapping.ID = uuid.New().String()
	}
	if mapping.CreatedAt.IsZero() {
		mapping.CreatedAt = time.Now().UTC()
	}

	extIDsJSON, err := json.Marshal(mapping.ExternalIdentifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal external identifiers: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO anonymization_map (id, anonymized_id, real_name, real_email, external_identifiers, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, mapping.ID, mapping.AnonymizedID, mapping.RealName, mapping.RealEmail, string(extIDsJSON), mapping.Enabled, mapping.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create mapping: %w", err)
	}

	return nil
}

func (s *AnonymizationStore) GetMappingByReal(realIdentifier string) (*models.AnonymizationMapping, error) {
	var mapping models.AnonymizationMapping
	var extIDsJSON string
	var realEmail sql.NullString

	err := s.db.QueryRow(`
		SELECT id, anonymized_id, real_name, real_email, external_identifiers, enabled, created_at
		FROM anonymization_map WHERE real_name = ? OR real_email = ?
	`, realIdentifier, realIdentifier).Scan(&mapping.ID, &mapping.AnonymizedID, &mapping.RealName, &realEmail, &extIDsJSON, &mapping.Enabled, &mapping.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}

	if realEmail.Valid {
		mapping.RealEmail = realEmail.String
	}

	if err := json.Unmarshal([]byte(extIDsJSON), &mapping.ExternalIdentifiers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal external identifiers: %w", err)
	}

	return &mapping, nil
}

func (s *AnonymizationStore) GetMappingByAnonymized(anonymizedID string) (*models.AnonymizationMapping, error) {
	var mapping models.AnonymizationMapping
	var extIDsJSON string
	var realEmail sql.NullString

	err := s.db.QueryRow(`
		SELECT id, anonymized_id, real_name, real_email, external_identifiers, enabled, created_at
		FROM anonymization_map WHERE anonymized_id = ?
	`, anonymizedID).Scan(&mapping.ID, &mapping.AnonymizedID, &mapping.RealName, &realEmail, &extIDsJSON, &mapping.Enabled, &mapping.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}

	if realEmail.Valid {
		mapping.RealEmail = realEmail.String
	}

	if err := json.Unmarshal([]byte(extIDsJSON), &mapping.ExternalIdentifiers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal external identifiers: %w", err)
	}

	return &mapping, nil
}

func (s *AnonymizationStore) ListMappings(limit int) ([]*models.AnonymizationMapping, error) {
	// Cap at DefaultQueryLimit to prevent memory exhaustion
	if limit <= 0 || limit > DefaultQueryLimit {
		limit = DefaultQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, anonymized_id, real_name, real_email, external_identifiers, enabled, created_at
		FROM anonymization_map ORDER BY created_at LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list mappings: %w", err)
	}
	defer rows.Close()

	var mappings []*models.AnonymizationMapping
	for rows.Next() {
		var mapping models.AnonymizationMapping
		var extIDsJSON string
		var realEmail sql.NullString

		err := rows.Scan(&mapping.ID, &mapping.AnonymizedID, &mapping.RealName, &realEmail, &extIDsJSON, &mapping.Enabled, &mapping.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}

		if realEmail.Valid {
			mapping.RealEmail = realEmail.String
		}

		if err := json.Unmarshal([]byte(extIDsJSON), &mapping.ExternalIdentifiers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal external identifiers: %w", err)
		}

		mappings = append(mappings, &mapping)
	}

	return mappings, nil
}

func (s *AnonymizationStore) CountMappings() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM anonymization_map").Scan(&count)
	return count, err
}

// CreateMappingAtomic creates a mapping with atomic count for sequential ID generation
func (s *AnonymizationStore) CreateMappingAtomic(mapping *models.AnonymizationMapping, strategy models.AnonymizationStrategy) error {
	// Get dedicated connection from pool to ensure all operations use same connection
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultTimeout)
	defer cancel()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Close()

	// SQLite: Use BEGIN IMMEDIATE for proper write locking from the start
	// This prevents concurrent transactions from both reading the same count
	_, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE TRANSACTION")
	if err != nil {
		return fmt.Errorf("failed to begin immediate transaction: %w", err)
	}

	// Ensure rollback on error
	committed := false
	defer func() {
		if !committed {
			conn.ExecContext(ctx, "ROLLBACK")
		}
	}()

	// For sequential strategy, get count (now with exclusive write lock on this connection)
	if strategy == models.StrategySequential {
		var count int
		err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM anonymization_map").Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count mappings: %w", err)
		}

		// Generate sequential ID based on locked count
		if count < 26 {
			mapping.AnonymizedID = fmt.Sprintf("Engineer_%c", 'A'+count)
		} else {
			mapping.AnonymizedID = fmt.Sprintf("Engineer_%d", count+1)
		}
	}

	// Insert the mapping on the same connection
	extIDsJSON, err := json.Marshal(mapping.ExternalIdentifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal external identifiers: %w", err)
	}

	_, err = conn.ExecContext(ctx, `
		INSERT INTO anonymization_map (id, anonymized_id, real_name, real_email, external_identifiers, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, mapping.ID, mapping.AnonymizedID, mapping.RealName, mapping.RealEmail, string(extIDsJSON), mapping.Enabled, mapping.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert mapping: %w", err)
	}

	// Commit the transaction on the same connection
	_, err = conn.ExecContext(ctx, "COMMIT")
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}

// Policy operations
func (s *AnonymizationStore) SavePolicy(policy *models.AnonymizationPolicy) error {
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now().UTC()
	}
	policy.UpdatedAt = time.Now().UTC()

	fieldsJSON, err := json.Marshal(policy.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO anonymization_policies (plugin_name, enabled, strategy, fields, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(plugin_name) DO UPDATE SET
			enabled = excluded.enabled,
			strategy = excluded.strategy,
			fields = excluded.fields,
			updated_at = excluded.updated_at
	`, policy.PluginName, policy.Enabled, policy.Strategy, string(fieldsJSON), policy.CreatedAt, policy.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save policy: %w", err)
	}

	return nil
}

func (s *AnonymizationStore) GetPolicy(pluginName string) (*models.AnonymizationPolicy, error) {
	var policy models.AnonymizationPolicy
	var fieldsJSON string

	err := s.db.QueryRow(`
		SELECT plugin_name, enabled, strategy, fields, created_at, updated_at
		FROM anonymization_policies WHERE plugin_name = ?
	`, pluginName).Scan(&policy.PluginName, &policy.Enabled, &policy.Strategy, &fieldsJSON, &policy.CreatedAt, &policy.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if err := json.Unmarshal([]byte(fieldsJSON), &policy.Fields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
	}

	return &policy, nil
}

func (s *AnonymizationStore) ListPolicies(limit int) ([]*models.AnonymizationPolicy, error) {
	// Cap at DefaultQueryLimit to prevent memory exhaustion
	if limit <= 0 || limit > DefaultQueryLimit {
		limit = DefaultQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT plugin_name, enabled, strategy, fields, created_at, updated_at
		FROM anonymization_policies ORDER BY plugin_name LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	defer rows.Close()

	var policies []*models.AnonymizationPolicy
	for rows.Next() {
		var policy models.AnonymizationPolicy
		var fieldsJSON string

		err := rows.Scan(&policy.PluginName, &policy.Enabled, &policy.Strategy, &fieldsJSON, &policy.CreatedAt, &policy.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		if err := json.Unmarshal([]byte(fieldsJSON), &policy.Fields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}
