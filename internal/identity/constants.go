package identity

// Match confidence weights and thresholds for identity resolution algorithm

const (
	// MinMatchConfidence is the minimum confidence score (0.0-1.0) required
	// to suggest a match between an unresolved identity and an engineer.
	MinMatchConfidence = 0.5

	// EmailDomainMatchWeight is the confidence boost for matching email domains.
	EmailDomainMatchWeight = 0.3

	// EmailUsernameMatchWeight is the confidence boost for similar email usernames
	// (Levenshtein distance <= EmailUsernameSimilarityThreshold).
	EmailUsernameMatchWeight = 0.5

	// NameSimilarityWeight is the confidence boost for similar names
	// (Levenshtein distance <= LevenshteinThreshold).
	NameSimilarityWeight = 0.4

	// UsernamePatternWeight is the confidence boost for matching common username
	// patterns (e.g., "asmith" derived from "Alice Smith").
	UsernamePatternWeight = 0.6

	// LevenshteinThreshold is the maximum edit distance for name similarity matching.
	LevenshteinThreshold = 5

	// EmailUsernameSimilarityThreshold is the maximum edit distance for
	// email username similarity matching.
	EmailUsernameSimilarityThreshold = 2

	// MaxNameLength is the maximum allowed length for engineer names.
	MaxNameLength = 255

	// MaxEmailLength is the maximum allowed length for email addresses.
	MaxEmailLength = 255

	// MaxManagerLength is the maximum allowed length for manager names.
	MaxManagerLength = 255

	// MaxIdentifiers is the maximum number of identifiers allowed per engineer.
	MaxIdentifiers = 50

	// MaxValidationLength is the maximum string length for Levenshtein distance validation
	// to prevent memory exhaustion DoS attacks
	MaxValidationLength = 100
)
