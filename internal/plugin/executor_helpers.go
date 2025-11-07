package plugin

import "time"

// Helper functions for type conversion

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func getTime(m map[string]interface{}, key string) time.Time {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			// Try to parse RFC3339 format
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t
			}
		}
		// If it's already a time.Time, return it
		if t, ok := v.(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}
