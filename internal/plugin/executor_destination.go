package plugin

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// ExportToDestination exports data to a destination plugin
func (e *Executor) ExportToDestination(pluginName string, params sdk.ExportParams, anonymize bool) (*sdk.ExportResult, error) {
	// Load plugin config
	pluginConfig, err := e.pluginStore.GetByName(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin config: %w", err)
	}
	if pluginConfig == nil {
		return nil, fmt.Errorf("plugin not configured: %s", pluginName)
	}

	// Get plugin info
	info, err := e.GetPluginInfo(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin info: %w", err)
	}

	// Get anonymization policy
	policy, err := e.anonStore.GetPolicy(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to get anonymization policy: %w", err)
	}

	// Determine if anonymization is needed
	needsAnonymization := anonymize || info.Anonymization.Required || (policy != nil && policy.Enabled)

	// Anonymize export data if needed
	if needsAnonymization {
		// When anonymization is required, validate that data structure is correct
		eventsData, ok := params.Data["events"]
		if !ok {
			return nil, fmt.Errorf("anonymization required but 'events' field missing from export data")
		}

		// Convert to models.Event slice
		eventsList, ok := eventsData.([]interface{})
		if !ok {
			return nil, fmt.Errorf("anonymization required but 'events' field has invalid type (expected array, got %T)", eventsData)
		}

		{
			// Enforce max event count to prevent unbounded resource usage
			if len(eventsList) > MaxPluginEvents {
				return nil, fmt.Errorf("too many events to export: %d (max %d)", len(eventsList), MaxPluginEvents)
			}

			var modelEvents []*models.Event
			for _, e := range eventsList {
				if eventMap, ok := e.(map[string]interface{}); ok {
					// Validate required fields before conversion
					eventType := getString(eventMap, "type")
					eventTimestamp := getTime(eventMap, "timestamp")
					eventSource := getString(eventMap, "source")
					eventActor := getString(eventMap, "actor")

					if eventType == "" {
						return nil, fmt.Errorf("invalid event: missing required field 'type'")
					}
					if eventTimestamp.IsZero() {
						return nil, fmt.Errorf("invalid event: missing or invalid required field 'timestamp'")
					}
					if eventSource == "" {
						return nil, fmt.Errorf("invalid event: missing required field 'source'")
					}
					if eventActor == "" {
						return nil, fmt.Errorf("invalid event: missing required field 'actor'")
					}

					// Validate data field structure
					eventData := getMap(eventMap, "data")
					if eventData == nil {
						return nil, fmt.Errorf("invalid event: missing or invalid required field 'data'")
					}

					// Convert to model event (preserve all fields)
					modelEvent := &models.Event{
						ID:        getString(eventMap, "id"),
						Type:      eventType,
						Source:    eventSource,
						SourceID:  getString(eventMap, "source_id"),
						Timestamp: eventTimestamp,
						Actor:     eventActor,
						Data:      eventData,
					}
					modelEvents = append(modelEvents, modelEvent)
				}
			}

			// Get strategy
			strategy := models.StrategySequential
			if policy != nil {
				strategy = policy.Strategy
			} else if info.Anonymization.Strategy != "" {
				strategy = models.AnonymizationStrategy(info.Anonymization.Strategy)
			}

			// Anonymize events
			anonEvents, err := e.anonService.CloneAndAnonymizeEvents(modelEvents, info.Anonymization.Fields, strategy)
			if err != nil {
				return nil, fmt.Errorf("failed to anonymize events: %w", err)
			}

			// Convert back to interface{} for export (preserve all fields)
			var anonEventsList []interface{}
			for _, event := range anonEvents {
				anonEventsList = append(anonEventsList, map[string]interface{}{
					"id":        event.ID,
					"type":      event.Type,
					"source":    event.Source,
					"source_id": event.SourceID,
					"timestamp": event.Timestamp.Format(time.RFC3339),
					"actor":     event.Actor,
					"data":      event.Data,
				})
			}
			params.Data["events"] = anonEventsList
		}
	}

	// Decrypt config
	var secretFields []string
	for _, field := range info.ConfigFields {
		if field.Secret {
			secretFields = append(secretFields, field.Name)
		}
	}

	decryptedConfig, err := e.keyStore.DecryptPluginConfig(pluginConfig.Config, secretFields)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt config: %w", err)
	}

	// Start plugin
	pluginPath, err := e.loader.GetPluginPath(pluginName)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}

	client, err := NewClient(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}
	defer client.Close()

	// Configure plugin
	if err := client.Configure(decryptedConfig); err != nil {
		return nil, fmt.Errorf("failed to configure plugin: %w", err)
	}

	// Export data
	result, err := client.Export(params)
	if err != nil {
		return nil, fmt.Errorf("export failed: %w", err)
	}

	// Log audit entry (MUST succeed for compliance)
	if e.auditStore != nil {
		destination := "unknown"
		if result != nil && result.URL != "" {
			destination = result.URL
		}

		auditEntry := &models.AuditLogEntry{
			PluginName:    pluginName,
			Action:        "destination.export",
			EventCount:    0, // Exports don't have event counts
			Anonymized:    needsAnonymization,
			Destination:   destination,
			UserInitiated: true,
			DataSummary: map[string]interface{}{
				"data_type": params.DataType,
			},
		}
		if err := e.auditStore.LogEntry(auditEntry); err != nil {
			return nil, fmt.Errorf("failed to log audit entry (export aborted for compliance): %w", err)
		}
	}

	return result, nil
}
