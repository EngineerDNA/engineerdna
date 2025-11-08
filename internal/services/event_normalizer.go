package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/internal/plugin"
)

// EventNormalizer normalizes events using the event type registry
type EventNormalizer struct {
	registry *plugin.EventTypeRegistry
}

// NewEventNormalizer creates a new event normalizer
func NewEventNormalizer(registry *plugin.EventTypeRegistry) *EventNormalizer {
	return &EventNormalizer{
		registry: registry,
	}
}

// NormalizeEvent applies normalization to an event based on registry mappings
func (n *EventNormalizer) NormalizeEvent(event *models.Event) error {
	// Look up registration by event source + event type
	registration, ok := n.registry.GetRegistration(event.Source, event.Type)
	if !ok {
		// No normalization mapping found - this is OK, not all events need normalization
		return nil
	}

	// Set normalized type
	event.NormalizedType = registration.NormalizedType

	// Apply normalization mappings
	normalizedData := make(map[string]interface{})
	for normalizedField, sourcePath := range registration.NormalizationMap {
		// Extract value from event.Data using source path
		value, err := n.extractValue(event.Data, sourcePath)
		if err != nil {
			// Log warning but continue - missing fields are acceptable
			log.Printf("Warning: failed to extract field %s from %s event: %v",
				sourcePath, event.Type, err)
			continue
		}

		// Set in normalized data
		normalizedData[normalizedField] = value
	}

	// Marshal normalized data to JSON string
	if len(normalizedData) > 0 {
		normalizedJSON, err := json.Marshal(normalizedData)
		if err != nil {
			return fmt.Errorf("failed to marshal normalized data: %w", err)
		}
		event.NormalizedData = string(normalizedJSON)
	}

	return nil
}

// NormalizeBatch normalizes multiple events
func (n *EventNormalizer) NormalizeBatch(events []*models.Event) error {
	for _, event := range events {
		if err := n.NormalizeEvent(event); err != nil {
			return fmt.Errorf("failed to normalize event %s: %w", event.ID, err)
		}
	}
	return nil
}

// extractValue extracts a value from a map using a path like "author.name"
func (n *EventNormalizer) extractValue(data map[string]interface{}, path string) (interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}

	// Split path by dots (e.g., "author.name" -> ["author", "name"])
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	// Traverse the path
	current := interface{}(data)
	for i, part := range parts {
		// Current must be a map
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot traverse path at %s: not a map", strings.Join(parts[:i+1], "."))
		}

		// Get the value at this level
		value, ok := currentMap[part]
		if !ok {
			return nil, fmt.Errorf("field not found: %s", strings.Join(parts[:i+1], "."))
		}

		current = value
	}

	return current, nil
}
