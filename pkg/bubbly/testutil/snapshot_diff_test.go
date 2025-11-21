package testutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGenerateDiff tests the generateDiff function with various scenarios
func TestGenerateDiff(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		wantDiff bool // true if diff should be non-empty
	}{
		{
			name:     "identical strings",
			expected: "hello world",
			actual:   "hello world",
			wantDiff: false,
		},
		{
			name:     "single line difference",
			expected: "hello world",
			actual:   "hello universe",
			wantDiff: true,
		},
		{
			name:     "multiline with additions",
			expected: "line 1\nline 2",
			actual:   "line 1\nline 2\nline 3",
			wantDiff: true,
		},
		{
			name:     "multiline with deletions",
			expected: "line 1\nline 2\nline 3",
			actual:   "line 1\nline 3",
			wantDiff: true,
		},
		{
			name:     "multiline with modifications",
			expected: "line 1\nline 2\nline 3",
			actual:   "line 1\nmodified line 2\nline 3",
			wantDiff: true,
		},
		{
			name:     "empty strings",
			expected: "",
			actual:   "",
			wantDiff: false,
		},
		{
			name:     "expected empty actual not",
			expected: "",
			actual:   "content",
			wantDiff: true,
		},
		{
			name:     "actual empty expected not",
			expected: "content",
			actual:   "",
			wantDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := generateDiff(tt.expected, tt.actual)

			if tt.wantDiff {
				assert.NotEmpty(t, diff, "Expected non-empty diff")
				// Verify diff contains expected and actual markers
				assert.Contains(t, diff, "---", "Diff should contain --- marker")
				assert.Contains(t, diff, "+++", "Diff should contain +++ marker")
			} else {
				assert.Empty(t, diff, "Expected empty diff for identical strings")
			}
		})
	}
}

// TestGenerateDiff_Format tests the format of generated diffs
func TestGenerateDiff_Format(t *testing.T) {
	expected := "line 1\nline 2\nline 3"
	actual := "line 1\nmodified\nline 3"

	diff := generateDiff(expected, actual)

	// Verify unified diff format
	assert.Contains(t, diff, "--- expected", "Should contain expected header")
	assert.Contains(t, diff, "+++ actual", "Should contain actual header")

	// Verify diff markers for changes
	lines := strings.Split(diff, "\n")
	var hasMinus, hasPlus bool
	for _, line := range lines {
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			hasMinus = true
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			hasPlus = true
		}
	}

	assert.True(t, hasMinus, "Should have deletion markers (-)")
	assert.True(t, hasPlus, "Should have addition markers (+)")
}

// TestHighlightDiff tests the highlightDiff function
func TestHighlightDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		wantAnsi bool // true if output should contain ANSI codes
	}{
		{
			name:     "empty diff",
			diff:     "",
			wantAnsi: false,
		},
		{
			name:     "diff with deletions",
			diff:     "--- expected\n+++ actual\n-deleted line",
			wantAnsi: true,
		},
		{
			name:     "diff with additions",
			diff:     "--- expected\n+++ actual\n+added line",
			wantAnsi: true,
		},
		{
			name:     "diff with context",
			diff:     "--- expected\n+++ actual\n context line\n-deleted\n+added",
			wantAnsi: true,
		},
		{
			name:     "diff with headers only",
			diff:     "--- expected\n+++ actual",
			wantAnsi: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			highlighted := highlightDiff(tt.diff)

			if tt.wantAnsi {
				// Check for ANSI escape codes (ESC character or \x1b)
				hasAnsi := strings.Contains(highlighted, "\x1b[") || strings.Contains(highlighted, "\033[")
				assert.True(t, hasAnsi, "Expected ANSI color codes in highlighted diff")
			}

			// Verify original content is preserved (minus ANSI codes)
			if tt.diff != "" {
				// The highlighted version should be longer due to ANSI codes
				if tt.wantAnsi {
					assert.GreaterOrEqual(t, len(highlighted), len(tt.diff),
						"Highlighted diff should be longer or equal due to ANSI codes")
				}
			}
		})
	}
}

// TestHighlightDiff_Colors tests that correct colors are applied
func TestHighlightDiff_Colors(t *testing.T) {
	diff := "--- expected\n+++ actual\n-deleted line\n+added line\n context line"
	highlighted := highlightDiff(diff)

	// Verify that highlighted output contains the original text
	assert.Contains(t, highlighted, "deleted line")
	assert.Contains(t, highlighted, "added line")
	assert.Contains(t, highlighted, "context line")

	// Verify ANSI codes are present
	assert.Contains(t, highlighted, "\x1b[", "Should contain ANSI escape sequences")
}

// TestFormatForTerminal tests the formatForTerminal function
func TestFormatForTerminal(t *testing.T) {
	tests := []struct {
		name string
		diff string
		want []string // strings that should be present in output
	}{
		{
			name: "simple diff",
			diff: "--- expected\n+++ actual\n-line 1\n+line 2",
			want: []string{"expected", "actual", "line 1", "line 2"},
		},
		{
			name: "multiline diff",
			diff: "--- expected\n+++ actual\n-old 1\n-old 2\n+new 1\n+new 2",
			want: []string{"old 1", "old 2", "new 1", "new 2"},
		},
		{
			name: "empty diff",
			diff: "",
			want: []string{},
		},
		{
			name: "diff with context",
			diff: "--- expected\n+++ actual\n context\n-deleted\n+added\n more context",
			want: []string{"context", "deleted", "added", "more context"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := formatForTerminal(tt.diff)

			for _, wantStr := range tt.want {
				assert.Contains(t, formatted, wantStr,
					"Formatted output should contain: %s", wantStr)
			}

			// Verify formatting adds structure
			if tt.diff != "" {
				// Should have some formatting applied (borders, padding, etc.)
				assert.GreaterOrEqual(t, len(formatted), len(tt.diff),
					"Formatted output should be longer or equal due to formatting")
			}
		})
	}
}

// TestFormatForTerminal_Readability tests that output is readable
func TestFormatForTerminal_Readability(t *testing.T) {
	diff := "--- expected\n+++ actual\n-deleted line\n+added line\n context line"
	formatted := formatForTerminal(diff)

	// Verify output is not empty
	assert.NotEmpty(t, formatted, "Formatted output should not be empty")

	// Verify it contains newlines for readability
	assert.Contains(t, formatted, "\n", "Should contain newlines for readability")

	// Verify original content is preserved
	assert.Contains(t, formatted, "deleted line")
	assert.Contains(t, formatted, "added line")
	assert.Contains(t, formatted, "context line")
}

// TestLargeDiff tests handling of large diffs
func TestLargeDiff(t *testing.T) {
	// Create large strings
	var expectedBuilder, actualBuilder strings.Builder
	for i := 0; i < 100; i++ {
		expectedBuilder.WriteString("line ")
		expectedBuilder.WriteString(string(rune('A' + (i % 26))))
		expectedBuilder.WriteString("\n")

		actualBuilder.WriteString("line ")
		actualBuilder.WriteString(string(rune('a' + (i % 26))))
		actualBuilder.WriteString("\n")
	}

	expected := expectedBuilder.String()
	actual := actualBuilder.String()

	// Generate diff
	diff := generateDiff(expected, actual)
	assert.NotEmpty(t, diff, "Large diff should not be empty")

	// Highlight diff
	highlighted := highlightDiff(diff)
	assert.NotEmpty(t, highlighted, "Highlighted large diff should not be empty")

	// Format for terminal
	formatted := formatForTerminal(diff)
	assert.NotEmpty(t, formatted, "Formatted large diff should not be empty")

	// Verify performance is acceptable (should complete quickly)
	// If this test times out, there's a performance issue
}

// TestDiffIntegration tests the full pipeline
func TestDiffIntegration(t *testing.T) {
	expected := "Hello, World!\nThis is line 2\nThis is line 3"
	actual := "Hello, Universe!\nThis is line 2\nThis is line 4"

	// Generate diff
	diff := generateDiff(expected, actual)
	assert.NotEmpty(t, diff, "Diff should not be empty")

	// Highlight diff
	highlighted := highlightDiff(diff)
	assert.NotEmpty(t, highlighted, "Highlighted diff should not be empty")
	assert.Contains(t, highlighted, "\x1b[", "Should contain ANSI codes")

	// Format for terminal
	formatted := formatForTerminal(diff)
	assert.NotEmpty(t, formatted, "Formatted diff should not be empty")

	// Verify all stages preserve content
	assert.Contains(t, diff, "World")
	assert.Contains(t, diff, "Universe")
	assert.Contains(t, highlighted, "World")
	assert.Contains(t, highlighted, "Universe")
	assert.Contains(t, formatted, "World")
	assert.Contains(t, formatted, "Universe")
}

// TestEdgeCases tests edge cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
	}{
		{
			name:     "very long lines",
			expected: strings.Repeat("a", 1000),
			actual:   strings.Repeat("b", 1000),
		},
		{
			name:     "unicode characters",
			expected: "Hello 世界\nこんにちは",
			actual:   "Hello World\nHello",
		},
		{
			name:     "special characters",
			expected: "line with\ttabs\nand\r\ncarriage returns",
			actual:   "line with  spaces\nand newlines",
		},
		{
			name:     "trailing whitespace",
			expected: "line 1  \nline 2\t",
			actual:   "line 1\nline 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic or error
			diff := generateDiff(tt.expected, tt.actual)
			highlighted := highlightDiff(diff)
			formatted := formatForTerminal(diff)

			// Basic sanity checks
			assert.NotNil(t, diff)
			assert.NotNil(t, highlighted)
			assert.NotNil(t, formatted)
		})
	}
}
