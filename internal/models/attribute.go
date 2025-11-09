package models

import "time"

// EntityAttribute represents a fact about an entity (team, engineer, org)
// with temporal validity tracking
type EntityAttribute struct {
	ID            string     `json:"id"`
	EntityType    string     `json:"entity_type"` // team, engineer, org
	EntityID      string     `json:"entity_id"`
	AttributeName string     `json:"attribute_name"`
	Value         string     `json:"value"`
	ValueType     string     `json:"value_type"`            // string, number, boolean, json
	ValidFrom     time.Time  `json:"valid_from"`            // When this value became valid
	ValidUntil    *time.Time `json:"valid_until,omitempty"` // When it stopped being valid (NULL = current)
	Source        string     `json:"source"`                // Plugin name
	CreatedAt     time.Time  `json:"created_at"`
}
