package models

import "time"

// CostConfiguration defines engineering costs by entity and role
type CostConfiguration struct {
	ID            string     `json:"id"`
	EntityType    string     `json:"entity_type"`         // 'engineer', 'team', 'org'
	EntityID      *string    `json:"entity_id,omitempty"` // engineer_id, team_id, or null
	Role          *string    `json:"role,omitempty"`      // 'junior', 'mid', 'senior', 'staff', 'principal'
	MonthlyCost   float64    `json:"monthly_cost"`        // fully-loaded cost
	Currency      string     `json:"currency"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"` // null = current
	Notes         string     `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// FeatureValue represents business value of a feature
type FeatureValue struct {
	ID                 string    `json:"id"`
	FeatureName        string    `json:"feature_name"`
	FeatureDescription string    `json:"feature_description,omitempty"`
	ValueType          string    `json:"value_type"` // 'arr_impact', 'customer_acquisition', etc.
	ValueAmount        *float64  `json:"value_amount,omitempty"`
	ValueCurrency      string    `json:"value_currency"`
	ConfidenceLevel    string    `json:"confidence_level"` // 'estimated', 'actual', 'validated'
	Source             string    `json:"source,omitempty"` // 'manual', 'productboard', 'jira'
	SourceReference    string    `json:"source_reference,omitempty"`
	TimePeriod         string    `json:"time_period,omitempty"` // 'Q4 2024'
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// FeatureWorkItem links features to engineering work
type FeatureWorkItem struct {
	ID           string     `json:"id"`
	FeatureID    string     `json:"feature_id"`
	WorkItemType string     `json:"work_item_type"` // 'story', 'epic', 'pr', 'issue'
	WorkItemID   string     `json:"work_item_id"`   // external ID
	StoryPoints  *int       `json:"story_points,omitempty"`
	ActualHours  *float64   `json:"actual_hours,omitempty"`
	EngineerID   *string    `json:"engineer_id,omitempty"`
	TeamID       *string    `json:"team_id,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// FeatureCost represents computed costs for a feature
type FeatureCost struct {
	ID                string    `json:"id"`
	FeatureID         string    `json:"feature_id"`
	TotalStoryPoints  int       `json:"total_story_points"`
	TotalHours        float64   `json:"total_hours"`
	TotalCost         float64   `json:"total_cost"`
	CostBreakdown     string    `json:"cost_breakdown,omitempty"` // JSON
	ComputationMethod string    `json:"computation_method"`       // 'story_points', 'hours', 'hybrid'
	ComputedAt        time.Time `json:"computed_at"`
}

// ROICalculation represents feature ROI analysis
type ROICalculation struct {
	ID            string    `json:"id"`
	FeatureID     string    `json:"feature_id"`
	Investment    float64   `json:"investment"`             // total engineering cost
	ReturnValue   *float64  `json:"return_value,omitempty"` // business value
	ROIPercentage *float64  `json:"roi_percentage,omitempty"`
	PaybackMonths *float64  `json:"payback_months,omitempty"`
	Confidence    string    `json:"confidence"` // 'low', 'medium', 'high'
	Notes         string    `json:"notes,omitempty"`
	CalculatedAt  time.Time `json:"calculated_at"`
}

// EngineeringInvestment tracks time allocation by category
type EngineeringInvestment struct {
	ID           string    `json:"id"`
	TimePeriod   string    `json:"time_period"` // 'Q4 2024', '2024-W45'
	TeamID       *string   `json:"team_id,omitempty"`
	Category     string    `json:"category"` // 'customer_features', 'tech_debt', etc.
	StoryPoints  int       `json:"story_points"`
	Hours        float64   `json:"hours"`
	Cost         float64   `json:"cost"`
	FeatureCount int       `json:"feature_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CostEfficiencyMetric represents derived efficiency metrics
type CostEfficiencyMetric struct {
	ID          string    `json:"id"`
	TimePeriod  string    `json:"time_period"`
	EntityType  string    `json:"entity_type"` // 'team', 'org'
	EntityID    *string   `json:"entity_id,omitempty"`
	MetricName  string    `json:"metric_name"` // 'cost_per_pr', 'cost_per_point'
	MetricValue float64   `json:"metric_value"`
	Currency    string    `json:"currency"`
	ComputedAt  time.Time `json:"computed_at"`
}

// ROISummary provides computed ROI view (not stored, computed)
type ROISummary struct {
	TimePeriod          string  `json:"time_period"`
	TotalFeatures       int     `json:"total_features"`
	TotalInvestment     float64 `json:"total_investment"`
	TotalReturnValue    float64 `json:"total_return_value"`
	AverageROI          float64 `json:"average_roi"`
	PositiveROIFeatures int     `json:"positive_roi_features"`
	NegativeROIFeatures int     `json:"negative_roi_features"`
	Currency            string  `json:"currency"`
}

// FeatureROIDetails combines feature, cost, and ROI data
type FeatureROIDetails struct {
	Feature   *FeatureValue      `json:"feature"`
	Cost      *FeatureCost       `json:"cost,omitempty"`
	ROI       *ROICalculation    `json:"roi,omitempty"`
	WorkItems []*FeatureWorkItem `json:"work_items,omitempty"`
}

// InvestmentBreakdown shows time allocation across categories
type InvestmentBreakdown struct {
	TimePeriod string                   `json:"time_period"`
	TeamID     *string                  `json:"team_id,omitempty"`
	TotalHours float64                  `json:"total_hours"`
	TotalCost  float64                  `json:"total_cost"`
	Categories []*EngineeringInvestment `json:"categories"`
	Currency   string                   `json:"currency"`
}
