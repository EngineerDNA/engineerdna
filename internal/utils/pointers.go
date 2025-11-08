package utils

import "time"

// PtrFloat64 returns a pointer to a float64 value
func PtrFloat64(f float64) *float64 {
	return &f
}

// PtrString returns a pointer to a string value
func PtrString(s string) *string {
	return &s
}

// PtrInt returns a pointer to an int value
func PtrInt(i int) *int {
	return &i
}

// PtrTime returns a pointer to a time.Time value
func PtrTime(t time.Time) *time.Time {
	return &t
}
