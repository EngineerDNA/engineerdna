package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// AttributeStore handles database operations for entity attributes
type AttributeStore struct {
	db *sql.DB
}

// NewAttributeStore creates a new attribute store
func NewAttributeStore(db *sql.DB) *AttributeStore {
	return &AttributeStore{db: db}
}

// Create inserts a new entity attribute
func (s *AttributeStore) Create(attr *models.EntityAttribute) error {
	if attr.ID == "" {
		attr.ID = uuid.New().String()
	}
	if attr.CreatedAt.IsZero() {
		attr.CreatedAt = time.Now().UTC()
	}

	var validUntil interface{}
	if attr.ValidUntil != nil {
		validUntil = *attr.ValidUntil
	}

	_, err := s.db.Exec(`
		INSERT INTO entity_attributes (id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, attr.ID, attr.EntityType, attr.EntityID, attr.AttributeName, attr.Value, attr.ValueType, attr.ValidFrom, validUntil, attr.Source, attr.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create attribute: %w", err)
	}

	return nil
}

// GetByID retrieves an attribute by ID
func (s *AttributeStore) GetByID(id string) (*models.EntityAttribute, error) {
	var attr models.EntityAttribute
	var validFromStr, createdAtStr string
	var validUntil sql.NullString

	err := s.db.QueryRow(`
		SELECT id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at
		FROM entity_attributes WHERE id = ?
	`, id).Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &validFromStr, &validUntil, &attr.Source, &createdAtStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	// Parse timestamps
	attr.ValidFrom, _ = parseTimestamp(validFromStr)
	attr.CreatedAt, _ = parseTimestamp(createdAtStr)

	if validUntil.Valid {
		t, _ := parseTimestamp(validUntil.String)
		attr.ValidUntil = &t
	}

	return &attr, nil
}

// List retrieves attributes with filters
func (s *AttributeStore) List(filters map[string]interface{}, limit, offset int) ([]*models.EntityAttribute, error) {
	query := "SELECT id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at FROM entity_attributes WHERE 1=1"
	args := []interface{}{}

	if entityType, ok := filters["entity_type"].(string); ok {
		query += " AND entity_type = ?"
		args = append(args, entityType)
	}
	if entityID, ok := filters["entity_id"].(string); ok {
		query += " AND entity_id = ?"
		args = append(args, entityID)
	}
	if attributeName, ok := filters["attribute_name"].(string); ok {
		query += " AND attribute_name = ?"
		args = append(args, attributeName)
	}
	if asOf, ok := filters["as_of"].(time.Time); ok {
		query += " AND valid_from <= ? AND (valid_until IS NULL OR valid_until > ?)"
		args = append(args, asOf, asOf)
	}

	query += " ORDER BY valid_from DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list attributes: %w", err)
	}
	defer rows.Close()

	var attributes []*models.EntityAttribute
	for rows.Next() {
		var attr models.EntityAttribute
		var validFromStr, createdAtStr string
		var validUntil sql.NullString

		err := rows.Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &validFromStr, &validUntil, &attr.Source, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", err)
		}

		// Parse timestamps
		attr.ValidFrom, _ = parseTimestamp(validFromStr)
		attr.CreatedAt, _ = parseTimestamp(createdAtStr)

		if validUntil.Valid {
			t, _ := parseTimestamp(validUntil.String)
			attr.ValidUntil = &t
		}

		attributes = append(attributes, &attr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attributes: %w", err)
	}

	return attributes, nil
}

// GetCurrentAttributes retrieves current attributes for an entity
func (s *AttributeStore) GetCurrentAttributes(entityType, entityID string) ([]*models.EntityAttribute, error) {
	rows, err := s.db.Query(`
		SELECT id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at
		FROM entity_attributes
		WHERE entity_type = ? AND entity_id = ? AND (valid_until IS NULL OR valid_until > ?)
		ORDER BY attribute_name, valid_from DESC
	`, entityType, entityID, time.Now().UTC())

	if err != nil {
		return nil, fmt.Errorf("failed to query current attributes: %w", err)
	}
	defer rows.Close()

	var attributes []*models.EntityAttribute
	for rows.Next() {
		var attr models.EntityAttribute
		var validFromStr, createdAtStr string
		var validUntil sql.NullString

		err := rows.Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &validFromStr, &validUntil, &attr.Source, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", err)
		}

		// Parse timestamps
		attr.ValidFrom, _ = parseTimestamp(validFromStr)
		attr.CreatedAt, _ = parseTimestamp(createdAtStr)

		if validUntil.Valid {
			t, _ := parseTimestamp(validUntil.String)
			attr.ValidUntil = &t
		}

		attributes = append(attributes, &attr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attributes: %w", err)
	}

	return attributes, nil
}

// GetEntityAttributes retrieves attributes for an entity by name
// Supports optional attribute name filtering
func (s *AttributeStore) GetEntityAttributes(entityType, entityID, attributeName string) ([]*models.EntityAttribute, error) {
	query := `
		SELECT id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at
		FROM entity_attributes
		WHERE entity_type = ? AND entity_id = ?
	`
	args := []interface{}{entityType, entityID}

	if attributeName != "" {
		query += " AND attribute_name = ?"
		args = append(args, attributeName)
	}

	query += " ORDER BY attribute_name, valid_from DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query entity attributes: %w", err)
	}
	defer rows.Close()

	var attributes []*models.EntityAttribute
	for rows.Next() {
		var attr models.EntityAttribute
		var validFromStr, createdAtStr string
		var validUntil sql.NullString

		err := rows.Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &validFromStr, &validUntil, &attr.Source, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", err)
		}

		// Parse timestamps
		attr.ValidFrom, _ = parseTimestamp(validFromStr)
		attr.CreatedAt, _ = parseTimestamp(createdAtStr)

		if validUntil.Valid {
			t, _ := parseTimestamp(validUntil.String)
			attr.ValidUntil = &t
		}

		attributes = append(attributes, &attr)
	}

	return attributes, rows.Err()
}

// GetEngineerCost retrieves current cost configuration for an engineer
// Replaces GetCurrentCostForEntity from cost_roi.go for engineers
func (s *AttributeStore) GetEngineerCost(engineerID string) (*models.EntityAttribute, error) {
	var attr models.EntityAttribute
	var validFromStr, createdAtStr string
	var validUntil sql.NullString

	now := time.Now().UTC()
	err := s.db.QueryRow(`
		SELECT id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at
		FROM entity_attributes
		WHERE entity_type = 'engineer'
		  AND entity_id = ?
		  AND attribute_name = 'monthly_cost'
		  AND valid_from <= ?
		  AND (valid_until IS NULL OR valid_until > ?)
		ORDER BY valid_from DESC
		LIMIT 1
	`, engineerID, now, now).Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &validFromStr, &validUntil, &attr.Source, &createdAtStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get engineer cost: %w", err)
	}

	// Parse timestamps
	attr.ValidFrom, _ = parseTimestamp(validFromStr)
	attr.CreatedAt, _ = parseTimestamp(createdAtStr)

	if validUntil.Valid {
		t, _ := parseTimestamp(validUntil.String)
		attr.ValidUntil = &t
	}

	return &attr, nil
}

// GetTeamCosts retrieves current cost configurations for all team members
func (s *AttributeStore) GetTeamCosts(teamID string) ([]*models.EntityAttribute, error) {
	// First get all engineers in the team
	// This would typically join with a team_members table or engineers table
	// For now, we'll query attributes directly

	now := time.Now().UTC()
	rows, err := s.db.Query(`
		SELECT id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at
		FROM entity_attributes
		WHERE entity_type IN ('engineer', 'team')
		  AND attribute_name = 'monthly_cost'
		  AND valid_from <= ?
		  AND (valid_until IS NULL OR valid_until > ?)
		ORDER BY entity_type, entity_id, valid_from DESC
	`, now, now)

	if err != nil {
		return nil, fmt.Errorf("failed to query team costs: %w", err)
	}
	defer rows.Close()

	var attributes []*models.EntityAttribute
	seen := make(map[string]bool) // Track entity_id to get only most recent

	for rows.Next() {
		var attr models.EntityAttribute
		var validUntil sql.NullTime

		err := rows.Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &attr.ValidFrom, &validUntil, &attr.Source, &attr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", err)
		}

		if validUntil.Valid {
			attr.ValidUntil = &validUntil.Time
		}

		// Only include the first (most recent) cost for each entity
		key := attr.EntityType + ":" + attr.EntityID
		if !seen[key] {
			seen[key] = true
			attributes = append(attributes, &attr)
		}
	}

	return attributes, rows.Err()
}

// GetCurrentAttributeValue retrieves the current value of a specific attribute
// Returns the value as a map if it's JSON, otherwise as a string
func (s *AttributeStore) GetCurrentAttributeValue(entityType, entityID, attributeName string) (interface{}, error) {
	var attr models.EntityAttribute
	var validFromStr, createdAtStr string
	var validUntil sql.NullString

	now := time.Now().UTC()
	err := s.db.QueryRow(`
		SELECT id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at
		FROM entity_attributes
		WHERE entity_type = ?
		  AND entity_id = ?
		  AND attribute_name = ?
		  AND valid_from <= ?
		  AND (valid_until IS NULL OR valid_until > ?)
		ORDER BY valid_from DESC
		LIMIT 1
	`, entityType, entityID, attributeName, now, now).Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &validFromStr, &validUntil, &attr.Source, &createdAtStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute value: %w", err)
	}

	// Parse timestamps (not used in return value but needed for scanning)
	attr.ValidFrom, _ = parseTimestamp(validFromStr)
	attr.CreatedAt, _ = parseTimestamp(createdAtStr)

	if validUntil.Valid {
		t, _ := parseTimestamp(validUntil.String)
		attr.ValidUntil = &t
	}

	// Parse value based on type
	switch attr.ValueType {
	case "json":
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(attr.Value), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON value: %w", err)
		}
		return result, nil
	default:
		return attr.Value, nil
	}
}
