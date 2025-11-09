package main

import (
	"testing"
)

func TestExtractStoryPoints(t *testing.T) {
	tests := []struct {
		name     string
		labels   []string
		expected int
		found    bool
	}{
		{
			name:     "story-points with hyphen and colon",
			labels:   []string{"story-points: 8", "bug", "priority: high"},
			expected: 8,
			found:    true,
		},
		{
			name:     "points with colon",
			labels:   []string{"bug", "points: 5"},
			expected: 5,
			found:    true,
		},
		{
			name:     "sp with colon",
			labels:   []string{"enhancement", "sp: 3"},
			expected: 3,
			found:    true,
		},
		{
			name:     "number before points",
			labels:   []string{"8 points", "feature"},
			expected: 8,
			found:    true,
		},
		{
			name:     "story points with space and colon",
			labels:   []string{"story points: 13"},
			expected: 13,
			found:    true,
		},
		{
			name:     "storypoints no space",
			labels:   []string{"storypoints: 5"},
			expected: 5,
			found:    true,
		},
		{
			name:     "pts abbreviation",
			labels:   []string{"pts: 2"},
			expected: 2,
			found:    true,
		},
		{
			name:     "no story points",
			labels:   []string{"bug", "enhancement", "wontfix"},
			expected: 0,
			found:    false,
		},
		{
			name:     "zero points should not match",
			labels:   []string{"points: 0"},
			expected: 0,
			found:    false,
		},
		{
			name:     "case insensitive",
			labels:   []string{"STORY-POINTS: 21"},
			expected: 21,
			found:    true,
		},
		{
			name:     "mixed case",
			labels:   []string{"Story Points: 8"},
			expected: 8,
			found:    true,
		},
		{
			name:     "multiple matches - first wins",
			labels:   []string{"sp: 3", "points: 5"},
			expected: 3,
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points, found := extractStoryPoints(tt.labels)
			if found != tt.found {
				t.Errorf("extractStoryPoints() found = %v, want %v", found, tt.found)
			}
			if points != tt.expected {
				t.Errorf("extractStoryPoints() points = %v, want %v", points, tt.expected)
			}
		})
	}
}
