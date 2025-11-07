package models

import "time"

type PluginType string

const (
	PluginTypeSource      PluginType = "source"
	PluginTypeDestination PluginType = "destination"
	PluginTypeProcessor   PluginType = "processor"
)

type PluginConfig struct {
	Name      string            `json:"name"`
	Type      PluginType        `json:"type"`
	Enabled   bool              `json:"enabled"`
	Config    map[string]string `json:"config"`
	LastSync  *time.Time        `json:"last_sync"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type PluginMetadata struct {
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Type          PluginType        `json:"type"`
	Description   string            `json:"description"`
	Author        string            `json:"author"`
	ConfigFields  []ConfigField     `json:"config_fields"`
	Anonymization AnonymizationSpec `json:"anonymization"`
	OAuth         *OAuthSpec        `json:"oauth,omitempty"`
	Schedule      *ScheduleSpec     `json:"schedule,omitempty"`
}

type ConfigField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Default     string   `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
	Secret      bool     `json:"secret"`
}

type AnonymizationSpec struct {
	Required bool     `json:"required"`
	Reason   string   `json:"reason"`
	Strategy string   `json:"strategy"`
	Fields   []string `json:"fields"`
}

type OAuthSpec struct {
	Provider string   `json:"provider"`
	Scopes   []string `json:"scopes"`
}

type ScheduleSpec struct {
	Supported bool   `json:"supported"`
	Default   string `json:"default"`
}
