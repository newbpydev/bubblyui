package testutil

import (
	"regexp"
	"strings"
)

// AssertRenderContains asserts that the component's rendered output contains
// the specified substring. The comparison is case-sensitive.
//
// If the assertion fails, it reports an error via t.Errorf with a message
// showing the substring and the actual rendered output.
//
// Parameters:
//   - substring: The substring to search for in the rendered output
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.AssertRenderContains("Hello")        // Passes if output contains "Hello"
//	ct.AssertRenderContains("Error:")       // Passes if output contains "Error:"
//	ct.AssertRenderContains("not-present")  // Fails if output doesn't contain "not-present"
func (ct *ComponentTest) AssertRenderContains(substring string) {
	ct.harness.t.Helper()

	// Get rendered output
	actual := ct.component.View()

	// Check if output contains substring
	if !strings.Contains(actual, substring) {
		ct.harness.t.Errorf("render output does not contain %q\nActual output:\n%s",
			substring, actual)
	}
}

// AssertRenderEquals asserts that the component's rendered output exactly matches
// the expected string. The comparison is case-sensitive and includes all whitespace.
//
// If the assertion fails, it reports an error via t.Errorf with a message
// showing both the expected and actual rendered output.
//
// Parameters:
//   - expected: The exact string that the rendered output should match
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.AssertRenderEquals("Counter: 0")           // Passes if output is exactly "Counter: 0"
//	ct.AssertRenderEquals("Line 1\nLine 2")       // Passes if output matches multiline string
//	ct.AssertRenderEquals("Different")            // Fails if output doesn't match exactly
func (ct *ComponentTest) AssertRenderEquals(expected string) {
	ct.harness.t.Helper()

	// Get rendered output
	actual := ct.component.View()

	// Compare strings exactly
	if actual != expected {
		ct.harness.t.Errorf("render output does not match expected\nExpected:\n%s\n\nActual:\n%s",
			expected, actual)
	}
}

// AssertRenderMatches asserts that the component's rendered output matches
// the specified regular expression pattern.
//
// The pattern should be a compiled *regexp.Regexp. The match is performed
// using the regexp's MatchString method, which returns true if the pattern
// matches any part of the output (unless anchors are used).
//
// If the assertion fails, it reports an error via t.Errorf with a message
// showing the pattern and the actual rendered output.
//
// Parameters:
//   - pattern: A compiled regular expression pattern to match against the output
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	pattern := regexp.MustCompile(`Counter: \d+`)
//	ct.AssertRenderMatches(pattern)  // Passes if output matches pattern
//
//	pattern2 := regexp.MustCompile(`^Error:`)
//	ct.AssertRenderMatches(pattern2) // Passes if output starts with "Error:"
func (ct *ComponentTest) AssertRenderMatches(pattern *regexp.Regexp) {
	ct.harness.t.Helper()

	// Get rendered output
	actual := ct.component.View()

	// Match pattern against output
	if !pattern.MatchString(actual) {
		ct.harness.t.Errorf("render output does not match pattern %q\nActual output:\n%s",
			pattern.String(), actual)
	}
}
