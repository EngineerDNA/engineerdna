package db

import (
	"database/sql"
	"fmt"
)

type EngineerStore struct {
	db *sql.DB
}

func NewEngineerStore(db *sql.DB) *EngineerStore {
	return &EngineerStore{db: db}
}

// GetEngineerName retrieves an engineer's name by ID
func (s *EngineerStore) GetEngineerName(id string) (string, error) {
	var name string
	err := s.db.QueryRow("SELECT name FROM engineers WHERE id = ?", id).Scan(&name)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("engineer not found: %s", id)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get engineer name: %w", err)
	}
	return name, nil
}

// EngineerPerformance represents performance data for an engineer
type EngineerPerformance struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Score   float64 `json:"score"`
	PRCount float64 `json:"pr_count"`
}

// GetPerformanceTable retrieves engineer performance data with pagination and sorting
func (s *EngineerStore) GetPerformanceTable(limit, offset int, sortBy string) ([]EngineerPerformance, error) {
	// Validate sort_by parameter to prevent SQL injection
	allowedSortColumns := map[string]bool{
		"score":    true,
		"pr_count": true,
		"name":     true,
	}

	if !allowedSortColumns[sortBy] {
		sortBy = "score"
	}

	query := `
		SELECT
			e.id,
			e.canonical_name as name,
			e.email,
			COALESCE(ps.total_score, 0) as score,
			COUNT(DISTINCT ev.id) as pr_count
		FROM engineers e
		LEFT JOIN performance_scores ps ON e.id = ps.engineer_id
		LEFT JOIN events ev ON e.email = ev.actor AND ev.type = 'pull_request'
		GROUP BY e.id, e.canonical_name, e.email, ps.total_score
		ORDER BY ` + sortBy + ` DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query engineer performance: %w", err)
	}
	defer rows.Close()

	var engineers []EngineerPerformance
	for rows.Next() {
		var eng EngineerPerformance
		if err := rows.Scan(&eng.ID, &eng.Name, &eng.Email, &eng.Score, &eng.PRCount); err != nil {
			return nil, fmt.Errorf("failed to scan engineer row: %w", err)
		}
		engineers = append(engineers, eng)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating engineer rows: %w", err)
	}

	return engineers, nil
}
