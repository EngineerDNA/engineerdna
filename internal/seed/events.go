package seed

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SeedEvents generates demo events over 8 weeks
func SeedEvents(data *SeedData) error {
	fmt.Println("\n5. Generating events...")
	events := []*models.Event{}

	eventCounter := 1
	for weekOffset := 0; weekOffset < 8; weekOffset++ {
		weekStart := data.Now.AddDate(0, 0, -7*weekOffset)

		// More recent weeks have more events
		eventsThisWeek := 30 + (8-weekOffset)*2

		for i := 0; i < eventsThisWeek; i++ {
			// Pick random engineer
			engIndex := i % len(data.EngineerIDs)
			engineerID := data.EngineerIDs[engIndex]
			eng := data.EngineerMap[engineerID]

			// Pick event type based on weights
			typeIndex := i % 10
			var eventType, source string
			if typeIndex < 4 {
				eventType = "pull_request"
				source = "github"
			} else if typeIndex < 7 {
				eventType = "commit"
				source = "github"
			} else if typeIndex < 9 {
				eventType = "issue"
				source = "jira"
			} else {
				eventType = "code_review"
				source = "github"
			}

			// Random day within the week
			dayOffset := i % 7
			timestamp := weekStart.AddDate(0, 0, -dayOffset)

			// Get identifier for this source
			identifier := eng.Identifiers[source]
			if identifier == "" {
				identifier = eng.Email
			}

			event := &models.Event{
				Type:       eventType,
				Source:     source,
				SourceID:   fmt.Sprintf("demo-%s-%d", eventType, eventCounter),
				Timestamp:  timestamp,
				Actor:      identifier,
				EngineerID: engineerID,
				Data: map[string]interface{}{
					"title":  fmt.Sprintf("Demo %s #%d", eventType, eventCounter),
					"status": "completed",
				},
				Anonymized: false,
			}

			events = append(events, event)
			eventCounter++
		}
	}

	// Save events in batch
	if err := data.EventStore.CreateBatch(events); err != nil {
		log.Fatalf("Failed to save demo events: %v", err)
	}
	fmt.Printf("  Created %d events over 8 weeks\n", len(events))

	return nil
}
