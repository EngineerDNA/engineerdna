package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
	"github.com/google/uuid"
)

// CostROIStore handles cost and ROI storage operations
type CostROIStore struct {
	db *sql.DB
}

// NewCostROIStore creates a new cost and ROI store
func NewCostROIStore(db *sql.DB) *CostROIStore {
	return &CostROIStore{db: db}
}

// Feature Value operations

// CreateFeatureValue creates a new feature value
func (s *CostROIStore) CreateFeatureValue(feature *models.FeatureValue) error {
	if feature.ID == "" {
		feature.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if feature.CreatedAt.IsZero() {
		feature.CreatedAt = now
	}
	feature.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO feature_values (
			id, feature_name, feature_description, value_type, value_amount,
			value_currency, confidence_level, source, source_reference,
			time_period, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		feature.ID, feature.FeatureName, feature.FeatureDescription,
		feature.ValueType, feature.ValueAmount, feature.ValueCurrency,
		feature.ConfidenceLevel, feature.Source, feature.SourceReference,
		feature.TimePeriod, feature.CreatedAt, feature.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create feature value: %w", err)
	}

	return nil
}

// GetFeatureValue retrieves a feature value by ID
func (s *CostROIStore) GetFeatureValue(id string) (*models.FeatureValue, error) {
	var feature models.FeatureValue
	var description, source, sourceRef, timePeriod sql.NullString
	var valueAmount sql.NullFloat64
	var createdAtStr, updatedAtStr string

	err := s.db.QueryRow(`
		SELECT id, feature_name, feature_description, value_type, value_amount,
		       value_currency, confidence_level, source, source_reference,
		       time_period, created_at, updated_at
		FROM feature_values
		WHERE id = ?
	`, id).Scan(
		&feature.ID, &feature.FeatureName, &description, &feature.ValueType,
		&valueAmount, &feature.ValueCurrency, &feature.ConfidenceLevel,
		&source, &sourceRef, &timePeriod, &createdAtStr, &updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feature value: %w", err)
	}

	if description.Valid {
		feature.FeatureDescription = description.String
	}
	if valueAmount.Valid {
		feature.ValueAmount = &valueAmount.Float64
	}
	if source.Valid {
		feature.Source = source.String
	}
	if sourceRef.Valid {
		feature.SourceReference = sourceRef.String
	}
	if timePeriod.Valid {
		feature.TimePeriod = timePeriod.String
	}

	// Parse timestamps
	feature.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	feature.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &feature, nil
}

// ListFeatureValues retrieves feature values with optional filtering
func (s *CostROIStore) ListFeatureValues(timePeriod string, limit, offset int) ([]*models.FeatureValue, int, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = DefaultQueryLimit
	}
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	// Build WHERE clause
	whereClause := " WHERE 1=1"
	args := []interface{}{}

	if timePeriod != "" {
		whereClause += " AND time_period = ?"
		args = append(args, timePeriod)
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM feature_values" + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count feature values: %w", err)
	}

	// Query with pagination
	query := `
		SELECT id, feature_name, feature_description, value_type, value_amount,
		       value_currency, confidence_level, source, source_reference,
		       time_period, created_at, updated_at
		FROM feature_values` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	paginatedArgs := append(args, limit, offset)
	rows, err := s.db.Query(query, paginatedArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list feature values: %w", err)
	}
	defer rows.Close()

	var features []*models.FeatureValue
	for rows.Next() {
		var feature models.FeatureValue
		var description, source, sourceRef, timePeriod sql.NullString
		var valueAmount sql.NullFloat64
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&feature.ID, &feature.FeatureName, &description, &feature.ValueType,
			&valueAmount, &feature.ValueCurrency, &feature.ConfidenceLevel,
			&source, &sourceRef, &timePeriod, &createdAtStr, &updatedAtStr,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan feature value: %w", err)
		}

		if description.Valid {
			feature.FeatureDescription = description.String
		}
		if valueAmount.Valid {
			feature.ValueAmount = &valueAmount.Float64
		}
		if source.Valid {
			feature.Source = source.String
		}
		if sourceRef.Valid {
			feature.SourceReference = sourceRef.String
		}
		if timePeriod.Valid {
			feature.TimePeriod = timePeriod.String
		}

		// Parse timestamps
		feature.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		feature.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

		features = append(features, &feature)
	}

	return features, total, rows.Err()
}

// UpdateFeatureValue updates a feature value
func (s *CostROIStore) UpdateFeatureValue(feature *models.FeatureValue) error {
	feature.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE feature_values
		SET feature_name = ?, feature_description = ?, value_type = ?,
		    value_amount = ?, value_currency = ?, confidence_level = ?,
		    source = ?, source_reference = ?, time_period = ?, updated_at = ?
		WHERE id = ?
	`,
		feature.FeatureName, feature.FeatureDescription, feature.ValueType,
		feature.ValueAmount, feature.ValueCurrency, feature.ConfidenceLevel,
		feature.Source, feature.SourceReference, feature.TimePeriod,
		feature.UpdatedAt, feature.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update feature value: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("feature value not found: %s", feature.ID)
	}

	return nil
}

// Feature Work Item operations

// CreateFeatureWorkItem creates a new feature work item
func (s *CostROIStore) CreateFeatureWorkItem(item *models.FeatureWorkItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO feature_work_items (
			id, feature_id, work_item_type, work_item_id, story_points,
			actual_hours, engineer_id, team_id, completed_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		item.ID, item.FeatureID, item.WorkItemType, item.WorkItemID,
		item.StoryPoints, item.ActualHours, item.EngineerID, item.TeamID,
		item.CompletedAt, item.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create feature work item: %w", err)
	}

	return nil
}

// GetFeatureWorkItems retrieves work items for a feature
func (s *CostROIStore) GetFeatureWorkItems(featureID string, limit int) ([]*models.FeatureWorkItem, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = DefaultQueryLimit
	}
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, feature_id, work_item_type, work_item_id, story_points,
		       actual_hours, engineer_id, team_id, completed_at, created_at
		FROM feature_work_items
		WHERE feature_id = ?
		ORDER BY created_at ASC
		LIMIT ?
	`, featureID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature work items: %w", err)
	}
	defer rows.Close()

	var items []*models.FeatureWorkItem
	for rows.Next() {
		var item models.FeatureWorkItem
		var storyPoints sql.NullInt64
		var actualHours sql.NullFloat64
		var engineerID, teamID sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&item.ID, &item.FeatureID, &item.WorkItemType, &item.WorkItemID,
			&storyPoints, &actualHours, &engineerID, &teamID,
			&completedAt, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feature work item: %w", err)
		}

		if storyPoints.Valid {
			sp := int(storyPoints.Int64)
			item.StoryPoints = &sp
		}
		if actualHours.Valid {
			item.ActualHours = &actualHours.Float64
		}
		if engineerID.Valid {
			item.EngineerID = &engineerID.String
		}
		if teamID.Valid {
			item.TeamID = &teamID.String
		}
		if completedAt.Valid {
			item.CompletedAt = &completedAt.Time
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

// Feature Cost operations

// CreateFeatureCost creates or updates a feature cost
func (s *CostROIStore) CreateFeatureCost(cost *models.FeatureCost) error {
	if cost.ID == "" {
		cost.ID = uuid.New().String()
	}
	if cost.ComputedAt.IsZero() {
		cost.ComputedAt = time.Now().UTC()
	}

	// Delete existing cost for this feature
	s.db.Exec("DELETE FROM feature_costs WHERE feature_id = ?", cost.FeatureID)

	_, err := s.db.Exec(`
		INSERT INTO feature_costs (
			id, feature_id, total_story_points, total_hours, total_cost,
			cost_breakdown, computation_method, computed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		cost.ID, cost.FeatureID, cost.TotalStoryPoints, cost.TotalHours,
		cost.TotalCost, cost.CostBreakdown, cost.ComputationMethod, cost.ComputedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create feature cost: %w", err)
	}

	return nil
}

// GetFeatureCost retrieves cost for a feature
func (s *CostROIStore) GetFeatureCost(featureID string) (*models.FeatureCost, error) {
	var cost models.FeatureCost
	var breakdown sql.NullString

	err := s.db.QueryRow(`
		SELECT id, feature_id, total_story_points, total_hours, total_cost,
		       cost_breakdown, computation_method, computed_at
		FROM feature_costs
		WHERE feature_id = ?
	`, featureID).Scan(
		&cost.ID, &cost.FeatureID, &cost.TotalStoryPoints, &cost.TotalHours,
		&cost.TotalCost, &breakdown, &cost.ComputationMethod, &cost.ComputedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feature cost: %w", err)
	}

	if breakdown.Valid {
		cost.CostBreakdown = breakdown.String
	}

	return &cost, nil
}

// ROI Calculation operations

// CreateROICalculation creates or updates an ROI calculation
func (s *CostROIStore) CreateROICalculation(roi *models.ROICalculation) error {
	if roi.ID == "" {
		roi.ID = uuid.New().String()
	}
	if roi.CalculatedAt.IsZero() {
		roi.CalculatedAt = time.Now().UTC()
	}

	// Delete existing ROI for this feature
	s.db.Exec("DELETE FROM roi_calculations WHERE feature_id = ?", roi.FeatureID)

	_, err := s.db.Exec(`
		INSERT INTO roi_calculations (
			id, feature_id, investment, return_value, roi_percentage,
			payback_months, confidence, notes, calculated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		roi.ID, roi.FeatureID, roi.Investment, roi.ReturnValue,
		roi.ROIPercentage, roi.PaybackMonths, roi.Confidence,
		roi.Notes, roi.CalculatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create ROI calculation: %w", err)
	}

	return nil
}

// GetROICalculation retrieves ROI for a feature
func (s *CostROIStore) GetROICalculation(featureID string) (*models.ROICalculation, error) {
	var roi models.ROICalculation
	var returnValue, roiPct, paybackMonths sql.NullFloat64
	var notes sql.NullString

	err := s.db.QueryRow(`
		SELECT id, feature_id, investment, return_value, roi_percentage,
		       payback_months, confidence, notes, calculated_at
		FROM roi_calculations
		WHERE feature_id = ?
	`, featureID).Scan(
		&roi.ID, &roi.FeatureID, &roi.Investment, &returnValue, &roiPct,
		&paybackMonths, &roi.Confidence, &notes, &roi.CalculatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get ROI calculation: %w", err)
	}

	if returnValue.Valid {
		roi.ReturnValue = &returnValue.Float64
	}
	if roiPct.Valid {
		roi.ROIPercentage = &roiPct.Float64
	}
	if paybackMonths.Valid {
		roi.PaybackMonths = &paybackMonths.Float64
	}
	if notes.Valid {
		roi.Notes = notes.String
	}

	return &roi, nil
}

// ListROICalculations retrieves all ROI calculations
func (s *CostROIStore) ListROICalculations(limit int) ([]*models.ROICalculation, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = DefaultQueryLimit
	}
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	rows, err := s.db.Query(`
		SELECT id, feature_id, investment, return_value, roi_percentage,
		       payback_months, confidence, notes, calculated_at
		FROM roi_calculations
		ORDER BY calculated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list ROI calculations: %w", err)
	}
	defer rows.Close()

	var rois []*models.ROICalculation
	for rows.Next() {
		var roi models.ROICalculation
		var returnValue, roiPct, paybackMonths sql.NullFloat64
		var notes sql.NullString

		err := rows.Scan(
			&roi.ID, &roi.FeatureID, &roi.Investment, &returnValue, &roiPct,
			&paybackMonths, &roi.Confidence, &notes, &roi.CalculatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ROI calculation: %w", err)
		}

		if returnValue.Valid {
			roi.ReturnValue = &returnValue.Float64
		}
		if roiPct.Valid {
			roi.ROIPercentage = &roiPct.Float64
		}
		if paybackMonths.Valid {
			roi.PaybackMonths = &paybackMonths.Float64
		}
		if notes.Valid {
			roi.Notes = notes.String
		}

		rois = append(rois, &roi)
	}

	return rois, rows.Err()
}

// Engineering Investment operations

// CreateEngineeringInvestment creates or updates investment record
func (s *CostROIStore) CreateEngineeringInvestment(inv *models.EngineeringInvestment) error {
	if inv.ID == "" {
		inv.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if inv.CreatedAt.IsZero() {
		inv.CreatedAt = now
	}
	inv.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO engineering_investment (
			id, time_period, team_id, category, story_points,
			hours, cost, feature_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		inv.ID, inv.TimePeriod, inv.TeamID, inv.Category,
		inv.StoryPoints, inv.Hours, inv.Cost, inv.FeatureCount,
		inv.CreatedAt, inv.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create engineering investment: %w", err)
	}

	return nil
}

// GetEngineeringInvestment retrieves investment breakdown
func (s *CostROIStore) GetEngineeringInvestment(timePeriod string, teamID *string, limit int) ([]*models.EngineeringInvestment, error) {
	// Apply default and max limits for pagination
	if limit <= 0 {
		limit = DefaultQueryLimit
	}
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	query := `
		SELECT id, time_period, team_id, category, story_points,
		       hours, cost, feature_count, created_at, updated_at
		FROM engineering_investment
		WHERE time_period = ?
	`
	args := []interface{}{timePeriod}

	if teamID != nil {
		query += " AND team_id = ?"
		args = append(args, *teamID)
	}

	query += " ORDER BY category LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get engineering investment: %w", err)
	}
	defer rows.Close()

	var investments []*models.EngineeringInvestment
	for rows.Next() {
		var inv models.EngineeringInvestment
		var teamID sql.NullString

		err := rows.Scan(
			&inv.ID, &inv.TimePeriod, &teamID, &inv.Category,
			&inv.StoryPoints, &inv.Hours, &inv.Cost, &inv.FeatureCount,
			&inv.CreatedAt, &inv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan engineering investment: %w", err)
		}

		if teamID.Valid {
			inv.TeamID = &teamID.String
		}

		investments = append(investments, &inv)
	}

	return investments, rows.Err()
}
