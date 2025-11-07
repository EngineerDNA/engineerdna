package plugin

import (
	"fmt"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

// AnalyzeWithProcessor analyzes events with a processor plugin
func (e *Executor) AnalyzeWithProcessor(pluginName string, events []*models.Event, analysisType string, context map[string]interface{}) (*sdk.AnalyzeResult, error) {
	// Enforce max event count to prevent unbounded resource usage
	const maxEvents = 50000
	if len(events) > maxEvents {
		return nil, fmt.Errorf("too many events to analyze: %d (max %d)", len(events), maxEvents)
	}

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
	needsAnonymization := info.Anonymization.Required || (policy != nil && policy.Enabled)

	var sdkEvents []sdk.Event
	if needsAnonymization {
		// Get strategy
		strategy := models.StrategySequential
		if policy != nil {
			strategy = policy.Strategy
		} else if info.Anonymization.Strategy != "" {
			strategy = models.AnonymizationStrategy(info.Anonymization.Strategy)
		}

		// Clone and anonymize events
		anonEvents, err := e.anonService.CloneAndAnonymizeEvents(events, info.Anonymization.Fields, strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to anonymize events: %w", err)
		}

		// Convert to SDK events
		sdkEvents = make([]sdk.Event, len(anonEvents))
		for i, event := range anonEvents {
			sdkEvents[i] = sdk.Event{
				ID:        event.ID,
				Type:      event.Type,
				Source:    event.Source,
				SourceID:  event.SourceID,
				Timestamp: event.Timestamp,
				Actor:     event.Actor,
				Data:      event.Data,
			}
		}
	} else {
		// Use original events
		sdkEvents = make([]sdk.Event, len(events))
		for i, event := range events {
			sdkEvents[i] = sdk.Event{
				ID:        event.ID,
				Type:      event.Type,
				Source:    event.Source,
				SourceID:  event.SourceID,
				Timestamp: event.Timestamp,
				Actor:     event.Actor,
				Data:      event.Data,
			}
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

	// Analyze
	result, err := client.Analyze(sdk.AnalyzeRequest{
		AnalysisType: analysisType,
		Events:       sdkEvents,
		Context:      context,
	})
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Log audit entry (MUST succeed for compliance)
	if e.auditStore != nil {
		auditEntry := &models.AuditLogEntry{
			PluginName:    pluginName,
			Action:        "processor.analyze",
			EventCount:    len(events),
			Anonymized:    needsAnonymization,
			Destination:   "AI API",
			UserInitiated: true,
			DataSummary: map[string]interface{}{
				"analysis_type": analysisType,
			},
		}
		if err := e.auditStore.LogEntry(auditEntry); err != nil {
			return nil, fmt.Errorf("failed to log audit entry (analysis aborted for compliance): %w", err)
		}
	}

	return result, nil
}
