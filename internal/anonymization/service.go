package anonymization

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// Service handles anonymization operations
type Service struct {
	store *db.AnonymizationStore
}

// NewService creates a new anonymization service
func NewService(store *db.AnonymizationStore) *Service {
	return &Service{store: store}
}

// Anonymize converts a real identifier to an anonymized ID
func (s *Service) Anonymize(realIdentifier string, strategy models.AnonymizationStrategy) (string, error) {
	if realIdentifier == "" {
		return "", nil
	}

	// Check existing mapping
	mapping, err := s.store.GetMappingByReal(realIdentifier)
	if err != nil {
		return "", fmt.Errorf("failed to check mapping: %w", err)
	}

	if mapping != nil && mapping.Enabled {
		return mapping.AnonymizedID, nil
	}

	// Create mapping
	newMapping := &models.AnonymizationMapping{
		ID:        uuid.New().String(),
		RealName:  realIdentifier,
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
	}

	// For sequential strategy, use atomic method with retry on conflict
	if strategy == models.StrategySequential {
		// Retry up to 3 times if we hit a unique constraint violation
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			if err := s.store.CreateMappingAtomic(newMapping, strategy); err != nil {
				// Check if it's a unique constraint error (duplicate anonymized_id)
				if i < maxRetries-1 && isUniqueConstraintError(err) {
					// Retry - another concurrent transaction created a mapping
					continue
				}
				return "", fmt.Errorf("failed to create mapping: %w", err)
			}
			return newMapping.AnonymizedID, nil
		}
		return "", fmt.Errorf("failed to create mapping after %d retries", maxRetries)
	}

	// For other strategies, generate ID first then create
	newMapping.AnonymizedID = s.generateAnonymizedID(realIdentifier, strategy)
	if err := s.store.CreateMapping(newMapping); err != nil {
		return "", fmt.Errorf("failed to create mapping: %w", err)
	}

	return newMapping.AnonymizedID, nil
}

// Deanonymize converts an anonymized ID back to the real identifier
func (s *Service) Deanonymize(anonID string) (string, error) {
	if anonID == "" {
		return "", nil
	}

	mapping, err := s.store.GetMappingByAnonymized(anonID)
	if err != nil {
		return "", fmt.Errorf("failed to get mapping: %w", err)
	}

	if mapping == nil {
		return anonID, nil // Return as-is if no mapping found
	}

	return mapping.RealName, nil
}

func (s *Service) generateAnonymizedID(realIdentifier string, strategy models.AnonymizationStrategy) string {
	switch strategy {
	case models.StrategySequential:
		count, err := s.store.CountMappings()
		if err != nil {
			// If count fails, fall back to timestamp-based ID to avoid duplicates
			return fmt.Sprintf("Engineer_%d", time.Now().UTC().Unix())
		}
		if count < 26 {
			return fmt.Sprintf("Engineer_%c", 'A'+count)
		}
		return fmt.Sprintf("Engineer_%d", count+1)

	case models.StrategyHash:
		hash := sha256.Sum256([]byte(realIdentifier))
		return "eng_" + hex.EncodeToString(hash[:8])

	case models.StrategyUUID:
		return uuid.New().String()

	default:
		return fmt.Sprintf("Engineer_%d", time.Now().UTC().Unix())
	}
}

// AnonymizeEvent anonymizes specified fields in an event
func (s *Service) AnonymizeEvent(event *models.Event, fields []string, strategy models.AnonymizationStrategy) error {
	if event.Actor != "" {
		anonActor, err := s.Anonymize(event.Actor, strategy)
		if err != nil {
			return fmt.Errorf("failed to anonymize actor: %w", err)
		}
		event.Actor = anonActor
	}

	for _, field := range fields {
		if val, ok := event.Data[field].(string); ok && val != "" {
			anonVal, err := s.Anonymize(val, strategy)
			if err != nil {
				return fmt.Errorf("failed to anonymize field %s: %w", field, err)
			}
			event.Data[field] = anonVal
		}
	}

	event.Anonymized = true
	return nil
}

// DeanonymizeEvent deanonymizes specified fields in an event
func (s *Service) DeanonymizeEvent(event *models.Event, fields []string) error {
	if event.Actor != "" {
		realActor, err := s.Deanonymize(event.Actor)
		if err == nil {
			event.Actor = realActor
		}
	}

	for _, field := range fields {
		if val, ok := event.Data[field].(string); ok && val != "" {
			realVal, err := s.Deanonymize(val)
			if err == nil {
				event.Data[field] = realVal
			}
		}
	}

	event.Anonymized = false
	return nil
}

// AnonymizeRepoName creates an anonymized repository name
func (s *Service) AnonymizeRepoName(repoName string) string {
	hash := sha256.Sum256([]byte(repoName))
	return "Project_" + hex.EncodeToString(hash[:4])
}

// SanitizeCommitMessage removes PII from commit messages
func (s *Service) SanitizeCommitMessage(message string) string {
	// Remove email addresses
	emailRegex := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	sanitized := emailRegex.ReplaceAllString(message, "[email]")

	// Remove names that look like "Alice Johnson"
	nameRegex := regexp.MustCompile(`\b[A-Z][a-z]+ [A-Z][a-z]+\b`)
	sanitized = nameRegex.ReplaceAllString(sanitized, "[name]")

	return sanitized
}

// AnonymizeTimestamp rounds timestamp to reduce precision
func (s *Service) AnonymizeTimestamp(t time.Time, granularity time.Duration) time.Time {
	return t.Truncate(granularity)
}

// CloneAndAnonymizeEvents creates anonymized copies of events
func (s *Service) CloneAndAnonymizeEvents(events []*models.Event, fields []string, strategy models.AnonymizationStrategy) ([]*models.Event, error) {
	anonEvents := make([]*models.Event, len(events))

	for i, event := range events {
		// Deep copy
		dataCopy := make(map[string]interface{})
		for k, v := range event.Data {
			dataCopy[k] = v
		}

		anonEvent := &models.Event{
			ID:         event.ID,
			Type:       event.Type,
			Source:     event.Source,
			SourceID:   event.SourceID,
			Timestamp:  event.Timestamp,
			Actor:      event.Actor,
			Data:       dataCopy,
			Anonymized: event.Anonymized,
			CreatedAt:  event.CreatedAt,
			UpdatedAt:  event.UpdatedAt,
		}

		// Anonymize
		if err := s.AnonymizeEvent(anonEvent, fields, strategy); err != nil {
			return nil, fmt.Errorf("failed to anonymize event %s: %w", event.ID, err)
		}

		// Anonymize repo names if present
		if repo, ok := anonEvent.Data["repo"].(string); ok {
			anonEvent.Data["repo"] = s.AnonymizeRepoName(repo)
		}

		// Sanitize commit messages
		if msg, ok := anonEvent.Data["message"].(string); ok {
			anonEvent.Data["message"] = s.SanitizeCommitMessage(msg)
		}

		// Round timestamps
		anonEvent.Timestamp = s.AnonymizeTimestamp(anonEvent.Timestamp, time.Hour)

		anonEvents[i] = anonEvent
	}

	return anonEvents, nil
}

// isUniqueConstraintError checks if an error is a SQLite unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite unique constraint errors contain "UNIQUE constraint"
	return regexp.MustCompile(`UNIQUE constraint`).MatchString(err.Error())
}
