package testutil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNormalizeTimestamps tests timestamp normalization with various formats
func TestNormalizeTimestamps(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ISO 8601 with Z timezone",
			input:    "Created at 2024-01-15T10:30:45Z",
			expected: "Created at TIMESTAMP",
		},
		{
			name:     "ISO 8601 with milliseconds",
			input:    "Updated at 2024-01-15T10:30:45.123Z",
			expected: "Updated at TIMESTAMP",
		},
		{
			name:     "RFC 3339 with positive offset",
			input:    "Timestamp: 2024-01-15T10:30:45+00:00",
			expected: "Timestamp: TIMESTAMP",
		},
		{
			name:     "RFC 3339 with negative offset",
			input:    "Time: 2024-01-15T10:30:45-05:00",
			expected: "Time: TIMESTAMP",
		},
		{
			name:     "Date only ISO format",
			input:    "Date: 2024-01-15",
			expected: "Date: DATE",
		},
		{
			name:     "US date format",
			input:    "Date: 01/15/2024",
			expected: "Date: DATE",
		},
		{
			name:     "US date format single digit",
			input:    "Date: 1/5/2024",
			expected: "Date: DATE",
		},
		{
			name:     "Time with milliseconds",
			input:    "Time: 10:30:45.123",
			expected: "Time: TIME",
		},
		{
			name:     "Time without milliseconds",
			input:    "Time: 10:30:45",
			expected: "Time: TIME",
		},
		{
			name:     "Unix timestamp",
			input:    "Timestamp: 1234567890",
			expected: "Timestamp: UNIX_TIMESTAMP",
		},
		{
			name:     "Multiple timestamps",
			input:    "From 2024-01-15T10:30:45Z to 2024-01-16T15:45:30Z",
			expected: "From TIMESTAMP to TIMESTAMP",
		},
		{
			name:     "Mixed formats",
			input:    "Date: 2024-01-15, Time: 10:30:45, Unix: 1234567890",
			expected: "Date: DATE, Time: TIME, Unix: UNIX_TIMESTAMP",
		},
		{
			name:     "No timestamps",
			input:    "This is just text without any timestamps",
			expected: "This is just text without any timestamps",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeTimestamps(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizeUUIDs tests UUID normalization
func TestNormalizeUUIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Standard UUID lowercase",
			input:    "User ID: 550e8400-e29b-41d4-a716-446655440000",
			expected: "User ID: UUID",
		},
		{
			name:     "Standard UUID uppercase",
			input:    "ID: 550E8400-E29B-41D4-A716-446655440000",
			expected: "ID: UUID",
		},
		{
			name:     "Mixed case UUID",
			input:    "GUID: 550e8400-E29B-41d4-A716-446655440000",
			expected: "GUID: UUID",
		},
		{
			name:     "Multiple UUIDs",
			input:    "From 550e8400-e29b-41d4-a716-446655440000 to 123e4567-e89b-12d3-a456-426614174000",
			expected: "From UUID to UUID",
		},
		{
			name:     "UUID in sentence",
			input:    "The transaction 550e8400-e29b-41d4-a716-446655440000 was successful",
			expected: "The transaction UUID was successful",
		},
		{
			name:     "No UUIDs",
			input:    "This is just text without any UUIDs",
			expected: "This is just text without any UUIDs",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeUUIDs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizeIDs tests ID normalization
func TestNormalizeIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Numeric ID with equals",
			input:    "Record id=12345",
			expected: "Record id=ID",
		},
		{
			name:     "Alphanumeric ID with equals",
			input:    "User id=abc123",
			expected: "User id=ID",
		},
		{
			name:     "ID with colon",
			input:    "Item id:789",
			expected: "Item id:ID",
		},
		{
			name:     "Multiple IDs",
			input:    "user_id=123 and post_id:456",
			expected: "user_id=ID and post_id:ID",
		},
		{
			name:     "ID with spaces",
			input:    "id = 999",
			expected: "id = ID", // Spaces are preserved in the capture group
		},
		{
			name:     "No IDs",
			input:    "This is just text",
			expected: "This is just text",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIDs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizer_Normalize tests custom pattern normalization
func TestNormalizer_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		patterns []NormalizePattern
		input    string
		expected string
	}{
		{
			name: "Single pattern - numbers",
			patterns: []NormalizePattern{
				{
					Pattern:     regexp.MustCompile(`\d+`),
					Replacement: "N",
				},
			},
			input:    "Count: 42, Total: 100",
			expected: "Count: N, Total: N",
		},
		{
			name: "Multiple patterns",
			patterns: []NormalizePattern{
				{
					Pattern:     regexp.MustCompile(`\b[A-Z]+\b`), // Only match whole uppercase words
					Replacement: "CAPS",
				},
				{
					Pattern:     regexp.MustCompile(`\d+`),
					Replacement: "NUM",
				},
			},
			input:    "User ABC has 123 points",
			expected: "User CAPS has NUM points",
		},
		{
			name: "Email normalization",
			patterns: []NormalizePattern{
				{
					Pattern:     regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
					Replacement: "EMAIL",
				},
			},
			input:    "Contact: user@example.com",
			expected: "Contact: EMAIL",
		},
		{
			name: "IP address normalization",
			patterns: []NormalizePattern{
				{
					Pattern:     regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
					Replacement: "IP",
				},
			},
			input:    "Server: 192.168.1.1",
			expected: "Server: IP",
		},
		{
			name: "URL normalization",
			patterns: []NormalizePattern{
				{
					Pattern:     regexp.MustCompile(`https?://[^\s]+`),
					Replacement: "URL",
				},
			},
			input:    "Visit https://example.com for more info",
			expected: "Visit URL for more info",
		},
		{
			name:     "Empty patterns",
			patterns: []NormalizePattern{},
			input:    "No changes",
			expected: "No changes",
		},
		{
			name: "No matches",
			patterns: []NormalizePattern{
				{
					Pattern:     regexp.MustCompile(`\d+`),
					Replacement: "N",
				},
			},
			input:    "No numbers here",
			expected: "No numbers here",
		},
		{
			name: "Empty input",
			patterns: []NormalizePattern{
				{
					Pattern:     regexp.MustCompile(`\d+`),
					Replacement: "N",
				},
			},
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewNormalizer(tt.patterns)
			result := normalizer.Normalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizeAll tests comprehensive normalization
func TestNormalizeAll(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "All types",
			input:    "Created at 2024-01-15T10:30:45Z with ID 550e8400-e29b-41d4-a716-446655440000 and user_id=123",
			expected: "Created at TIMESTAMP with ID UUID and user_id=ID",
		},
		{
			name:     "Timestamps and UUIDs",
			input:    "Event 2024-01-15T10:30:45Z: 550e8400-e29b-41d4-a716-446655440000",
			expected: "Event TIMESTAMP: UUID",
		},
		{
			name:     "Multiple of each type",
			input:    "From 2024-01-15 to 2024-01-16, users 550e8400-e29b-41d4-a716-446655440000 and 123e4567-e89b-12d3-a456-426614174000, ids id=1 and id=2",
			expected: "From DATE to DATE, users UUID and UUID, ids id=ID and id=ID",
		},
		{
			name:     "No dynamic content",
			input:    "This is static text",
			expected: "This is static text",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAll(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNewNormalizer tests normalizer creation
func TestNewNormalizer(t *testing.T) {
	tests := []struct {
		name     string
		patterns []NormalizePattern
	}{
		{
			name: "With patterns",
			patterns: []NormalizePattern{
				{
					Pattern:     regexp.MustCompile(`\d+`),
					Replacement: "N",
				},
			},
		},
		{
			name:     "Empty patterns",
			patterns: []NormalizePattern{},
		},
		{
			name:     "Nil patterns",
			patterns: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewNormalizer(tt.patterns)
			assert.NotNil(t, normalizer)
			assert.Equal(t, len(tt.patterns), len(normalizer.patterns))
		})
	}
}

// TestNormalizer_Performance tests normalization performance
func TestNormalizer_Performance(t *testing.T) {
	// Create a large input with many timestamps and UUIDs
	input := ""
	for i := 0; i < 1000; i++ {
		input += "Event 2024-01-15T10:30:45Z with ID 550e8400-e29b-41d4-a716-446655440000\n"
	}

	// This should complete quickly (< 100ms for 1000 lines)
	result := NormalizeAll(input)

	// Verify normalization worked
	assert.Contains(t, result, "TIMESTAMP")
	assert.Contains(t, result, "UUID")
	assert.NotContains(t, result, "2024-01-15")
	assert.NotContains(t, result, "550e8400")
}

// TestNormalizer_OrderMatters tests that pattern order matters
func TestNormalizer_OrderMatters(t *testing.T) {
	input := "Value: 123"

	// First pattern matches, second won't see original
	normalizer1 := NewNormalizer([]NormalizePattern{
		{Pattern: regexp.MustCompile(`\d+`), Replacement: "NUM"},
		{Pattern: regexp.MustCompile(`123`), Replacement: "SPECIFIC"},
	})
	result1 := normalizer1.Normalize(input)
	assert.Equal(t, "Value: NUM", result1)

	// Reverse order - specific pattern first
	normalizer2 := NewNormalizer([]NormalizePattern{
		{Pattern: regexp.MustCompile(`123`), Replacement: "SPECIFIC"},
		{Pattern: regexp.MustCompile(`\d+`), Replacement: "NUM"},
	})
	result2 := normalizer2.Normalize(input)
	assert.Equal(t, "Value: SPECIFIC", result2)
}

// TestNormalizer_ComplexPatterns tests complex regex patterns
func TestNormalizer_ComplexPatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  *regexp.Regexp
		input    string
		expected string
	}{
		{
			name:     "Word boundaries",
			pattern:  regexp.MustCompile(`\btest\b`),
			input:    "test testing tested",
			expected: "WORD testing tested",
		},
		{
			name:     "Character classes",
			pattern:  regexp.MustCompile(`[A-Z][a-z]+`),
			input:    "Hello World",
			expected: "WORD WORD",
		},
		{
			name:     "Anchors",
			pattern:  regexp.MustCompile(`(?m)^Start`), // Multiline mode
			input:    "Start of line\nStart again",
			expected: "WORD of line\nWORD again",
		},
		{
			name:     "Alternation",
			pattern:  regexp.MustCompile(`cat|dog`),
			input:    "I have a cat and a dog",
			expected: "I have a WORD and a WORD",
		},
		{
			name:     "Quantifiers",
			pattern:  regexp.MustCompile(`\b\w{3,5}\b`), // Word boundaries to match whole words only
			input:    "a ab abc abcd abcde abcdef",
			expected: "a ab WORD WORD WORD abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewNormalizer([]NormalizePattern{
				{Pattern: tt.pattern, Replacement: "WORD"},
			})
			result := normalizer.Normalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizer_SpecialCharacters tests normalization with special characters
func TestNormalizer_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Newlines preserved",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Tabs preserved",
			input:    "Col1\tCol2\tCol3",
			expected: "Col1\tCol2\tCol3",
		},
		{
			name:     "Unicode characters",
			input:    "Hello 世界 🌍",
			expected: "Hello 世界 🌍",
		},
		{
			name:     "Escape sequences",
			input:    "Path: C:\\Users\\test",
			expected: "Path: C:\\Users\\test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use empty normalizer to verify special chars are preserved
			normalizer := NewNormalizer([]NormalizePattern{})
			result := normalizer.Normalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
