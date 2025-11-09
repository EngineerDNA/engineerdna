package identity

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Service handles identity resolution and management
type Service struct {
	store  *db.IdentityStore
	logger *slog.Logger
}

// NewService creates a new identity service
func NewService(store *db.IdentityStore) *Service {
	return &Service{
		store:  store,
		logger: slog.Default(),
	}
}

// ResolveIdentity attempts to resolve a source+identifier to an engineer ID.
// If resolved, returns engineer_id. If not, creates/updates unresolved entry and returns empty string.
func (s *Service) ResolveIdentity(source, identifier string) (string, error) {
	if source == "" || identifier == "" {
		return "", fmt.Errorf("source and identifier are required")
	}

	// Try to find existing engineer with this identifier
	engineer, err := s.store.GetEngineerByIdentifier(source, identifier)
	if err != nil {
		return "", fmt.Errorf("failed to lookup engineer: %w", err)
	}

	if engineer != nil {
		return engineer.ID, nil
	}

	// Not resolved - track as unresolved
	unresolved, err := s.store.GetUnresolvedBySourceIdentifier(source, identifier)
	if err != nil {
		return "", fmt.Errorf("failed to lookup unresolved identity: %w", err)
	}

	if unresolved != nil {
		// Already exists, increment count
		if err := s.store.IncrementUnresolvedEventCount(source, identifier); err != nil {
			return "", fmt.Errorf("failed to increment unresolved count: %w", err)
		}
	} else {
		// Create new unresolved identity
		unresolved = &models.UnresolvedIdentity{
			Source:     source,
			Identifier: identifier,
			FirstSeen:  time.Now().UTC(),
			EventCount: 1,
		}
		if err := s.store.CreateUnresolvedIdentity(unresolved); err != nil {
			return "", fmt.Errorf("failed to create unresolved identity: %w", err)
		}
	}

	return "", nil
}

// AssignIdentity assigns an unresolved identity to an engineer and updates all related events
func (s *Service) AssignIdentity(engineerID, source, identifier string) error {
	// Verify engineer exists
	engineer, err := s.store.GetEngineerByID(engineerID)
	if err != nil {
		return fmt.Errorf("failed to get engineer: %w", err)
	}
	if engineer == nil {
		return fmt.Errorf("engineer not found: %s", engineerID)
	}

	// Add identifier to engineer if not already present
	if engineer.Identifiers == nil {
		engineer.Identifiers = make(map[string]string)
	}
	engineer.Identifiers[source] = identifier

	// Perform all operations in a transaction
	if err := s.store.AssignIdentityTx(engineer, source, identifier); err != nil {
		return fmt.Errorf("failed to assign identity: %w", err)
	}

	s.logger.Info("assigned identity to engineer",
		"source", source,
		"identifier", identifier,
		"engineer_id", engineerID,
		"engineer_name", engineer.Name)

	return nil
}

// validateEngineerInput validates engineer input fields
func validateEngineerInput(name, email, manager string, identifiers map[string]string) error {
	// Name validation
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > MaxNameLength {
		return fmt.Errorf("name exceeds maximum length of %d characters", MaxNameLength)
	}

	// Email validation
	if email != "" {
		if len(email) > MaxEmailLength {
			return fmt.Errorf("email exceeds maximum length of %d characters", MaxEmailLength)
		}
		// Basic email format validation
		if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
			return fmt.Errorf("invalid email format")
		}
	}

	// Manager validation
	if len(manager) > MaxManagerLength {
		return fmt.Errorf("manager exceeds maximum length of %d characters", MaxManagerLength)
	}

	// Identifiers validation
	if len(identifiers) > MaxIdentifiers {
		return fmt.Errorf("identifiers map exceeds maximum size of %d", MaxIdentifiers)
	}

	return nil
}

// CreateEngineer creates a new engineer with optional initial identifiers
func (s *Service) CreateEngineer(name, email, manager string, identifiers map[string]string) (string, error) {
	// Validate input
	if err := validateEngineerInput(name, email, manager, identifiers); err != nil {
		return "", err
	}

	if identifiers == nil {
		identifiers = make(map[string]string)
	}

	engineer := &models.Engineer{
		Name:        name,
		Email:       email,
		Manager:     manager,
		Identifiers: identifiers,
		Active:      true,
	}

	// CreateEngineer now checks for duplicate identifiers atomically within transaction
	if err := s.store.CreateEngineer(engineer); err != nil {
		return "", fmt.Errorf("failed to create engineer: %w", err)
	}

	return engineer.ID, nil
}

// CreateEngineerFromUnresolved creates a new engineer from an unresolved identity in a transaction
func (s *Service) CreateEngineerFromUnresolved(name, email, manager, source, identifier string) (string, error) {
	if source == "" || identifier == "" {
		return "", fmt.Errorf("source and identifier are required")
	}

	identifiers := map[string]string{
		source: identifier,
	}

	// Validate input
	if err := validateEngineerInput(name, email, manager, identifiers); err != nil {
		return "", err
	}

	engineer := &models.Engineer{
		Name:        name,
		Email:       email,
		Manager:     manager,
		Identifiers: identifiers,
		Active:      true,
	}

	// CreateEngineerFromUnresolvedTx now checks for duplicate identifiers atomically within transaction
	if err := s.store.CreateEngineerFromUnresolvedTx(engineer, source, identifier); err != nil {
		return "", fmt.Errorf("failed to create engineer from unresolved: %w", err)
	}

	s.logger.Info("created engineer from unresolved identity",
		"engineer_id", engineer.ID,
		"engineer_name", engineer.Name,
		"source", source,
		"identifier", identifier)

	return engineer.ID, nil
}

// UpdateEngineer updates an existing engineer
func (s *Service) UpdateEngineer(engineerID string, name, email, manager string, identifiers map[string]string, active bool) error {
	engineer, err := s.store.GetEngineerByID(engineerID)
	if err != nil {
		return fmt.Errorf("failed to get engineer: %w", err)
	}
	if engineer == nil {
		return fmt.Errorf("engineer not found: %s", engineerID)
	}

	// Validate input
	if err := validateEngineerInput(name, email, manager, identifiers); err != nil {
		return err
	}

	if name != "" {
		engineer.Name = name
	}
	engineer.Email = email
	engineer.Manager = manager
	if identifiers != nil {
		engineer.Identifiers = identifiers
	}
	engineer.Active = active

	// Use transactional update with atomic duplicate check
	if err := s.store.UpdateEngineerTx(engineer); err != nil {
		return fmt.Errorf("failed to update engineer: %w", err)
	}

	return nil
}

// DeleteEngineer soft-deletes an engineer
func (s *Service) DeleteEngineer(engineerID string) error {
	return s.store.DeleteEngineer(engineerID)
}

// ListEngineers retrieves all engineers with pagination
func (s *Service) ListEngineers(activeOnly bool, limit, offset int) ([]*models.Engineer, error) {
	if limit <= 0 {
		limit = db.DefaultQueryLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.store.ListEngineers(activeOnly, limit, offset)
}

// CountEngineers returns the total count of engineers with optional filters
func (s *Service) CountEngineers(activeOnly bool) (int, error) {
	return s.store.CountEngineers(activeOnly)
}

// GetEngineer retrieves a single engineer by ID
func (s *Service) GetEngineer(engineerID string) (*models.Engineer, error) {
	return s.store.GetEngineerByID(engineerID)
}

// GetUnresolved retrieves all unresolved identities
func (s *Service) GetUnresolved() ([]*models.UnresolvedIdentity, error) {
	return s.store.ListUnresolvedIdentities(db.DefaultQueryLimit)
}

// GetIgnored retrieves all ignored unresolved identities
func (s *Service) GetIgnored() ([]*models.UnresolvedIdentity, error) {
	return s.store.ListIgnoredIdentities(db.DefaultQueryLimit)
}

// GetUnresolvedByID retrieves a single unresolved identity by ID
func (s *Service) GetUnresolvedByID(id string) (*models.UnresolvedIdentity, error) {
	return s.store.GetUnresolvedByID(id)
}

// DeleteUnresolvedIdentity removes an unresolved identity
func (s *Service) DeleteUnresolvedIdentity(source, identifier string) error {
	return s.store.DeleteUnresolvedIdentity(source, identifier)
}

// UpdateEventEngineerID updates all events matching source+identifier to reference the given engineer
func (s *Service) UpdateEventEngineerID(engineerID, source, identifier string) (int64, error) {
	return s.store.UpdateEventEngineerID(engineerID, source, identifier)
}

// SetUnresolvedIgnored sets the ignored flag for an unresolved identity
func (s *Service) SetUnresolvedIgnored(unresolvedID string, ignored bool) error {
	return s.store.SetUnresolvedIgnored(unresolvedID, ignored)
}

// SuggestMatches generates match suggestions between unresolved identities and engineers
func (s *Service) SuggestMatches() ([]*models.MatchSuggestion, error) {
	unresolved, err := s.store.ListUnresolvedIdentities(db.DefaultQueryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get unresolved identities: %w", err)
	}

	engineers, err := s.store.ListEngineers(true, db.DefaultQueryLimit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineers: %w", err)
	}

	// Use map to aggregate confidence per (unresolved, engineer) pair
	suggestionMap := make(map[string]*models.MatchSuggestion)

	for _, u := range unresolved {
		for _, eng := range engineers {
			confidence, reason := s.calculateMatchConfidence(u, eng)
			if confidence >= MinMatchConfidence {
				key := u.ID + ":" + eng.ID

				// Keep highest confidence for this pair
				if existing, ok := suggestionMap[key]; ok {
					if confidence > existing.Confidence {
						suggestionMap[key] = &models.MatchSuggestion{
							UnresolvedID: u.ID,
							EngineerID:   eng.ID,
							EngineerName: eng.Name,
							Confidence:   confidence,
							Reason:       reason,
						}
					}
				} else {
					suggestionMap[key] = &models.MatchSuggestion{
						UnresolvedID: u.ID,
						EngineerID:   eng.ID,
						EngineerName: eng.Name,
						Confidence:   confidence,
						Reason:       reason,
					}
				}
			}
		}
	}

	// Convert map to slice
	var suggestions []*models.MatchSuggestion
	for _, suggestion := range suggestionMap {
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// calculateMatchConfidence computes a confidence score for matching an unresolved identity to an engineer
func (s *Service) calculateMatchConfidence(unresolved *models.UnresolvedIdentity, engineer *models.Engineer) (float64, string) {
	confidence := 0.0
	reasons := []string{}

	identifier := unresolved.Identifier

	// Check if engineer has identifiers that could be used for matching
	for _, existingIdentifier := range engineer.Identifiers {
		// Email domain match
		if strings.Contains(identifier, "@") && strings.Contains(existingIdentifier, "@") {
			identifierDomain := extractDomain(identifier)
			existingDomain := extractDomain(existingIdentifier)
			if identifierDomain == existingDomain && identifierDomain != "" {
				confidence += EmailDomainMatchWeight
				reasons = append(reasons, fmt.Sprintf("Email domain match: %s", identifierDomain))
			}

			// Email username similarity
			identifierUser := extractEmailUsername(identifier)
			existingUser := extractEmailUsername(existingIdentifier)
			if levenshteinDistance(identifierUser, existingUser) <= EmailUsernameSimilarityThreshold {
				confidence += EmailUsernameMatchWeight
				reasons = append(reasons, fmt.Sprintf("Email username similarity: %s vs %s", identifierUser, existingUser))
			}
		}
	}

	// Name similarity (check identifier against engineer name) - only once
	nameDist := levenshteinDistance(strings.ToLower(identifier), strings.ToLower(engineer.Name))
	if nameDist <= LevenshteinThreshold {
		confidence += NameSimilarityWeight
		reasons = append(reasons, fmt.Sprintf("Name similarity: '%s' vs '%s'", identifier, engineer.Name))
	}

	// Username pattern matching (e.g., "asmith" from "Alice Smith") - only once
	if isUsernamePattern(identifier, engineer.Name) {
		confidence += UsernamePatternWeight
		reasons = append(reasons, fmt.Sprintf("Username pattern: '%s' matches '%s'", identifier, engineer.Name))
	}

	// Check engineer email if available
	if engineer.Email != "" {
		if strings.Contains(identifier, "@") {
			identifierDomain := extractDomain(identifier)
			emailDomain := extractDomain(engineer.Email)
			if identifierDomain == emailDomain && identifierDomain != "" {
				confidence += EmailDomainMatchWeight
				reasons = append(reasons, fmt.Sprintf("Email domain match: %s", identifierDomain))
			}

			identifierUser := extractEmailUsername(identifier)
			emailUser := extractEmailUsername(engineer.Email)
			if levenshteinDistance(identifierUser, emailUser) <= EmailUsernameSimilarityThreshold {
				confidence += EmailUsernameMatchWeight
				reasons = append(reasons, fmt.Sprintf("Email username similarity: %s vs %s", identifierUser, emailUser))
			}
		}
	}

	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	reason := strings.Join(reasons, "; ")
	return confidence, reason
}

// GetEngineerActivity retrieves activity metrics for an engineer over a time period
func (s *Service) GetEngineerActivity(engineerID string, days int) (*models.ActivityMetrics, error) {
	periodEnd := time.Now().UTC()
	periodStart := periodEnd.AddDate(0, 0, -days)

	return s.store.GetEngineerActivityMetrics(engineerID, periodStart, periodEnd)
}

// MergeEngineers merges two engineer records (combines identifiers, keeps the first one)
func (s *Service) MergeEngineers(keepID, mergeID string) error {
	keepEngineer, err := s.store.GetEngineerByID(keepID)
	if err != nil {
		return fmt.Errorf("failed to get keep engineer: %w", err)
	}
	if keepEngineer == nil {
		return fmt.Errorf("keep engineer not found: %s", keepID)
	}

	mergeEngineer, err := s.store.GetEngineerByID(mergeID)
	if err != nil {
		return fmt.Errorf("failed to get merge engineer: %w", err)
	}
	if mergeEngineer == nil {
		return fmt.Errorf("merge engineer not found: %s", mergeID)
	}

	// Merge identifiers
	if keepEngineer.Identifiers == nil {
		keepEngineer.Identifiers = make(map[string]string)
	}
	for source, identifier := range mergeEngineer.Identifiers {
		keepEngineer.Identifiers[source] = identifier
	}

	// Perform merge in transaction
	if err := s.store.MergeEngineersTx(keepEngineer, mergeEngineer); err != nil {
		return fmt.Errorf("failed to merge engineers: %w", err)
	}

	return nil
}

// Helper functions

func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func extractEmailUsername(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) >= 1 {
		return parts[0]
	}
	return email
}

// levenshteinDistance computes the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	// Prevent memory exhaustion DoS by limiting string length
	if len(s1) > MaxValidationLength || len(s2) > MaxValidationLength {
		return 999 // Return large distance to indicate no match
	}

	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// isUsernamePattern checks if identifier matches a common username pattern derived from name
// e.g., "asmith" matches "Alice Smith" (first initial + last name)
func isUsernamePattern(identifier, name string) bool {
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	name = strings.ToLower(strings.TrimSpace(name))

	parts := strings.Fields(name)
	if len(parts) < 2 {
		return false
	}

	firstName := parts[0]
	lastName := parts[len(parts)-1]

	// Pattern: first initial + last name (e.g., "asmith")
	pattern1 := string(firstName[0]) + lastName
	if identifier == pattern1 {
		return true
	}

	// Pattern: first name + last initial (e.g., "alices")
	pattern2 := firstName + string(lastName[0])
	if identifier == pattern2 {
		return true
	}

	// Pattern: first + last (e.g., "alicesmith")
	pattern3 := firstName + lastName
	if identifier == pattern3 {
		return true
	}

	return false
}
