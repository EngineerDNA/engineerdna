package identity

import (
	"database/sql"
	"testing"

	"github.com/engineerdna/engineerdna/internal/db"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *db.IdentityStore) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	_, err = database.Exec(`
		CREATE TABLE engineers (
			id TEXT PRIMARY KEY,
			canonical_name TEXT NOT NULL,
			email TEXT,
			manager TEXT,
			identifiers TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE unresolved_identities (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			identifier TEXT NOT NULL,
			first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			event_count INTEGER NOT NULL DEFAULT 1,
			ignored BOOLEAN NOT NULL DEFAULT 0,
			UNIQUE(source, identifier)
		);

		CREATE TABLE events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			source TEXT NOT NULL,
			source_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			actor TEXT NOT NULL,
			engineer_id TEXT,
			data TEXT NOT NULL,
			anonymized BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	store := db.NewIdentityStore(database)
	return database, store
}

func TestCreateEngineer(t *testing.T) {
	database, store := setupTestDB(t)
	defer database.Close()

	service := NewService(store)

	identifiers := map[string]string{
		"github": "asmith",
		"jira":   "alice.smith@example.com",
	}

	engineerID, err := service.CreateEngineer("Alice Smith", "alice@example.com", "Bob Manager", identifiers)
	if err != nil {
		t.Fatalf("Failed to create engineer: %v", err)
	}

	if engineerID == "" {
		t.Fatal("Expected non-empty engineer ID")
	}

	// Verify engineer was created
	engineer, err := service.GetEngineer(engineerID)
	if err != nil {
		t.Fatalf("Failed to get engineer: %v", err)
	}

	if engineer == nil {
		t.Fatal("Expected engineer to exist")
	}

	if engineer.Name != "Alice Smith" {
		t.Errorf("Expected name 'Alice Smith', got '%s'", engineer.Name)
	}

	if engineer.Email != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got '%s'", engineer.Email)
	}

	if len(engineer.Identifiers) != 2 {
		t.Errorf("Expected 2 identifiers, got %d", len(engineer.Identifiers))
	}
}

func TestResolveIdentity(t *testing.T) {
	database, store := setupTestDB(t)
	defer database.Close()

	service := NewService(store)

	// Create an engineer with GitHub identifier
	identifiers := map[string]string{
		"github": "asmith",
	}
	engineerID, err := service.CreateEngineer("Alice Smith", "alice@example.com", "", identifiers)
	if err != nil {
		t.Fatalf("Failed to create engineer: %v", err)
	}

	// Resolve the same identifier - should return engineer ID
	resolvedID, err := service.ResolveIdentity("github", "asmith")
	if err != nil {
		t.Fatalf("Failed to resolve identity: %v", err)
	}

	if resolvedID != engineerID {
		t.Errorf("Expected resolved ID '%s', got '%s'", engineerID, resolvedID)
	}

	// Resolve unknown identifier - should return empty string and create unresolved
	resolvedID2, err := service.ResolveIdentity("jira", "bob@example.com")
	if err != nil {
		t.Fatalf("Failed to resolve identity: %v", err)
	}

	if resolvedID2 != "" {
		t.Errorf("Expected empty string for unresolved identity, got '%s'", resolvedID2)
	}

	// Verify unresolved identity was created
	unresolved, err := service.GetUnresolved()
	if err != nil {
		t.Fatalf("Failed to get unresolved: %v", err)
	}

	if len(unresolved) != 1 {
		t.Fatalf("Expected 1 unresolved identity, got %d", len(unresolved))
	}

	if unresolved[0].Source != "jira" || unresolved[0].Identifier != "bob@example.com" {
		t.Errorf("Unexpected unresolved identity: %+v", unresolved[0])
	}
}

func TestAssignIdentity(t *testing.T) {
	database, store := setupTestDB(t)
	defer database.Close()

	service := NewService(store)

	// Create an engineer
	engineerID, err := service.CreateEngineer("Alice Smith", "alice@example.com", "", map[string]string{
		"github": "asmith",
	})
	if err != nil {
		t.Fatalf("Failed to create engineer: %v", err)
	}

	// Create an unresolved identity by attempting to resolve unknown identifier
	_, err = service.ResolveIdentity("jira", "alice.smith@company.com")
	if err != nil {
		t.Fatalf("Failed to create unresolved: %v", err)
	}

	// Assign the unresolved identity to the engineer
	err = service.AssignIdentity(engineerID, "jira", "alice.smith@company.com")
	if err != nil {
		t.Fatalf("Failed to assign identity: %v", err)
	}

	// Verify the identifier was added to engineer
	engineer, err := service.GetEngineer(engineerID)
	if err != nil {
		t.Fatalf("Failed to get engineer: %v", err)
	}

	if engineer.Identifiers["jira"] != "alice.smith@company.com" {
		t.Errorf("Expected jira identifier 'alice.smith@company.com', got '%s'", engineer.Identifiers["jira"])
	}

	// Verify unresolved identity was removed
	unresolved, err := service.GetUnresolved()
	if err != nil {
		t.Fatalf("Failed to get unresolved: %v", err)
	}

	if len(unresolved) != 0 {
		t.Errorf("Expected 0 unresolved identities after assignment, got %d", len(unresolved))
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "adc", 1},
		{"kitten", "sitting", 3},
		{"alice", "alise", 1},
		{"smith", "smyth", 1},
	}

	for _, tt := range tests {
		result := levenshteinDistance(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d", tt.s1, tt.s2, result, tt.expected)
		}
	}
}

func TestIsUsernamePattern(t *testing.T) {
	tests := []struct {
		identifier string
		name       string
		expected   bool
	}{
		{"asmith", "Alice Smith", true},     // first initial + last name
		{"alices", "Alice Smith", true},     // first name + last initial
		{"alicesmith", "Alice Smith", true}, // first + last
		{"bob", "Bob", false},               // only one name part
		{"xyz", "Alice Smith", false},       // doesn't match
		{"asmith", "Andrew Smith", true},    // matches different first name
	}

	for _, tt := range tests {
		result := isUsernamePattern(tt.identifier, tt.name)
		if result != tt.expected {
			t.Errorf("isUsernamePattern(%q, %q) = %v, expected %v", tt.identifier, tt.name, result, tt.expected)
		}
	}
}

func TestAssignIdentityUpdatesAllEvents(t *testing.T) {
	database, store := setupTestDB(t)
	defer database.Close()

	service := NewService(store)

	// Create an engineer
	engineerID, err := service.CreateEngineer("Alice Smith", "alice@example.com", "", map[string]string{
		"github": "asmith",
	})
	if err != nil {
		t.Fatalf("Failed to create engineer: %v", err)
	}

	// Create an unresolved identity by attempting to resolve unknown identifier
	unresolvedID, err := service.ResolveIdentity("jira", "alice.smith@company.com")
	if err != nil {
		t.Fatalf("Failed to create unresolved: %v", err)
	}

	// Get the unresolved identity ID
	unresolved, err := service.GetUnresolved()
	if err != nil {
		t.Fatalf("Failed to get unresolved: %v", err)
	}
	if len(unresolved) != 1 {
		t.Fatalf("Expected 1 unresolved identity, got %d", len(unresolved))
	}
	unresolvedID = unresolved[0].ID

	// Create events with the unresolved identity ID
	// This simulates what happens during seeding
	_, err = database.Exec(`
		INSERT INTO events (id, type, source, source_id, timestamp, actor, engineer_id, data)
		VALUES
			('event1', 'issue', 'jira', 'PROJ-1', datetime('now'), 'alice.smith@company.com', ?, '{}'),
			('event2', 'issue', 'jira', 'PROJ-2', datetime('now'), 'alice.smith@company.com', ?, '{}'),
			('event3', 'issue', 'jira', 'PROJ-3', datetime('now'), 'alice.smith@company.com', ?, '{}')
	`, unresolvedID, unresolvedID, unresolvedID)
	if err != nil {
		t.Fatalf("Failed to create events: %v", err)
	}

	// Verify events have unresolved identity ID
	var count int
	err = database.QueryRow(`SELECT COUNT(*) FROM events WHERE engineer_id = ?`, unresolvedID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count events: %v", err)
	}
	if count != 3 {
		t.Fatalf("Expected 3 events with unresolved ID, got %d", count)
	}

	// Assign the unresolved identity to the engineer
	err = service.AssignIdentity(engineerID, "jira", "alice.smith@company.com")
	if err != nil {
		t.Fatalf("Failed to assign identity: %v", err)
	}

	// Verify ALL events are now updated to engineer ID
	err = database.QueryRow(`SELECT COUNT(*) FROM events WHERE engineer_id = ?`, engineerID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count events: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 events with engineer ID after assignment, got %d", count)
	}

	// Verify NO events still have unresolved identity ID (orphaned references)
	err = database.QueryRow(`SELECT COUNT(*) FROM events WHERE engineer_id = ?`, unresolvedID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count orphaned events: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 events with orphaned unresolved ID, got %d (BUG: orphaned foreign keys)", count)
	}

	// Verify engineer_id matches for all events
	var allEventsHaveEngineerID bool
	err = database.QueryRow(`
		SELECT COUNT(*) = 3
		FROM events
		WHERE source = 'jira' AND actor = 'alice.smith@company.com' AND engineer_id = ?
	`, engineerID).Scan(&allEventsHaveEngineerID)
	if err != nil {
		t.Fatalf("Failed to verify events: %v", err)
	}
	if !allEventsHaveEngineerID {
		t.Error("Not all events were updated with engineer ID")
	}
}
