package roi

import (
	"fmt"
	"sort"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Service handles ROI calculation business logic
type Service struct {
	store *db.CostROIStore
}

// NewService creates a new ROI service
func NewService(store *db.CostROIStore) *Service {
	return &Service{
		store: store,
	}
}

// CalculateFeatureROI computes ROI for a feature
func (s *Service) CalculateFeatureROI(featureID string) (*models.ROICalculation, error) {
	// Get feature
	feature, err := s.store.GetFeatureValue(featureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature: %w", err)
	}
	if feature == nil {
		return nil, fmt.Errorf("feature not found: %s", featureID)
	}

	// Get cost
	cost, err := s.store.GetFeatureCost(featureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature cost: %w", err)
	}
	if cost == nil {
		return nil, fmt.Errorf("feature cost not calculated: %s", featureID)
	}

	// Check if feature has value assigned
	if feature.ValueAmount == nil {
		return nil, fmt.Errorf("feature has no value assigned: %s", featureID)
	}

	investment := cost.TotalCost
	returnValue := *feature.ValueAmount

	// Calculate ROI percentage
	var roiPct float64
	if investment > 0 {
		roiPct = ((returnValue - investment) / investment) * 100.0
	}

	// Calculate payback period (months)
	// Assume value is ARR, so monthly value = ARR / 12
	var paybackMonths float64
	if returnValue > 0 {
		monthlyValue := returnValue / 12.0
		if monthlyValue > 0 {
			paybackMonths = investment / monthlyValue
		}
	}

	// Determine confidence
	confidence := "low"
	if feature.ConfidenceLevel == "actual" {
		confidence = "high"
	} else if feature.ConfidenceLevel == "validated" {
		confidence = "medium"
	}

	roi := &models.ROICalculation{
		FeatureID:     featureID,
		Investment:    investment,
		ReturnValue:   &returnValue,
		ROIPercentage: &roiPct,
		PaybackMonths: &paybackMonths,
		Confidence:    confidence,
		CalculatedAt:  time.Now().UTC(),
	}

	// Save to database
	if err := s.store.CreateROICalculation(roi); err != nil {
		return nil, fmt.Errorf("failed to save ROI calculation: %w", err)
	}

	return roi, nil
}

// CalculatePaybackPeriod calculates months to break even
func (s *Service) CalculatePaybackPeriod(investment, annualReturn float64) float64 {
	if annualReturn <= 0 {
		return 0
	}

	monthlyReturn := annualReturn / 12.0
	return investment / monthlyReturn
}

// GenerateROIReport creates portfolio view
func (s *Service) GenerateROIReport(timePeriod string) (*models.ROISummary, error) {
	// Get all features for period
	features, _, err := s.store.ListFeatureValues(timePeriod, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list features: %w", err)
	}

	summary := &models.ROISummary{
		TimePeriod:          timePeriod,
		TotalFeatures:       0,
		TotalInvestment:     0,
		TotalReturnValue:    0,
		AverageROI:          0,
		PositiveROIFeatures: 0,
		NegativeROIFeatures: 0,
		Currency:            "USD",
	}

	totalROI := 0.0
	roiCount := 0

	for _, feature := range features {
		summary.TotalFeatures++

		// Get ROI calculation
		roi, err := s.store.GetROICalculation(feature.ID)
		if err != nil || roi == nil {
			continue
		}

		summary.TotalInvestment += roi.Investment
		if roi.ReturnValue != nil {
			summary.TotalReturnValue += *roi.ReturnValue
		}

		if roi.ROIPercentage != nil {
			totalROI += *roi.ROIPercentage
			roiCount++

			if *roi.ROIPercentage > 0 {
				summary.PositiveROIFeatures++
			} else {
				summary.NegativeROIFeatures++
			}
		}
	}

	if roiCount > 0 {
		summary.AverageROI = totalROI / float64(roiCount)
	}

	return summary, nil
}

// CompareFeatureROI ranks features by ROI
func (s *Service) CompareFeatureROI(timePeriod string) ([]*models.FeatureROIDetails, error) {
	// Get all features
	features, _, err := s.store.ListFeatureValues(timePeriod, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list features: %w", err)
	}

	var details []*models.FeatureROIDetails

	for _, feature := range features {
		cost, _ := s.store.GetFeatureCost(feature.ID)
		roi, _ := s.store.GetROICalculation(feature.ID)
		workItems, _ := s.store.GetFeatureWorkItems(feature.ID, 100)

		detail := &models.FeatureROIDetails{
			Feature:   feature,
			Cost:      cost,
			ROI:       roi,
			WorkItems: workItems,
		}

		details = append(details, detail)
	}

	// Sort by ROI percentage (descending)
	sort.Slice(details, func(i, j int) bool {
		roiI := 0.0
		roiJ := 0.0

		if details[i].ROI != nil && details[i].ROI.ROIPercentage != nil {
			roiI = *details[i].ROI.ROIPercentage
		}
		if details[j].ROI != nil && details[j].ROI.ROIPercentage != nil {
			roiJ = *details[j].ROI.ROIPercentage
		}

		return roiI > roiJ
	})

	return details, nil
}

// PredictFutureROI forecasts based on trends
func (s *Service) PredictFutureROI(historicalPeriods []string) (float64, error) {
	if len(historicalPeriods) == 0 {
		return 0, fmt.Errorf("no historical periods provided")
	}

	totalROI := 0.0
	count := 0

	for _, period := range historicalPeriods {
		summary, err := s.GenerateROIReport(period)
		if err != nil {
			continue
		}

		totalROI += summary.AverageROI
		count++
	}

	if count == 0 {
		return 0, fmt.Errorf("no ROI data available")
	}

	// Simple average prediction
	return totalROI / float64(count), nil
}

// GetFeatureROIDetails retrieves complete ROI details for a feature
func (s *Service) GetFeatureROIDetails(featureID string) (*models.FeatureROIDetails, error) {
	feature, err := s.store.GetFeatureValue(featureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature: %w", err)
	}
	if feature == nil {
		return nil, fmt.Errorf("feature not found: %s", featureID)
	}

	cost, _ := s.store.GetFeatureCost(featureID)
	roi, _ := s.store.GetROICalculation(featureID)
	workItems, _ := s.store.GetFeatureWorkItems(featureID, 100)

	return &models.FeatureROIDetails{
		Feature:   feature,
		Cost:      cost,
		ROI:       roi,
		WorkItems: workItems,
	}, nil
}

// CalculateMarginalROI calculates incremental ROI of adding more features
func (s *Service) CalculateMarginalROI(baseInvestment, baseReturn, additionalInvestment, additionalReturn float64) float64 {
	baseROI := ((baseReturn - baseInvestment) / baseInvestment) * 100.0
	totalInvestment := baseInvestment + additionalInvestment
	totalReturn := baseReturn + additionalReturn
	newROI := ((totalReturn - totalInvestment) / totalInvestment) * 100.0

	return newROI - baseROI
}
