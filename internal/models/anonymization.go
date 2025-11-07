package models

import "time"

type AnonymizationStrategy string

const (
	StrategySequential AnonymizationStrategy = "sequential"
	StrategyUUID       AnonymizationStrategy = "uuid"
	StrategyHash       AnonymizationStrategy = "hash"
)

type AnonymizationMapping struct {
	ID                  string            `json:"id"`
	AnonymizedID        string            `json:"anonymized_id"`
	RealName            string            `json:"real_name"`
	RealEmail           string            `json:"real_email"`
	ExternalIdentifiers map[string]string `json:"external_identifiers"`
	Enabled             bool              `json:"enabled"`
	CreatedAt           time.Time         `json:"created_at"`
}

type AnonymizationPolicy struct {
	PluginName string                `json:"plugin_name"`
	Enabled    bool                  `json:"enabled"`
	Strategy   AnonymizationStrategy `json:"strategy"`
	Fields     []string              `json:"fields"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}
