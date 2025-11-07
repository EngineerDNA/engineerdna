package db

// DefaultQueryLimit is the default number of records returned by list operations
// when limit parameter is not specified.
const DefaultQueryLimit = 100

// MaxQueryLimit is the maximum number of records returned by list operations
// to prevent memory exhaustion and unbounded queries.
const MaxQueryLimit = 1000
