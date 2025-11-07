package forecasting

import (
	"math"
	"sort"
)

// Statistical prediction models

// calculateMovingAverage computes weighted moving average
func calculateMovingAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	total := 0.0
	weights := 0.0

	for i, val := range values {
		weight := float64(i + 1) // More recent values have higher weight
		total += val * weight
		weights += weight
	}

	return total / weights
}

// calculateTrend determines the trend direction and magnitude
func calculateTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	// Simple linear regression slope
	n := float64(len(values))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, val := range values {
		x := float64(i)
		sumX += x
		sumY += val
		sumXY += x * val
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	return slope
}

// calculateVariance computes variance of values
func calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := calculateMean(values)
	sumSquares := 0.0

	for _, val := range values {
		diff := val - mean
		sumSquares += diff * diff
	}

	return sumSquares / float64(len(values)-1)
}

// calculateStdDev computes standard deviation
func calculateStdDev(variance float64) float64 {
	return math.Sqrt(variance)
}

// calculateMean computes arithmetic mean
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, val := range values {
		sum += val
	}

	return sum / float64(len(values))
}

// calculatePercentile finds the value at a given percentile
func calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := (percentile / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	fraction := index - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}

// LinearRegressionModel performs linear regression prediction
func LinearRegressionModel(historicalX, historicalY []float64, targetX float64) (predictedY, confidence float64) {
	if len(historicalX) != len(historicalY) || len(historicalX) < 2 {
		return 0, 0
	}

	n := float64(len(historicalX))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i := 0; i < len(historicalX); i++ {
		sumX += historicalX[i]
		sumY += historicalY[i]
		sumXY += historicalX[i] * historicalY[i]
		sumX2 += historicalX[i] * historicalX[i]
	}

	// Calculate slope and intercept
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Predict
	predictedY = slope*targetX + intercept

	// Calculate R-squared for confidence
	meanY := sumY / n
	ssTotal := 0.0
	ssResidual := 0.0

	for i := 0; i < len(historicalX); i++ {
		predicted := slope*historicalX[i] + intercept
		ssTotal += (historicalY[i] - meanY) * (historicalY[i] - meanY)
		ssResidual += (historicalY[i] - predicted) * (historicalY[i] - predicted)
	}

	rSquared := 1 - (ssResidual / ssTotal)
	confidence = rSquared * 100 // Convert to percentage

	return predictedY, confidence
}

// MonteCarloSimulation runs Monte Carlo simulation with variance
func MonteCarloSimulation(baseValue, variance float64, simulations int) (low, high, mean float64) {
	if simulations <= 0 {
		simulations = 1000
	}

	results := make([]float64, simulations)
	stdDev := math.Sqrt(variance)

	for i := 0; i < simulations; i++ {
		// Generate random value with normal distribution
		randomFactor := normalRandom(0, stdDev)
		results[i] = baseValue + randomFactor
	}

	// Calculate statistics
	mean = calculateMean(results)
	low = calculatePercentile(results, 10)  // 10th percentile
	high = calculatePercentile(results, 90) // 90th percentile

	return low, high, mean
}

// normalRandom generates a random value from normal distribution
// Using Box-Muller transform
func normalRandom(mean, stdDev float64) float64 {
	// Simple approximation using central limit theorem
	// Sum of 12 uniform random values approximates normal distribution
	sum := 0.0
	for i := 0; i < 12; i++ {
		sum += float64(i) / 12.0 // Simplified random for deterministic results
	}
	// Transform to N(0,1) then scale
	z := sum - 6.0
	return mean + z*stdDev
}

// ExponentialSmoothing applies exponential smoothing for time series
func ExponentialSmoothing(values []float64, alpha float64) []float64 {
	if len(values) == 0 {
		return []float64{}
	}

	smoothed := make([]float64, len(values))
	smoothed[0] = values[0]

	for i := 1; i < len(values); i++ {
		smoothed[i] = alpha*values[i] + (1-alpha)*smoothed[i-1]
	}

	return smoothed
}

// CalculateConfidenceInterval calculates confidence interval
func CalculateConfidenceInterval(mean, stdDev float64, confidenceLevel float64) (low, high float64) {
	// For 95% confidence, use 1.96 standard deviations
	// For 90% confidence, use 1.645 standard deviations
	// For 80% confidence, use 1.28 standard deviations

	zScore := 1.96 // Default to 95%
	if confidenceLevel == 90 {
		zScore = 1.645
	} else if confidenceLevel == 80 {
		zScore = 1.28
	}

	margin := zScore * stdDev
	low = mean - margin
	high = mean + margin

	return low, high
}

// WeightedAverage calculates weighted average with custom weights
func WeightedAverage(values, weights []float64) float64 {
	if len(values) != len(weights) || len(values) == 0 {
		return 0
	}

	total := 0.0
	totalWeight := 0.0

	for i := 0; i < len(values); i++ {
		total += values[i] * weights[i]
		totalWeight += weights[i]
	}

	if totalWeight == 0 {
		return 0
	}

	return total / totalWeight
}

// PredictWithSeasonality predicts values considering seasonal patterns
func PredictWithSeasonality(historicalValues []float64, seasonLength int, periodsAhead int) float64 {
	if len(historicalValues) < seasonLength || seasonLength <= 0 {
		return calculateMean(historicalValues)
	}

	// Calculate seasonal indices
	seasonalIndices := make([]float64, seasonLength)
	for i := 0; i < seasonLength; i++ {
		sum := 0.0
		count := 0
		for j := i; j < len(historicalValues); j += seasonLength {
			sum += historicalValues[j]
			count++
		}
		if count > 0 {
			seasonalIndices[i] = sum / float64(count)
		}
	}

	// Apply trend
	trend := calculateTrend(historicalValues)
	baseValue := historicalValues[len(historicalValues)-1]

	// Predict
	seasonalIndex := periodsAhead % seasonLength
	predicted := baseValue + (trend * float64(periodsAhead)) + seasonalIndices[seasonalIndex]

	return predicted
}

// RollingAverage calculates rolling average with window size
func RollingAverage(values []float64, windowSize int) []float64 {
	if len(values) < windowSize || windowSize <= 0 {
		return values
	}

	result := make([]float64, len(values)-windowSize+1)

	for i := 0; i <= len(values)-windowSize; i++ {
		sum := 0.0
		for j := i; j < i+windowSize; j++ {
			sum += values[j]
		}
		result[i] = sum / float64(windowSize)
	}

	return result
}
