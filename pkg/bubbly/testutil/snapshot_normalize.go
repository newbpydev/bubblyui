package testutil

import (
	"regexp"
)

// Normalizer applies a series of normalization patterns to content.
//
// This is useful for snapshot testing when content contains dynamic values
// like timestamps, UUIDs, or other non-deterministic data that should be
// normalized before comparison.
//
// Example:
//
//	normalizer := &Normalizer{
//	    patterns: []NormalizePattern{
//	        {Pattern: regexp.MustCompile(`\d{4}-\d{2}-\d{2}`), Replacement: "YYYY-MM-DD"},
//	        {Pattern: regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`), Replacement: "UUID"},
//	    },
//	}
//	normalized := normalizer.Normalize("Created at 2024-01-15 with ID abc123-...")
//	// Result: "Created at YYYY-MM-DD with ID UUID"
type Normalizer struct {
	patterns []NormalizePattern
}

// NormalizePattern defines a regular expression pattern and its replacement.
//
// When applied to content, all matches of the pattern will be replaced with
// the replacement string.
type NormalizePattern struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// NewNormalizer creates a new Normalizer with the specified patterns.
//
// Parameters:
//   - patterns: List of normalization patterns to apply
//
// Returns:
//   - *Normalizer: A new normalizer instance
//
// Example:
//
//	normalizer := NewNormalizer([]NormalizePattern{
//	    {Pattern: regexp.MustCompile(`\d+`), Replacement: "N"},
//	})
func NewNormalizer(patterns []NormalizePattern) *Normalizer {
	return &Normalizer{
		patterns: patterns,
	}
}

// Normalize applies all normalization patterns to the content.
//
// Patterns are applied in the order they were added to the normalizer.
// Each pattern's regex is used to find matches in the content, and all
// matches are replaced with the pattern's replacement string.
//
// Parameters:
//   - content: The content to normalize
//
// Returns:
//   - string: The normalized content with all patterns applied
//
// Example:
//
//	normalizer := NewNormalizer([]NormalizePattern{
//	    {Pattern: regexp.MustCompile(`\d+`), Replacement: "N"},
//	})
//	result := normalizer.Normalize("Count: 42, Total: 100")
//	// Result: "Count: N, Total: N"
func (n *Normalizer) Normalize(content string) string {
	result := content

	// Apply each pattern in sequence
	for _, pattern := range n.patterns {
		result = pattern.Pattern.ReplaceAllString(result, pattern.Replacement)
	}

	return result
}

// NormalizeTimestamps normalizes common timestamp formats in content.
//
// This function replaces various timestamp formats with placeholder strings
// to make snapshots deterministic. It handles:
//   - ISO 8601 timestamps (2024-01-15T10:30:45Z, 2024-01-15T10:30:45.123Z)
//   - RFC 3339 timestamps (2024-01-15T10:30:45+00:00)
//   - Date-only formats (2024-01-15, 01/15/2024)
//   - Time-only formats (10:30:45, 10:30:45.123)
//   - Unix timestamps (1234567890)
//
// Parameters:
//   - content: The content containing timestamps
//
// Returns:
//   - string: Content with timestamps replaced by "TIMESTAMP"
//
// Example:
//
//	input := "Created at 2024-01-15T10:30:45Z"
//	output := NormalizeTimestamps(input)
//	// Result: "Created at TIMESTAMP"
func NormalizeTimestamps(content string) string {
	patterns := []NormalizePattern{
		// ISO 8601 with timezone (2024-01-15T10:30:45Z, 2024-01-15T10:30:45.123Z)
		{
			Pattern:     regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z`),
			Replacement: "TIMESTAMP",
		},
		// RFC 3339 with timezone offset (2024-01-15T10:30:45+00:00, 2024-01-15T10:30:45-05:00)
		{
			Pattern:     regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?[+-]\d{2}:\d{2}`),
			Replacement: "TIMESTAMP",
		},
		// Date-only ISO format (2024-01-15)
		{
			Pattern:     regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
			Replacement: "DATE",
		},
		// US date format (01/15/2024, 1/15/2024)
		{
			Pattern:     regexp.MustCompile(`\d{1,2}/\d{1,2}/\d{4}`),
			Replacement: "DATE",
		},
		// Time with milliseconds (10:30:45.123)
		{
			Pattern:     regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d+`),
			Replacement: "TIME",
		},
		// Time without milliseconds (10:30:45)
		{
			Pattern:     regexp.MustCompile(`\d{2}:\d{2}:\d{2}`),
			Replacement: "TIME",
		},
		// Unix timestamp (10 digits, representing seconds since epoch)
		{
			Pattern:     regexp.MustCompile(`\b\d{10}\b`),
			Replacement: "UNIX_TIMESTAMP",
		},
	}

	normalizer := NewNormalizer(patterns)
	return normalizer.Normalize(content)
}

// NormalizeUUIDs normalizes UUID strings in content.
//
// This function replaces UUIDs (both standard and Microsoft formats) with
// a placeholder string to make snapshots deterministic. It handles:
//   - Standard UUIDs (8-4-4-4-12 format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
//   - Microsoft GUIDs (same format, case-insensitive)
//   - Uppercase and lowercase variants
//
// Parameters:
//   - content: The content containing UUIDs
//
// Returns:
//   - string: Content with UUIDs replaced by "UUID"
//
// Example:
//
//	input := "User ID: 550e8400-e29b-41d4-a716-446655440000"
//	output := NormalizeUUIDs(input)
//	// Result: "User ID: UUID"
func NormalizeUUIDs(content string) string {
	patterns := []NormalizePattern{
		// Standard UUID format (case-insensitive)
		// Matches: 550e8400-e29b-41d4-a716-446655440000
		{
			Pattern:     regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`),
			Replacement: "UUID",
		},
	}

	normalizer := NewNormalizer(patterns)
	return normalizer.Normalize(content)
}

// NormalizeIDs normalizes common ID formats in content.
//
// This function replaces various ID formats with placeholder strings.
// It handles:
//   - Numeric IDs (id=123, id:456)
//   - Alphanumeric IDs (id=abc123, id:xyz789)
//   - Prefixed IDs (user_id=123, post_id:456)
//
// Parameters:
//   - content: The content containing IDs
//
// Returns:
//   - string: Content with IDs replaced by "ID"
//
// Example:
//
//	input := "Record id=12345 and user_id:abc123"
//	output := NormalizeIDs(input)
//	// Result: "Record id=ID and user_id:ID"
func NormalizeIDs(content string) string {
	patterns := []NormalizePattern{
		// ID with equals sign (id=123, user_id=abc123)
		// Matches word ending with _id or just id, followed by = and alphanumeric value
		{
			Pattern:     regexp.MustCompile(`(\b\w*id\s*=\s*)[a-zA-Z0-9]+`),
			Replacement: "${1}ID",
		},
		// ID with colon (id:123, user_id:abc123)
		// Matches word ending with _id or just id, followed by : and alphanumeric value
		{
			Pattern:     regexp.MustCompile(`(\b\w*id\s*:\s*)[a-zA-Z0-9]+`),
			Replacement: "${1}ID",
		},
	}

	normalizer := NewNormalizer(patterns)
	return normalizer.Normalize(content)
}

// NormalizeAll applies all common normalizations to content.
//
// This is a convenience function that applies timestamp, UUID, and ID
// normalizations in sequence. Use this when you want comprehensive
// normalization of dynamic content.
//
// Parameters:
//   - content: The content to normalize
//
// Returns:
//   - string: Content with all normalizations applied
//
// Example:
//
//	input := "Created at 2024-01-15T10:30:45Z with ID 550e8400-e29b-41d4-a716-446655440000"
//	output := NormalizeAll(input)
//	// Result: "Created at TIMESTAMP with ID UUID"
func NormalizeAll(content string) string {
	content = NormalizeTimestamps(content)
	content = NormalizeUUIDs(content)
	content = NormalizeIDs(content)
	return content
}
