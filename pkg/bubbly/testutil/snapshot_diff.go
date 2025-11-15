package testutil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pmezard/go-difflib/difflib"
)

// generateDiff generates a unified diff between expected and actual content.
// Returns an empty string if the strings are identical.
//
// The diff uses unified diff format with:
//   - "--- expected" header for the expected content
//   - "+++ actual" header for the actual content
//   - "-" prefix for lines removed from expected
//   - "+" prefix for lines added in actual
//   - " " prefix for context lines (unchanged)
//
// Example:
//
//	expected := "line 1\nline 2\nline 3"
//	actual := "line 1\nmodified\nline 3"
//	diff := generateDiff(expected, actual)
//	// Returns:
//	// --- expected
//	// +++ actual
//	// @@ -1,3 +1,3 @@
//	//  line 1
//	// -line 2
//	// +modified
//	//  line 3
func generateDiff(expected, actual string) string {
	// If identical, return empty string
	if expected == actual {
		return ""
	}

	// Use difflib to generate unified diff
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		B:        difflib.SplitLines(actual),
		FromFile: "expected",
		ToFile:   "actual",
		Context:  3, // Show 3 lines of context around changes
	}

	result, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		// Fallback to simple diff if unified diff fails
		return fmt.Sprintf("--- expected\n%s\n\n+++ actual\n%s", expected, actual)
	}

	return result
}

// ANSI color codes for terminal output
const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[38;5;196m" // Bright red for deletions
	ansiGreen  = "\x1b[38;5;46m"  // Bright green for additions
	ansiCyan   = "\x1b[38;5;51m"  // Bright cyan for headers
	ansiYellow = "\x1b[38;5;226m" // Bright yellow for hunks
	ansiGray   = "\x1b[38;5;250m" // Light gray for context
)

// highlightDiff applies terminal color highlighting to a diff string.
// Uses ANSI color codes:
//   - Red for deletion lines (starting with -)
//   - Green for addition lines (starting with +)
//   - Cyan for header lines (starting with --- or +++)
//   - Yellow for hunk headers (starting with @@)
//   - Gray for context lines
//
// Example:
//
//	diff := "--- expected\n+++ actual\n-old line\n+new line"
//	highlighted := highlightDiff(diff)
//	// Returns the same content with ANSI color codes applied
func highlightDiff(diff string) string {
	if diff == "" {
		return ""
	}

	lines := strings.Split(diff, "\n")
	var highlighted strings.Builder

	for i, line := range lines {
		if line == "" {
			highlighted.WriteString("\n")
			continue
		}

		// Apply ANSI color codes based on line prefix
		var colorCode string
		switch {
		case strings.HasPrefix(line, "---"):
			colorCode = ansiCyan
		case strings.HasPrefix(line, "+++"):
			colorCode = ansiCyan
		case strings.HasPrefix(line, "@@"):
			colorCode = ansiYellow
		case strings.HasPrefix(line, "-"):
			colorCode = ansiRed
		case strings.HasPrefix(line, "+"):
			colorCode = ansiGreen
		default:
			// Context line (starts with space or no prefix)
			colorCode = ansiGray
		}

		highlighted.WriteString(colorCode)
		highlighted.WriteString(line)
		highlighted.WriteString(ansiReset)

		if i < len(lines)-1 {
			highlighted.WriteString("\n")
		}
	}

	return highlighted.String()
}

// formatForTerminal formats a diff for optimal terminal display.
// Adds visual structure with borders, padding, and clear sections.
//
// The formatted output includes:
//   - A bordered box around the diff
//   - Clear section headers
//   - Proper padding and spacing
//   - Line wrapping for long lines (if needed)
//
// Example:
//
//	diff := "--- expected\n+++ actual\n-old\n+new"
//	formatted := formatForTerminal(diff)
//	// Returns a nicely formatted, bordered diff suitable for terminal display
func formatForTerminal(diff string) string {
	if diff == "" {
		return ""
	}

	// First, highlight the diff
	highlighted := highlightDiff(diff)

	// Define border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")). // Dark gray
		Padding(1, 2)

	// Add a title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")). // White
		Padding(0, 1)

	title := titleStyle.Render("Snapshot Diff")

	// Combine title and diff
	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n\n")
	content.WriteString(highlighted)

	// Apply border
	formatted := borderStyle.Render(content.String())

	return formatted
}
