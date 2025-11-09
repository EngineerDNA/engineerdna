package forecasting

// Scenario default values
const (
	// DefaultVelocity is the default team velocity in story points per sprint
	DefaultVelocity = 30.0

	// DefaultCostPerEngineer is the default monthly fully-loaded cost per engineer in dollars
	DefaultCostPerEngineer = 15000.0

	// DefaultFeasibility is the default feasibility percentage for timeline scenarios
	DefaultFeasibility = 100.0

	// DefaultBudget is the default budget in dollars for cost scenarios
	DefaultBudget = 100000.0

	// DefaultRampWeeks is the default number of weeks for new engineers to ramp up
	DefaultRampWeeks = 6

	// MaxVelocityStretchFactor is the maximum stretch factor for team velocity (30% above baseline)
	MaxVelocityStretchFactor = 1.3
)

// Confidence level thresholds
const (
	// ConfidenceHigh represents high confidence level (90%)
	ConfidenceHigh = 90.0

	// ConfidenceMedium represents medium confidence level (85%)
	ConfidenceMedium = 85.0

	// ConfidenceLow represents low confidence level (60%)
	ConfidenceLow = 60.0

	// ConfidenceMinimum represents minimum confidence level (50%)
	ConfidenceMinimum = 50.0

	// ConfidenceDefault represents default confidence level (40%)
	ConfidenceDefault = 40.0

	// MaxProgressPercent is the maximum progress percentage (100%)
	MaxProgressPercent = 100.0
)

// Monte Carlo simulation constants
const (
	// DefaultMonteCarloSimulations is the default number of Monte Carlo simulations to run
	DefaultMonteCarloSimulations = 1000

	// ConfidenceLevel90 is the 90th percentile for confidence intervals
	ConfidenceLevel90 = 90

	// ConfidenceLevel80 is the 80th percentile for confidence intervals
	ConfidenceLevel80 = 80

	// ConfidenceLevel10 is the 10th percentile for confidence intervals
	ConfidenceLevel10 = 10
)

// Role-based productivity multipliers
const (
	// ProductivityJunior is the productivity multiplier for junior engineers
	ProductivityJunior = 0.6

	// ProductivityMid is the productivity multiplier for mid-level engineers
	ProductivityMid = 1.0

	// ProductivitySenior is the productivity multiplier for senior engineers
	ProductivitySenior = 1.3

	// ProductivityDuringRamp is the productivity multiplier during ramp-up period
	ProductivityDuringRamp = 0.5
)

// Capacity and efficiency constants
const (
	// DefaultTeamSize is the default team size for capacity calculations
	DefaultTeamSize = 5.0

	// EfficiencyFactor is the efficiency factor for budget-to-capacity conversion (80%)
	EfficiencyFactor = 0.8

	// SprintDurationDays is the standard sprint duration in days
	SprintDurationDays = 14

	// MinimumPace is the minimum pace to avoid division by zero in calculations
	MinimumPace = 0.1

	// MinimumProgressRate is the minimum progress rate to avoid division by zero
	MinimumProgressRate = 0.1

	// MinimumDaysElapsed is the minimum days elapsed to avoid division by zero
	MinimumDaysElapsed = 1.0
)

// Confidence interval constants
const (
	// ConfidenceIntervalLowMultiplier is the multiplier for lower confidence bound (80%)
	ConfidenceIntervalLowMultiplier = 0.8

	// ConfidenceIntervalHighMultiplier is the multiplier for upper confidence bound (120%)
	ConfidenceIntervalHighMultiplier = 1.2
)

// Historical data analysis constants
const (
	// DefaultHistoricalSprintCount is the default number of historical sprints to analyze
	DefaultHistoricalSprintCount = 5

	// MaxHistoricalSprintCount is the maximum number of historical sprints to analyze
	MaxHistoricalSprintCount = 10

	// DefaultVariance is the default variance for teams with insufficient historical data
	DefaultVariance = 0.5

	// ConfidenceStdDevFactor is the factor to convert std dev to confidence (5% per std dev)
	ConfidenceStdDevFactor = 5.0
)
