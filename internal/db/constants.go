package db

// DefaultQueryLimit is the default number of records returned by list operations
// when limit parameter is not specified.
const DefaultQueryLimit = 100

// MaxQueryLimit is the maximum number of records returned by list operations
// to prevent memory exhaustion and unbounded queries.
const MaxQueryLimit = 1000

// MaxEventQueryLimit is the higher limit specifically for event queries.
// Events can have larger result sets for analytics and reporting.
const MaxEventQueryLimit = 10000

// ReviewTimePercentOfCycleTime is the estimated percentage of cycle time
// spent in code review. Used as a proxy when detailed review time data
// is not available (V1 implementation).
const ReviewTimePercentOfCycleTime = 0.6
