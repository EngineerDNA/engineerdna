package context

import (
	"strings"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

type Service struct {
	contextStore  *db.ContextStore
	identityStore *db.IdentityStore
}

func NewService(contextStore *db.ContextStore, identityStore *db.IdentityStore) *Service {
	return &Service{
		contextStore:  contextStore,
		identityStore: identityStore,
	}
}

// GetEngineerContext returns complete context profile for an engineer
func (s *Service) GetEngineerContext(engineerID string) (*models.EngineerContext, error) {
	// Get engineer details
	engineer, err := s.identityStore.GetEngineerByID(engineerID)
	if err != nil {
		return nil, err
	}
	if engineer == nil {
		return nil, nil
	}

	// Get all notes for this engineer (both as subject and as manager)
	notes, _, err := s.contextStore.ListManagerNotes("", "engineer", engineerID, []string{"private", "shared_with_subject", "team", "org"}, 100, 0)
	if err != nil {
		return nil, err
	}

	// Get annotations
	annotations, _, err := s.contextStore.GetContextAnnotations("engineer", engineerID, 100, 0)
	if err != nil {
		return nil, err
	}

	// Extract action items from notes
	actionItems := s.extractActionItems(notes)

	return &models.EngineerContext{
		EngineerID:   engineerID,
		EngineerName: engineer.Name,
		Notes:        notesToList(notes),
		Annotations:  annotationsToList(annotations),
		ActionItems:  actionItems,
	}, nil
}

// GetMetricContext returns context explaining metric changes
func (s *Service) GetMetricContext(metricType, metricID string) ([]*models.ContextAnnotation, error) {
	annotations, _, err := s.contextStore.GetContextAnnotations(metricType, metricID, 100, 0)
	return annotations, err
}

// GetEventContext returns context for a specific event
func (s *Service) GetEventContext(eventID string) ([]*models.ContextAnnotation, error) {
	annotations, _, err := s.contextStore.GetContextAnnotations("event", eventID, 100, 0)
	return annotations, err
}

// LinkContextToMetric associates an explanation with a metric
func (s *Service) LinkContextToMetric(metricType, metricID, content, authorID string) error {
	annotation := &models.ContextAnnotation{
		EntityType:     metricType,
		EntityID:       metricID,
		AnnotationType: "explanation",
		Content:        content,
		AuthorID:       authorID,
		Visibility:     "team",
	}
	return s.contextStore.CreateContextAnnotation(annotation)
}

// SearchNotes searches manager notes by content, tags, or subject
func (s *Service) SearchNotes(managerID, searchTerm string) ([]*models.ManagerNote, error) {
	// Get all notes for the manager
	notes, _, err := s.contextStore.ListManagerNotes(managerID, "", "", []string{"private", "shared_with_subject", "team", "org"}, 1000, 0)
	if err != nil {
		return nil, err
	}

	// Filter by search term
	var filtered []*models.ManagerNote
	searchLower := strings.ToLower(searchTerm)
	for _, note := range notes {
		if strings.Contains(strings.ToLower(note.Title), searchLower) ||
			strings.Contains(strings.ToLower(note.Content), searchLower) ||
			containsTag(note.Tags, searchLower) {
			filtered = append(filtered, note)
		}
	}

	return filtered, nil
}

// GetActionItems extracts all action items from notes
func (s *Service) GetActionItems(managerID string) ([]string, error) {
	notes, _, err := s.contextStore.ListManagerNotes(managerID, "", "", []string{"private", "shared_with_subject", "team", "org"}, 1000, 0)
	if err != nil {
		return nil, err
	}

	return s.extractActionItems(notes), nil
}

// Helper functions

func (s *Service) extractActionItems(notes []*models.ManagerNote) []string {
	var items []string
	for _, note := range notes {
		items = append(items, note.ActionItems...)
	}
	return items
}

func notesToList(notes []*models.ManagerNote) []models.ManagerNote {
	if notes == nil {
		return []models.ManagerNote{}
	}
	result := make([]models.ManagerNote, len(notes))
	for i, note := range notes {
		result[i] = *note
	}
	return result
}

func annotationsToList(annotations []*models.ContextAnnotation) []models.ContextAnnotation {
	if annotations == nil {
		return []models.ContextAnnotation{}
	}
	result := make([]models.ContextAnnotation, len(annotations))
	for i, ann := range annotations {
		result[i] = *ann
	}
	return result
}

func containsTag(tags []string, search string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), search) {
			return true
		}
	}
	return false
}
