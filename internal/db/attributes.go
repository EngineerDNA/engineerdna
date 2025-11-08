package db

import (
	"database/sql"
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
	var validUntil sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, entity_type, entity_id, attribute_name, value, value_type, valid_from, valid_until, source, created_at
		FROM entity_attributes WHERE id = ?
	`, id).Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &attr.ValidFrom, &validUntil, &attr.Source, &attr.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	if validUntil.Valid {
		attr.ValidUntil = &validUntil.Time
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
		var validUntil sql.NullTime

		err := rows.Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &attr.ValidFrom, &validUntil, &attr.Source, &attr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", err)
		}

		if validUntil.Valid {
			attr.ValidUntil = &validUntil.Time
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
		var validUntil sql.NullTime

		err := rows.Scan(&attr.ID, &attr.EntityType, &attr.EntityID, &attr.AttributeName, &attr.Value, &attr.ValueType, &attr.ValidFrom, &validUntil, &attr.Source, &attr.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attribute: %w", err)
		}

		if validUntil.Valid {
			attr.ValidUntil = &validUntil.Time
		}

		attributes = append(attributes, &attr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attributes: %w", err)
	}

	return attributes, nil
}
