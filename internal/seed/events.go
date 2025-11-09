package seed

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SeedEvents generates 12 months of realistic event data
func SeedEvents(data *SeedData) error {
	fmt.Println("\n5. Generating events (12 months)...")

	// Seed random for deterministic data
	rand.Seed(42)

	events := []*models.Event{}
	eventCounter := 1

	// Generate 12 months of data ending TODAY
	// Start 12 months ago, generate through yesterday (not today to avoid timezone issues)
	endDate := data.Now.AddDate(0, 0, -1)  // Yesterday
	startDate := endDate.AddDate(-1, 0, 1) // 12 months before yesterday, plus 1 day
	totalEvents := 0

	for month := 0; month < 12; month++ {
		monthStart := startDate.AddDate(0, month, 0)
		monthEvents := generateMonthEvents(data, monthStart, &eventCounter)
		events = append(events, monthEvents...)
		totalEvents += len(monthEvents)

		if month%3 == 2 {
			fmt.Printf("  Generated Q%d (%d events)\n", (month/3)+1, len(monthEvents))
		}
	}

	// Generate remaining days up to today to ensure current week has data
	// The 12-month loop generates 4 weeks per month (336 days total)
	// We need to fill the gap from the last generated day to today
	lastMonthStart := startDate.AddDate(0, 11, 0)        // Start of month 11 (last month)
	lastDayGenerated := lastMonthStart.AddDate(0, 0, 27) // After 4 weeks (day 27)
	currentDate := lastDayGenerated.AddDate(0, 0, 1)     // Start from day after last generated

	remainingEvents := 0
	for currentDate.Before(data.Now) || currentDate.Equal(data.Now.Truncate(24*time.Hour)) {
		// Only generate events for weekdays
		if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
			seasonalMultiplier := getSeasonalMultiplier(currentDate.Month())
			dayEvents := generateDayEvents(data, currentDate, seasonalMultiplier, &eventCounter)
			events = append(events, dayEvents...)
			totalEvents += len(dayEvents)
			remainingEvents += len(dayEvents)
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	if remainingEvents > 0 {
		fmt.Printf("  Generated remaining days through today (%d events)\n", remainingEvents)
	}

	// Save events in batches (to avoid transaction timeouts)
	batchSize := 1000
	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}
		batch := events[i:end]

		if err := data.EventStore.CreateBatch(batch); err != nil {
			return fmt.Errorf("failed to save events batch: %w", err)
		}
	}

	fmt.Printf("\nTotal events generated: %d over 12 months\n", totalEvents)
	fmt.Printf("Average per engineer per week: %.1f\n", float64(totalEvents)/(float64(len(data.EngineerIDs))*52))

	return nil
}

func generateMonthEvents(data *SeedData, monthStart time.Time, counter *int) []*models.Event {
	events := []*models.Event{}

	// Seasonal patterns
	seasonalMultiplier := getSeasonalMultiplier(monthStart.Month())

	// Generate ~4 weeks of events per month
	for week := 0; week < 4; week++ {
		weekStart := monthStart.AddDate(0, 0, week*7)
		weekEvents := generateWeekEvents(data, weekStart, seasonalMultiplier, counter)
		events = append(events, weekEvents...)
	}

	return events
}

func generateWeekEvents(data *SeedData, weekStart time.Time, seasonalMultiplier float64, counter *int) []*models.Event {
	events := []*models.Event{}

	// Generate events for each weekday (Mon-Fri)
	for day := 0; day < 5; day++ {
		dayStart := weekStart.AddDate(0, 0, day)
		dayEvents := generateDayEvents(data, dayStart, seasonalMultiplier, counter)
		events = append(events, dayEvents...)
	}

	return events
}

func generateDayEvents(data *SeedData, dayStart time.Time, seasonalMultiplier float64, counter *int) []*models.Event {
	events := []*models.Event{}

	// Each engineer gets events based on their seniority
	for i, engineerID := range data.EngineerIDs {
		eng := data.EngineerMap[engineerID]

		// Get activity level based on role
		activityLevel := getActivityLevel(eng.Role)
		activityLevel = activityLevel * seasonalMultiplier

		// Distribute events throughout the day (9am-6pm local time)
		// More events early in the week (Mon/Tue), fewer on Friday
		dayOfWeek := dayStart.Weekday()
		dayMultiplier := getDayMultiplier(dayOfWeek)

		numEvents := int(activityLevel * dayMultiplier)
		if numEvents < 1 {
			numEvents = 1
		}

		for e := 0; e < numEvents; e++ {
			event := generateSingleEvent(data, eng, engineerID, dayStart, i, counter)
			events = append(events, event)
		}
	}

	return events
}

func generateSingleEvent(data *SeedData, eng EngineerDef, engineerID string, dayStart time.Time, engineerIndex int, counter *int) *models.Event {
	// Random time during work hours (9am-6pm)
	hour := 9 + rand.Intn(9)
	minute := rand.Intn(60)
	timestamp := time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), hour, minute, 0, 0, time.UTC)

	// Event type distribution varies by role
	eventType, source := selectEventType(eng.Role, engineerIndex, *counter)

	// Get identifier for this source
	identifier := eng.Identifiers[source]
	if identifier == "" {
		identifier = eng.Email
	}

	event := &models.Event{
		Type:       eventType,
		Source:     source,
		SourceID:   fmt.Sprintf("seed-%s-%d", eventType, *counter),
		Timestamp:  timestamp,
		Actor:      identifier,
		EngineerID: engineerID,
		Data:       generateEventData(eventType, eng.Name, *counter),
		Anonymized: false,
	}

	*counter++
	return event
}

func getActivityLevel(role string) float64 {
	switch role {
	case "Junior Engineer":
		return 3.0 // 3 events per day
	case "Mid-Level Engineer":
		return 5.0 // 5 events per day
	case "Senior Engineer":
		return 8.0 // 8 events per day
	case "Staff Engineer":
		return 6.0 // 6 events per day (more reviews, fewer PRs)
	default:
		return 4.0
	}
}

func getSeasonalMultiplier(month time.Month) float64 {
	switch month {
	case time.December, time.January:
		return 0.6 // Holiday season - reduced activity
	case time.July, time.August:
		return 0.8 // Summer vacation - slightly reduced
	case time.April, time.May, time.September, time.October:
		return 1.2 // Peak productivity quarters
	default:
		return 1.0 // Normal activity
	}
}

func getDayMultiplier(day time.Weekday) float64 {
	switch day {
	case time.Monday:
		return 1.2 // More PRs at start of week
	case time.Tuesday, time.Wednesday:
		return 1.1 // High productivity mid-week
	case time.Thursday:
		return 0.9 // Winding down
	case time.Friday:
		return 0.7 // Lower activity end of week
	default:
		return 0.0 // No weekend events
	}
}

func selectEventType(role string, engineerIndex int, counter int) (string, string) {
	// Different roles have different event type distributions
	roll := counter % 100

	switch role {
	case "Staff Engineer", "Senior Engineer":
		// More reviews and architectural work
		if roll < 30 {
			return "pull_request", "github"
		} else if roll < 50 {
			return "commit", "github"
		} else if roll < 75 {
			return "code_review", "github"
		} else if roll < 85 {
			return "issue", "jira"
		} else {
			return "deployment", "github"
		}

	case "Mid-Level Engineer":
		// Balanced distribution
		if roll < 35 {
			return "pull_request", "github"
		} else if roll < 60 {
			return "commit", "github"
		} else if roll < 75 {
			return "code_review", "github"
		} else if roll < 90 {
			return "issue", "jira"
		} else {
			return "deployment", "github"
		}

	case "Junior Engineer":
		// More commits, fewer reviews
		if roll < 25 {
			return "pull_request", "github"
		} else if roll < 70 {
			return "commit", "github"
		} else if roll < 80 {
			return "code_review", "github"
		} else {
			return "issue", "jira"
		}

	default:
		return "commit", "github"
	}
}

func generateEventData(eventType string, name string, counter int) map[string]interface{} {
	data := make(map[string]interface{})

	switch eventType {
	case "pull_request":
		data["title"] = fmt.Sprintf("Feature: %s", getFeatureName(counter))
		data["state"] = getRandomState()
		data["additions"] = 50 + rand.Intn(500)
		data["deletions"] = 10 + rand.Intn(200)
		data["commits"] = 1 + rand.Intn(10)

	case "commit":
		data["message"] = fmt.Sprintf("Update %s", getComponentName(counter))
		data["additions"] = 5 + rand.Intn(100)
		data["deletions"] = 1 + rand.Intn(50)

	case "code_review":
		data["pull_request"] = fmt.Sprintf("PR #%d", 1000+counter)
		data["state"] = "approved"
		data["comments"] = rand.Intn(10)

	case "issue":
		data["title"] = fmt.Sprintf("Issue: %s", getIssueName(counter))
		data["type"] = getIssueType()
		data["priority"] = getPriority()

	case "deployment":
		data["environment"] = getEnvironment(counter)
		data["status"] = "success"
		data["commits"] = 3 + rand.Intn(15)
	}

	return data
}

func getFeatureName(counter int) string {
	features := []string{
		"User authentication improvements",
		"Dashboard performance optimization",
		"API rate limiting",
		"Database migration for new schema",
		"Enhanced error handling",
		"Add metrics collection",
		"Implement caching layer",
		"Refactor payment processing",
		"Add search functionality",
		"Improve mobile responsiveness",
	}
	return features[counter%len(features)]
}

func getComponentName(counter int) string {
	components := []string{
		"user service",
		"auth middleware",
		"database schema",
		"API handlers",
		"frontend components",
		"configuration",
		"tests",
		"documentation",
	}
	return components[counter%len(components)]
}

func getIssueName(counter int) string {
	issues := []string{
		"Fix login timeout",
		"Improve dashboard load time",
		"Add validation for user input",
		"Update dependencies",
		"Fix memory leak in worker",
		"Improve error messages",
		"Add logging for audit trail",
		"Optimize database queries",
	}
	return issues[counter%len(issues)]
}

func getRandomState() string {
	states := []string{"open", "merged", "closed"}
	return states[rand.Intn(len(states))]
}

func getIssueType() string {
	types := []string{"bug", "feature", "task", "improvement"}
	return types[rand.Intn(len(types))]
}

func getPriority() string {
	priorities := []string{"low", "medium", "high", "critical"}
	return priorities[rand.Intn(len(priorities))]
}

func getEnvironment(counter int) string {
	// More production deployments from senior engineers
	if counter%5 == 0 {
		return "production"
	} else if counter%3 == 0 {
		return "staging"
	}
	return "development"
}
