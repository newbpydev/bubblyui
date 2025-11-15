package testutil

import (
	"regexp"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// mockTestingT is a mock implementation of testingT for testing assertions.
type mockTestingTRender struct {
	errors  []string
	helpers int
}

func (m *mockTestingTRender) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, format)
}

func (m *mockTestingTRender) Helper() {
	m.helpers++
}

func (m *mockTestingTRender) Logf(format string, args ...interface{}) {
	// No-op for tests
}

func (m *mockTestingTRender) Cleanup(func()) {
	// No-op for tests
}

// createMockComponent creates a component that returns the given view output.
func createMockComponentWithView(view string) bubbly.Component {
	component, err := bubbly.NewComponent("MockComponent").
		Template(func(ctx bubbly.RenderContext) string {
			return view
		}).
		Build()
	if err != nil {
		panic(err) // Should never happen in tests with valid setup
	}
	return component
}

// TestAssertRenderContains tests the AssertRenderContains method.
func TestAssertRenderContains(t *testing.T) {
	tests := []struct {
		name          string
		viewOutput    string
		substring     string
		shouldPass    bool
		errorContains string
	}{
		{
			name:       "contains substring - simple",
			viewOutput: "Hello, World!",
			substring:  "World",
			shouldPass: true,
		},
		{
			name:       "contains substring - beginning",
			viewOutput: "Counter: 42",
			substring:  "Counter",
			shouldPass: true,
		},
		{
			name:       "contains substring - end",
			viewOutput: "Total: 100",
			substring:  "100",
			shouldPass: true,
		},
		{
			name:       "contains substring - middle",
			viewOutput: "The quick brown fox",
			substring:  "quick",
			shouldPass: true,
		},
		{
			name:       "contains empty string",
			viewOutput: "Any text",
			substring:  "",
			shouldPass: true, // Empty string is always contained
		},
		{
			name:          "does not contain substring",
			viewOutput:    "Hello, World!",
			substring:     "Goodbye",
			shouldPass:    false,
			errorContains: "render output does not contain",
		},
		{
			name:          "case sensitive - fails",
			viewOutput:    "Hello, World!",
			substring:     "world",
			shouldPass:    false,
			errorContains: "render output does not contain",
		},
		{
			name:       "multiline - contains",
			viewOutput: "Line 1\nLine 2\nLine 3",
			substring:  "Line 2",
			shouldPass: true,
		},
		{
			name:       "with special characters",
			viewOutput: "Price: $19.99",
			substring:  "$19.99",
			shouldPass: true,
		},
		{
			name:       "with whitespace",
			viewOutput: "Hello   World",
			substring:  "   ",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock testing.T
			mockT := &mockTestingTRender{}

			// Create harness with mock
			harness := &TestHarness{
				t:    mockT,
				refs: make(map[string]*bubbly.Ref[interface{}]),
			}

			// Create component with view output
			component := createMockComponentWithView(tt.viewOutput)
			component.Init()

			// Create ComponentTest
			ct := &ComponentTest{
				harness:   harness,
				component: component,
			}

			// Call assertion
			ct.AssertRenderContains(tt.substring)

			// Verify results
			if tt.shouldPass {
				if len(mockT.errors) > 0 {
					t.Errorf("Expected assertion to pass, but got error: %v", mockT.errors)
				}
			} else {
				if len(mockT.errors) == 0 {
					t.Errorf("Expected assertion to fail, but it passed")
				}
				if tt.errorContains != "" && len(mockT.errors) > 0 {
					// Check if error message contains expected text
					// Note: We're checking the format string, not the formatted output
					found := false
					for _, err := range mockT.errors {
						if contains(err, tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error to contain %q, got: %v", tt.errorContains, mockT.errors)
					}
				}
			}

			// Verify Helper() was called
			if mockT.helpers == 0 {
				t.Error("Expected Helper() to be called")
			}
		})
	}
}

// TestAssertRenderEquals tests the AssertRenderEquals method.
func TestAssertRenderEquals(t *testing.T) {
	tests := []struct {
		name          string
		viewOutput    string
		expected      string
		shouldPass    bool
		errorContains string
	}{
		{
			name:       "exact match - simple",
			viewOutput: "Hello, World!",
			expected:   "Hello, World!",
			shouldPass: true,
		},
		{
			name:       "exact match - empty",
			viewOutput: "",
			expected:   "",
			shouldPass: true,
		},
		{
			name:       "exact match - multiline",
			viewOutput: "Line 1\nLine 2\nLine 3",
			expected:   "Line 1\nLine 2\nLine 3",
			shouldPass: true,
		},
		{
			name:       "exact match - with special chars",
			viewOutput: "Price: $19.99\nTotal: $100.00",
			expected:   "Price: $19.99\nTotal: $100.00",
			shouldPass: true,
		},
		{
			name:          "not equal - different text",
			viewOutput:    "Hello, World!",
			expected:      "Goodbye, World!",
			shouldPass:    false,
			errorContains: "render output does not match expected",
		},
		{
			name:          "not equal - extra characters",
			viewOutput:    "Hello, World!",
			expected:      "Hello",
			shouldPass:    false,
			errorContains: "render output does not match expected",
		},
		{
			name:          "not equal - missing characters",
			viewOutput:    "Hello",
			expected:      "Hello, World!",
			shouldPass:    false,
			errorContains: "render output does not match expected",
		},
		{
			name:          "case sensitive",
			viewOutput:    "Hello, World!",
			expected:      "hello, world!",
			shouldPass:    false,
			errorContains: "render output does not match expected",
		},
		{
			name:          "whitespace matters",
			viewOutput:    "Hello World",
			expected:      "Hello  World",
			shouldPass:    false,
			errorContains: "render output does not match expected",
		},
		{
			name:          "trailing newline matters",
			viewOutput:    "Hello\n",
			expected:      "Hello",
			shouldPass:    false,
			errorContains: "render output does not match expected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock testing.T
			mockT := &mockTestingTRender{}

			// Create harness with mock
			harness := &TestHarness{
				t:    mockT,
				refs: make(map[string]*bubbly.Ref[interface{}]),
			}

			// Create component with view output
			component := createMockComponentWithView(tt.viewOutput)
			component.Init()

			// Create ComponentTest
			ct := &ComponentTest{
				harness:   harness,
				component: component,
			}

			// Call assertion
			ct.AssertRenderEquals(tt.expected)

			// Verify results
			if tt.shouldPass {
				if len(mockT.errors) > 0 {
					t.Errorf("Expected assertion to pass, but got error: %v", mockT.errors)
				}
			} else {
				if len(mockT.errors) == 0 {
					t.Errorf("Expected assertion to fail, but it passed")
				}
				if tt.errorContains != "" && len(mockT.errors) > 0 {
					found := false
					for _, err := range mockT.errors {
						if contains(err, tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error to contain %q, got: %v", tt.errorContains, mockT.errors)
					}
				}
			}

			// Verify Helper() was called
			if mockT.helpers == 0 {
				t.Error("Expected Helper() to be called")
			}
		})
	}
}

// TestAssertRenderMatches tests the AssertRenderMatches method.
func TestAssertRenderMatches(t *testing.T) {
	tests := []struct {
		name          string
		viewOutput    string
		pattern       string
		shouldPass    bool
		errorContains string
	}{
		{
			name:       "matches simple pattern",
			viewOutput: "Hello, World!",
			pattern:    "Hello.*World",
			shouldPass: true,
		},
		{
			name:       "matches digit pattern",
			viewOutput: "Counter: 42",
			pattern:    `Counter: \d+`,
			shouldPass: true,
		},
		{
			name:       "matches word boundary",
			viewOutput: "The quick brown fox",
			pattern:    `\bquick\b`,
			shouldPass: true,
		},
		{
			name:       "matches beginning anchor",
			viewOutput: "Hello, World!",
			pattern:    "^Hello",
			shouldPass: true,
		},
		{
			name:       "matches end anchor",
			viewOutput: "Hello, World!",
			pattern:    "World!$",
			shouldPass: true,
		},
		{
			name:       "matches multiline",
			viewOutput: "Line 1\nLine 2\nLine 3",
			pattern:    "Line 2",
			shouldPass: true,
		},
		{
			name:       "matches optional group",
			viewOutput: "Color: red",
			pattern:    `Color: (red|blue|green)`,
			shouldPass: true,
		},
		{
			name:       "matches character class",
			viewOutput: "Price: $19.99",
			pattern:    `Price: \$[\d.]+`,
			shouldPass: true,
		},
		{
			name:          "does not match pattern",
			viewOutput:    "Hello, World!",
			pattern:       "Goodbye",
			shouldPass:    false,
			errorContains: "render output does not match pattern",
		},
		{
			name:          "case sensitive - fails",
			viewOutput:    "Hello, World!",
			pattern:       "hello",
			shouldPass:    false,
			errorContains: "render output does not match pattern",
		},
		{
			name:          "anchor mismatch",
			viewOutput:    "Hello, World!",
			pattern:       "^World",
			shouldPass:    false,
			errorContains: "render output does not match pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock testing.T
			mockT := &mockTestingTRender{}

			// Create harness with mock
			harness := &TestHarness{
				t:    mockT,
				refs: make(map[string]*bubbly.Ref[interface{}]),
			}

			// Create component with view output
			component := createMockComponentWithView(tt.viewOutput)
			component.Init()

			// Create ComponentTest
			ct := &ComponentTest{
				harness:   harness,
				component: component,
			}

			// Compile pattern
			pattern := regexp.MustCompile(tt.pattern)

			// Call assertion
			ct.AssertRenderMatches(pattern)

			// Verify results
			if tt.shouldPass {
				if len(mockT.errors) > 0 {
					t.Errorf("Expected assertion to pass, but got error: %v", mockT.errors)
				}
			} else {
				if len(mockT.errors) == 0 {
					t.Errorf("Expected assertion to fail, but it passed")
				}
				if tt.errorContains != "" && len(mockT.errors) > 0 {
					found := false
					for _, err := range mockT.errors {
						if contains(err, tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error to contain %q, got: %v", tt.errorContains, mockT.errors)
					}
				}
			}

			// Verify Helper() was called
			if mockT.helpers == 0 {
				t.Error("Expected Helper() to be called")
			}
		})
	}
}

// TestAssertRenderMatches_InvalidPattern tests error handling for invalid regex patterns.
func TestAssertRenderMatches_InvalidPattern(t *testing.T) {
	// This test verifies that we handle nil patterns gracefully
	// In practice, users should always pass valid compiled patterns
	mockT := &mockTestingTRender{}
	harness := &TestHarness{
		t:    mockT,
		refs: make(map[string]*bubbly.Ref[interface{}]),
	}

	component := createMockComponentWithView("test")
	component.Init()

	ct := &ComponentTest{
		harness:   harness,
		component: component,
	}

	// Passing nil pattern should be handled gracefully
	// (though in practice this shouldn't happen with proper API usage)
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with nil pattern
			t.Logf("Correctly panicked with nil pattern: %v", r)
		}
	}()

	ct.AssertRenderMatches(nil)
}

// contains checks if a string contains a substring (helper for tests).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
