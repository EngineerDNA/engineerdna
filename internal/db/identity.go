package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// IdentityStore provides database operations for engineers and unresolved identities
type IdentityStore struct {
	db *sql.DB
}

// NewIdentityStore creates a new identity store
func NewIdentityStore(db *sql.DB) *IdentityStore {
	return &IdentityStore{db: db}
}

// Engineer operations

// CheckIdentifierExists checks if any identifier already exists for an active engineer (within transaction)
func (s *IdentityStore) CheckIdentifierExists(tx *sql.Tx, identifiers map[string]string) (*models.Engineer, error) {
	for source, identifier := range identifiers {
		var engineer models.Engineer
		var identifiersJSON string

		err := tx.QueryRow(`
			SELECT id, canonical_name, email, manager, identifiers, active, created_at, updated_at
			FROM engineers
			WHERE active = 1
			AND JSON_EXTRACT(identifiers, '$.' || ?) = ?
		`, source, identifier).Scan(&engineer.ID, &engineer.Name, &engineer.Email, &engineer.Manager, &identifiersJSON, &engineer.Active, &engineer.CreatedAt, &engineer.UpdatedAt)

		if err == sql.ErrNoRows {
			continue // No match for this source+identifier, check next
		}
		if err != nil {
			return nil, fmt.Errorf("failed to query engineers: %w", err)
		}

		if err := json.Unmarshal([]byte(identifiersJSON), &engineer.Identifiers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal identifiers: %w", err)
		}

		return &engineer, nil
	}

	return nil, nil
}

// CreateEngineer inserts a new engineer record (with atomic duplicate check)
func (s *IdentityStore) CreateEngineer(engineer *models.Engineer) error {
	if engineer.ID == "" {
		engineer.ID = uuid.New().String()
	}
	if engineer.CreatedAt.IsZero() {
		engineer.CreatedAt = time.Now().UTC()
	}
	engineer.UpdatedAt = time.Now().UTC()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // No-op if committed

	// Check for duplicate identifiers within transaction
	if engineer.Identifiers != nil && len(engineer.Identifiers) > 0 {
		existing, err := s.CheckIdentifierExists(tx, engineer.Identifiers)
		if err != nil {
			return fmt.Errorf("failed to check for duplicate identifier: %w", err)
		}
		if existing != nil {
			return fmt.Errorf("identifier already exists for engineer %s", existing.Name)
		}
	}

	identifiersJSON, err := json.Marshal(engineer.Identifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal identifiers: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO engineers (id, canonical_name, email, manager, identifiers, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, engineer.ID, engineer.Name, engineer.Email, engineer.Manager, string(identifiersJSON), engineer.Active, engineer.CreatedAt, engineer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create engineer: %w", err)
	}

	return tx.Commit()
}

// GetEngineerByID retrieves an engineer by ID
func (s *IdentityStore) GetEngineerByID(id string) (*models.Engineer, error) {
	var engineer models.Engineer
	var identifiersJSON string

	err := s.db.QueryRow(`
		SELECT id, canonical_name, email, manager, identifiers, active, created_at, updated_at
		FROM engineers WHERE id = ?
	`, id).Scan(&engineer.ID, &engineer.Name, &engineer.Email, &engineer.Manager, &identifiersJSON, &engineer.Active, &engineer.CreatedAt, &engineer.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer: %w", err)
	}

	if err := json.Unmarshal([]byte(identifiersJSON), &engineer.Identifiers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal identifiers: %w", err)
	}

	return &engineer, nil
}

// GetEngineerByIdentifier looks up an engineer by source and identifier
func (s *IdentityStore) GetEngineerByIdentifier(source, identifier string) (*models.Engineer, error) {
	var engineer models.Engineer
	var identifiersJSON string

	err := s.db.QueryRow(`
		SELECT id, canonical_name, email, manager, identifiers, active, created_at, updated_at
		FROM engineers
		WHERE active = 1
		AND JSON_EXTRACT(identifiers, '$.' || ?) = ?
	`, source, identifier).Scan(&engineer.ID, &engineer.Name, &engineer.Email, &engineer.Manager, &identifiersJSON, &engineer.Active, &engineer.CreatedAt, &engineer.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query engineers: %w", err)
	}

	if err := json.Unmarshal([]byte(identifiersJSON), &engineer.Identifiers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal identifiers: %w", err)
	}

	return &engineer, nil
}

// ListEngineers retrieves all engineers with optional filters
func (s *IdentityStore) ListEngineers(activeOnly bool, limit, offset int) ([]*models.Engineer, error) {
	query := "SELECT id, canonical_name, email, manager, identifiers, active, created_at, updated_at FROM engineers WHERE 1=1"
	args := []interface{}{}

	if activeOnly {
		query += " AND active = ?"
		args = append(args, true)
	}

	query += " ORDER BY canonical_name ASC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list engineers: %w", err)
	}
	defer rows.Close()

	var engineers []*models.Engineer
	for rows.Next() {
		var engineer models.Engineer
		var identifiersJSON string

		err := rows.Scan(&engineer.ID, &engineer.Name, &engineer.Email, &engineer.Manager, &identifiersJSON, &engineer.Active, &engineer.CreatedAt, &engineer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan engineer: %w", err)
		}

		if err := json.Unmarshal([]byte(identifiersJSON), &engineer.Identifiers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal identifiers: %w", err)
		}

		engineers = append(engineers, &engineer)
	}

	return engineers, nil
}

// CountEngineers returns the total count of engineers with optional filters
func (s *IdentityStore) CountEngineers(activeOnly bool) (int, error) {
	query := "SELECT COUNT(*) FROM engineers WHERE 1=1"
	args := []interface{}{}

	if activeOnly {
		query += " AND active = ?"
		args = append(args, true)
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count engineers: %w", err)
	}

	return count, nil
}

// UpdateEngineer updates an existing engineer record (non-transactional, no duplicate check)
func (s *IdentityStore) UpdateEngineer(engineer *models.Engineer) error {
	engineer.UpdatedAt = time.Now().UTC()

	identifiersJSON, err := json.Marshal(engineer.Identifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal identifiers: %w", err)
	}

	_, err = s.db.Exec(`
		UPDATE engineers SET canonical_name = ?, email = ?, manager = ?, identifiers = ?, active = ?, updated_at = ?
		WHERE id = ?
	`, engineer.Name, engineer.Email, engineer.Manager, string(identifiersJSON), engineer.Active, engineer.UpdatedAt, engineer.ID)

	if err != nil {
		return fmt.Errorf("failed to update engineer: %w", err)
	}

	return nil
}

// UpdateEngineerTx updates an existing engineer record with atomic duplicate identifier check
func (s *IdentityStore) UpdateEngineerTx(engineer *models.Engineer) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // No-op if committed

	// Check for duplicate identifiers within transaction (excluding this engineer)
	if engineer.Identifiers != nil && len(engineer.Identifiers) > 0 {
		for source, identifier := range engineer.Identifiers {
			var existingID string
			var existingName string

			err := tx.QueryRow(`
				SELECT id, canonical_name
				FROM engineers
				WHERE active = 1
				AND id != ?
				AND JSON_EXTRACT(identifiers, '$.' || ?) = ?
			`, engineer.ID, source, identifier).Scan(&existingID, &existingName)

			if err == sql.ErrNoRows {
				continue // No conflict, check next identifier
			}
			if err != nil {
				return fmt.Errorf("failed to check for duplicate identifier: %w", err)
			}

			// Found a duplicate
			return fmt.Errorf("identifier %s:%s already exists for engineer %s", source, identifier, existingName)
		}
	}

	// Update engineer
	engineer.UpdatedAt = time.Now().UTC()

	identifiersJSON, err := json.Marshal(engineer.Identifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal identifiers: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE engineers SET canonical_name = ?, email = ?, manager = ?, identifiers = ?, active = ?, updated_at = ?
		WHERE id = ?
	`, engineer.Name, engineer.Email, engineer.Manager, string(identifiersJSON), engineer.Active, engineer.UpdatedAt, engineer.ID)

	if err != nil {
		return fmt.Errorf("failed to update engineer: %w", err)
	}

	return tx.Commit()
}

// DeleteEngineer soft-deletes an engineer (sets active = false)
func (s *IdentityStore) DeleteEngineer(id string) error {
	_, err := s.db.Exec(`
		UPDATE engineers SET active = ?, updated_at = ? WHERE id = ?
	`, false, time.Now().UTC(), id)

	if err != nil {
		return fmt.Errorf("failed to delete engineer: %w", err)
	}

	return nil
}

// Unresolved identity operations

// CreateUnresolvedIdentity inserts a new unresolved identity
func (s *IdentityStore) CreateUnresolvedIdentity(unresolved *models.UnresolvedIdentity) error {
	if unresolved.ID == "" {
		unresolved.ID = uuid.New().String()
	}
	if unresolved.FirstSeen.IsZero() {
		unresolved.FirstSeen = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO unresolved_identities (id, source, identifier, first_seen, event_count, ignored)
		VALUES (?, ?, ?, ?, ?, ?)
	`, unresolved.ID, unresolved.Source, unresolved.Identifier, unresolved.FirstSeen, unresolved.EventCount, unresolved.Ignored)

	if err != nil {
		return fmt.Errorf("failed to create unresolved identity: %w", err)
	}

	return nil
}

// GetUnresolvedByID retrieves an unresolved identity by ID
func (s *IdentityStore) GetUnresolvedByID(id string) (*models.UnresolvedIdentity, error) {
	var unresolved models.UnresolvedIdentity

	err := s.db.QueryRow(`
		SELECT id, source, identifier, first_seen, event_count, ignored
		FROM unresolved_identities WHERE id = ?
	`, id).Scan(&unresolved.ID, &unresolved.Source, &unresolved.Identifier, &unresolved.FirstSeen, &unresolved.EventCount, &unresolved.Ignored)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get unresolved identity: %w", err)
	}

	return &unresolved, nil
}

// GetUnresolvedBySourceIdentifier retrieves an unresolved identity by source and identifier
func (s *IdentityStore) GetUnresolvedBySourceIdentifier(source, identifier string) (*models.UnresolvedIdentity, error) {
	var unresolved models.UnresolvedIdentity

	err := s.db.QueryRow(`
		SELECT id, source, identifier, first_seen, event_count, ignored
		FROM unresolved_identities WHERE source = ? AND identifier = ?
	`, source, identifier).Scan(&unresolved.ID, &unresolved.Source, &unresolved.Identifier, &unresolved.FirstSeen, &unresolved.EventCount, &unresolved.Ignored)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get unresolved identity: %w", err)
	}

	return &unresolved, nil
}

// IncrementUnresolvedEventCount increments the event count for an unresolved identity
func (s *IdentityStore) IncrementUnresolvedEventCount(source, identifier string) error {
	_, err := s.db.Exec(`
		UPDATE unresolved_identities SET event_count = event_count + 1
		WHERE source = ? AND identifier = ?
	`, source, identifier)

	if err != nil {
		return fmt.Errorf("failed to increment event count: %w", err)
	}

	return nil
}

// ListUnresolvedIdentities retrieves all non-ignored unresolved identities
func (s *IdentityStore) ListUnresolvedIdentities(limit int) ([]*models.UnresolvedIdentity, error) {
	rows, err := s.db.Query(`
		SELECT id, source, identifier, first_seen, event_count, ignored
		FROM unresolved_identities
		WHERE ignored = 0
		ORDER BY event_count DESC, first_seen DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list unresolved identities: %w", err)
	}
	defer rows.Close()

	var unresolved []*models.UnresolvedIdentity
	for rows.Next() {
		var u models.UnresolvedIdentity
		err := rows.Scan(&u.ID, &u.Source, &u.Identifier, &u.FirstSeen, &u.EventCount, &u.Ignored)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unresolved identity: %w", err)
		}
		unresolved = append(unresolved, &u)
	}

	return unresolved, nil
}

// ListIgnoredIdentities retrieves all ignored unresolved identities
func (s *IdentityStore) ListIgnoredIdentities(limit int) ([]*models.UnresolvedIdentity, error) {
	rows, err := s.db.Query(`
		SELECT id, source, identifier, first_seen, event_count, ignored
		FROM unresolved_identities
		WHERE ignored = 1
		ORDER BY first_seen DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list ignored identities: %w", err)
	}
	defer rows.Close()

	var ignored []*models.UnresolvedIdentity
	for rows.Next() {
		var u models.UnresolvedIdentity
		err := rows.Scan(&u.ID, &u.Source, &u.Identifier, &u.FirstSeen, &u.EventCount, &u.Ignored)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ignored identity: %w", err)
		}
		ignored = append(ignored, &u)
	}

	return ignored, nil
}

// DeleteUnresolvedIdentity removes an unresolved identity (after resolution)
func (s *IdentityStore) DeleteUnresolvedIdentity(source, identifier string) error {
	_, err := s.db.Exec(`
		DELETE FROM unresolved_identities WHERE source = ? AND identifier = ?
	`, source, identifier)

	if err != nil {
		return fmt.Errorf("failed to delete unresolved identity: %w", err)
	}

	return nil
}

// SetUnresolvedIgnored sets the ignored status for an unresolved identity
func (s *IdentityStore) SetUnresolvedIgnored(id string, ignored bool) error {
	_, err := s.db.Exec(`
		UPDATE unresolved_identities
		SET ignored = ?
		WHERE id = ?
	`, ignored, id)

	if err != nil {
		return fmt.Errorf("failed to set ignored status: %w", err)
	}

	return nil
}

// GetEngineerActivityMetrics retrieves activity metrics for an engineer over a time period
func (s *IdentityStore) GetEngineerActivityMetrics(engineerID string, periodStart, periodEnd time.Time) (*models.ActivityMetrics, error) {
	engineer, err := s.GetEngineerByID(engineerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer: %w", err)
	}
	if engineer == nil {
		return nil, fmt.Errorf("engineer not found: %s", engineerID)
	}

	metrics := &models.ActivityMetrics{
		EngineerID:   engineerID,
		EngineerName: engineer.Name,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}

	// Count events by type (with GROUP_CONCAT limit to prevent memory exhaustion)
	rows, err := s.db.Query(`
		SELECT type, COUNT(*) as count,
		       SUBSTR(GROUP_CONCAT(DISTINCT source), 1, 500) as sources
		FROM events
		WHERE engineer_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY type
	`, engineerID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to query event counts: %w", err)
	}
	defer rows.Close()

	uniqueSourcesMap := make(map[string]bool)
	for rows.Next() {
		var eventType string
		var count int
		var sources string
		err := rows.Scan(&eventType, &count, &sources)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event count: %w", err)
		}

		// Map event types to metrics
		switch eventType {
		case "pull_request":
			metrics.PullRequests = count
		case "issue":
			metrics.Issues = count
		case "commit":
			metrics.Commits = count
		case "review":
			metrics.CodeReviews = count
		}

		metrics.TotalEvents += count

		// Track unique sources
		// SQLite GROUP_CONCAT with comma separator
		if sources != "" {
			for _, source := range splitSources(sources) {
				uniqueSourcesMap[source] = true
			}
		}
	}

	// Convert unique sources map to slice
	for source := range uniqueSourcesMap {
		metrics.UniqueSources = append(metrics.UniqueSources, source)
	}

	return metrics, nil
}

// UpdateEventEngineerID updates the engineer_id for all events matching source and identifier
func (s *IdentityStore) UpdateEventEngineerID(engineerID, source, identifier string) (int64, error) {
	result, err := s.db.Exec(`
		UPDATE events SET engineer_id = ?, updated_at = ?
		WHERE source = ? AND actor = ?
	`, engineerID, time.Now().UTC(), source, identifier)

	if err != nil {
		return 0, fmt.Errorf("failed to update event engineer_id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// AssignIdentityTx assigns an identity to an engineer in a transaction
func (s *IdentityStore) AssignIdentityTx(engineer *models.Engineer, source, identifier string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // No-op if committed

	// Update engineer identifiers
	engineer.UpdatedAt = time.Now().UTC()
	identifiersJSON, err := json.Marshal(engineer.Identifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal identifiers: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE engineers SET canonical_name = ?, email = ?, manager = ?, identifiers = ?, active = ?, updated_at = ?
		WHERE id = ?
	`, engineer.Name, engineer.Email, engineer.Manager, string(identifiersJSON), engineer.Active, engineer.UpdatedAt, engineer.ID)
	if err != nil {
		return fmt.Errorf("failed to update engineer: %w", err)
	}

	// Update all events with this source+identifier to reference the engineer
	_, err = tx.Exec(`
		UPDATE events SET engineer_id = ?, updated_at = ?
		WHERE source = ? AND actor = ?
	`, engineer.ID, time.Now().UTC(), source, identifier)
	if err != nil {
		return fmt.Errorf("failed to update events: %w", err)
	}

	// Delete unresolved identity
	_, err = tx.Exec(`
		DELETE FROM unresolved_identities WHERE source = ? AND identifier = ?
	`, source, identifier)
	if err != nil {
		return fmt.Errorf("failed to delete unresolved identity: %w", err)
	}

	return tx.Commit()
}

// CreateEngineerFromUnresolvedTx creates a new engineer, updates events, and deletes unresolved identity in a transaction
func (s *IdentityStore) CreateEngineerFromUnresolvedTx(engineer *models.Engineer, source, identifier string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // No-op if committed

	// Check for duplicate identifiers within transaction
	if engineer.Identifiers != nil && len(engineer.Identifiers) > 0 {
		existing, err := s.CheckIdentifierExists(tx, engineer.Identifiers)
		if err != nil {
			return fmt.Errorf("failed to check for duplicate identifier: %w", err)
		}
		if existing != nil {
			return fmt.Errorf("identifier already exists for engineer %s", existing.Name)
		}
	}

	// Create engineer
	if engineer.ID == "" {
		engineer.ID = uuid.New().String()
	}
	if engineer.CreatedAt.IsZero() {
		engineer.CreatedAt = time.Now().UTC()
	}
	engineer.UpdatedAt = time.Now().UTC()

	identifiersJSON, err := json.Marshal(engineer.Identifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal identifiers: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO engineers (id, canonical_name, email, manager, identifiers, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, engineer.ID, engineer.Name, engineer.Email, engineer.Manager, string(identifiersJSON), engineer.Active, engineer.CreatedAt, engineer.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create engineer: %w", err)
	}

	// Update all events with this source+identifier to reference the new engineer
	_, err = tx.Exec(`
		UPDATE events SET engineer_id = ?, updated_at = ?
		WHERE source = ? AND actor = ?
	`, engineer.ID, time.Now().UTC(), source, identifier)
	if err != nil {
		return fmt.Errorf("failed to update events: %w", err)
	}

	// Delete the unresolved identity
	_, err = tx.Exec(`
		DELETE FROM unresolved_identities WHERE source = ? AND identifier = ?
	`, source, identifier)
	if err != nil {
		return fmt.Errorf("failed to delete unresolved identity: %w", err)
	}

	return tx.Commit()
}

// MergeEngineersTx merges two engineers in a transaction
func (s *IdentityStore) MergeEngineersTx(keepEngineer, mergeEngineer *models.Engineer) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // No-op if committed

	// Update keep engineer with merged identifiers
	keepEngineer.UpdatedAt = time.Now().UTC()
	identifiersJSON, err := json.Marshal(keepEngineer.Identifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal identifiers: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE engineers SET canonical_name = ?, email = ?, manager = ?, identifiers = ?, active = ?, updated_at = ?
		WHERE id = ?
	`, keepEngineer.Name, keepEngineer.Email, keepEngineer.Manager, string(identifiersJSON), keepEngineer.Active, keepEngineer.UpdatedAt, keepEngineer.ID)
	if err != nil {
		return fmt.Errorf("failed to update keep engineer: %w", err)
	}

	// Update all events from merge engineer to point to keep engineer
	for source, identifier := range mergeEngineer.Identifiers {
		_, err := tx.Exec(`
			UPDATE events SET engineer_id = ?, updated_at = ?
			WHERE source = ? AND actor = ?
		`, keepEngineer.ID, time.Now().UTC(), source, identifier)
		if err != nil {
			return fmt.Errorf("failed to update events for %s:%s: %w", source, identifier, err)
		}
	}

	// Soft-delete merge engineer
	_, err = tx.Exec(`
		UPDATE engineers SET active = ?, updated_at = ? WHERE id = ?
	`, false, time.Now().UTC(), mergeEngineer.ID)
	if err != nil {
		return fmt.Errorf("failed to delete merge engineer: %w", err)
	}

	return tx.Commit()
}

// Helper function to split GROUP_CONCAT result
func splitSources(sources string) []string {
	// SQLite GROUP_CONCAT uses comma separator by default
	result := []string{}
	if sources == "" {
		return result
	}

	var current string
	for i := 0; i < len(sources); i++ {
		if sources[i] == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(sources[i])
		}
	}
	if current != "" {
		result = append(result, current)
	}

	return result
}
