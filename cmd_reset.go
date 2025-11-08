package main

import (
	"fmt"
	"log"

	"github.com/engineerdna/engineerdna/internal/config"
)

func cmdReset() {
	fmt.Println("Removing demo data...")

	cfg := config.DefaultConfig()

	// Initialize database
	database, err := config.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Count demo data across all tables
	var eventCount, engineerCount, teamCount, roleCount, scoreCount, briefingCount, sprintCount, storyCount int

	if err := database.DB.QueryRow("SELECT COUNT(*) FROM events WHERE source IN ('demo', 'github', 'jira')").Scan(&eventCount); err != nil {
		log.Printf("Warning: Failed to count events: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM engineers WHERE email LIKE '%@example.com'").Scan(&engineerCount); err != nil {
		log.Printf("Warning: Failed to count engineers: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM teams").Scan(&teamCount); err != nil {
		log.Printf("Warning: Failed to count teams: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM roles").Scan(&roleCount); err != nil {
		log.Printf("Warning: Failed to count roles: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM performance_scores").Scan(&scoreCount); err != nil {
		log.Printf("Warning: Failed to count performance scores: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM weekly_briefings").Scan(&briefingCount); err != nil {
		log.Printf("Warning: Failed to count briefings: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM sprints").Scan(&sprintCount); err != nil {
		log.Printf("Warning: Failed to count sprints: %v", err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM stories").Scan(&storyCount); err != nil {
		log.Printf("Warning: Failed to count stories: %v", err)
	}

	totalCount := eventCount + engineerCount + teamCount + roleCount + scoreCount + briefingCount + sprintCount + storyCount
	if totalCount == 0 {
		fmt.Println("No demo data found")
		return
	}

	// Show what will be deleted
	fmt.Println("\nFound demo data:")
	if eventCount > 0 {
		fmt.Printf("  %d events\n", eventCount)
	}
	if scoreCount > 0 {
		fmt.Printf("  %d performance scores\n", scoreCount)
	}
	if briefingCount > 0 {
		fmt.Printf("  %d weekly briefings\n", briefingCount)
	}
	if storyCount > 0 {
		fmt.Printf("  %d stories\n", storyCount)
	}
	if sprintCount > 0 {
		fmt.Printf("  %d sprints\n", sprintCount)
	}
	if engineerCount > 0 {
		fmt.Printf("  %d engineers\n", engineerCount)
	}
	if teamCount > 0 {
		fmt.Printf("  %d teams\n", teamCount)
	}
	if roleCount > 0 {
		fmt.Printf("  %d roles\n", roleCount)
	}

	// Confirm deletion
	fmt.Print("\nDelete all demo data? [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" {
		fmt.Println("Cancelled")
		return
	}

	// Delete in correct order (respecting foreign keys)
	deleted := 0

	// 1. Delete stories (references sprints and engineers)
	if storyCount > 0 {
		result, err := database.DB.Exec("DELETE FROM stories")
		if err != nil {
			log.Printf("Warning: Failed to delete stories: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d stories\n", rows)
		}
	}

	// 2. Delete sprints
	if sprintCount > 0 {
		result, err := database.DB.Exec("DELETE FROM sprints")
		if err != nil {
			log.Printf("Warning: Failed to delete sprints: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d sprints\n", rows)
		}
	}

	// 3. Delete weekly briefings
	if briefingCount > 0 {
		result, err := database.DB.Exec("DELETE FROM weekly_briefings")
		if err != nil {
			log.Printf("Warning: Failed to delete briefings: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d weekly briefings\n", rows)
		}
	}

	// 4. Delete performance scores
	if scoreCount > 0 {
		result, err := database.DB.Exec("DELETE FROM performance_scores")
		if err != nil {
			log.Printf("Warning: Failed to delete performance scores: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d performance scores\n", rows)
		}
	}

	// 5. Delete team performance scores
	result, err := database.DB.Exec("DELETE FROM team_performance_scores")
	if err != nil {
		log.Printf("Warning: Failed to delete team performance scores: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		if rows > 0 {
			deleted += int(rows)
			fmt.Printf("Deleted %d team performance scores\n", rows)
		}
	}

	// 6. Delete events
	if eventCount > 0 {
		result, err := database.DB.Exec("DELETE FROM events WHERE source IN ('demo', 'github', 'jira')")
		if err != nil {
			log.Printf("Warning: Failed to delete events: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d events\n", rows)
		}
	}

	// 7. Delete team memberships
	result, err = database.DB.Exec("DELETE FROM team_membership")
	if err != nil {
		log.Printf("Warning: Failed to delete team memberships: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		if rows > 0 {
			deleted += int(rows)
			fmt.Printf("Deleted %d team memberships\n", rows)
		}
	}

	// 8. Delete teams
	if teamCount > 0 {
		result, err := database.DB.Exec("DELETE FROM teams")
		if err != nil {
			log.Printf("Warning: Failed to delete teams: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d teams\n", rows)
		}
	}

	// 9. Delete engineers
	if engineerCount > 0 {
		result, err := database.DB.Exec("DELETE FROM engineers WHERE email LIKE '%@example.com'")
		if err != nil {
			log.Printf("Warning: Failed to delete engineers: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d engineers\n", rows)
		}
	}

	// 10. Delete roles
	if roleCount > 0 {
		result, err := database.DB.Exec("DELETE FROM roles")
		if err != nil {
			log.Printf("Warning: Failed to delete roles: %v", err)
		} else {
			rows, _ := result.RowsAffected()
			deleted += int(rows)
			fmt.Printf("Deleted %d roles\n", rows)
		}
	}

	// 11. Delete scoring weights
	result, err = database.DB.Exec("DELETE FROM scoring_weights")
	if err != nil {
		log.Printf("Warning: Failed to delete scoring weights: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		if rows > 0 {
			deleted += int(rows)
			fmt.Printf("Deleted %d scoring weights\n", rows)
		}
	}

	// 12. Delete unresolved identities
	result, err = database.DB.Exec("DELETE FROM unresolved_identities")
	if err != nil {
		log.Printf("Warning: Failed to delete unresolved identities: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		if rows > 0 {
			deleted += int(rows)
			fmt.Printf("Deleted %d unresolved identities\n", rows)
		}
	}

	fmt.Printf("\n========================================")
	fmt.Printf("\nTotal: Deleted %d demo records\n", deleted)
	fmt.Println("Database reset completed successfully!")
	fmt.Println("========================================")
}
