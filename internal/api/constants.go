package api

// MaxUploadSize is the maximum file size for CSV uploads (10MB).
const MaxUploadSize = 10 * 1024 * 1024

// MaxActivityDays is the maximum allowed value for days query parameters (1 year).
const MaxActivityDays = 365

// DefaultActivityDays is the default time period for engineer activity metrics (30 days).
const DefaultActivityDays = 30

// MaxImportRows is the maximum number of rows allowed in CSV import.
const MaxImportRows = 1000

// ValidIdentifierSources is the whitelist of allowed identifier sources
var ValidIdentifierSources = map[string]bool{
	"github": true,
	"jira":   true,
	"gitlab": true,
	"slack":  true,
	"zoom":   true,
	"csv":    true,
}
