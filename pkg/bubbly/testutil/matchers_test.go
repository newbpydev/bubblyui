package testutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestBeEmpty tests the BeEmpty matcher
func TestBeEmpty(t *testing.T) {
	tests := []struct {
		name        string
		actual      interface{}
		shouldMatch bool
	}{
		// Slices
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1, 2, 3}, false},
		{"nil slice", []int(nil), true},

		// Maps
		{"empty map", map[string]int{}, true},
		{"non-empty map", map[string]int{"a": 1}, false},
		{"nil map", map[string]int(nil), true},

		// Strings
		{"empty string", "", true},
		{"non-empty string", "hello", false},

		// Arrays
		{"empty array", [0]int{}, true},
		{"non-empty array", [3]int{1, 2, 3}, false},

		// Channels
		{"nil channel", (chan int)(nil), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := BeEmpty()
			matched, err := matcher.Match(tt.actual)

			assert.NoError(t, err)
			assert.Equal(t, tt.shouldMatch, matched, "BeEmpty() match result")

			// Test failure message
			if !matched {
				msg := matcher.FailureMessage(tt.actual)
				assert.NotEmpty(t, msg, "FailureMessage should not be empty")
			}
		})
	}

	// Test error cases - unsupported types
	errorTests := []struct {
		name   string
		actual interface{}
	}{
		{"integer", 42},
		{"float", 3.14},
		{"boolean", true},
		{"function", func() {}},
		{"struct", struct{}{}},
	}

	for _, tt := range errorTests {
		t.Run(tt.name+" error", func(t *testing.T) {
			matcher := BeEmpty()
			matched, err := matcher.Match(tt.actual)

			assert.Error(t, err, "BeEmpty() should return error for unsupported type")
			assert.False(t, matched, "BeEmpty() should not match unsupported type")
			assert.Contains(t, err.Error(), "BeEmpty matcher expects", "Error message should be descriptive")
		})
	}
}

// TestHaveLength tests the HaveLength matcher
func TestHaveLength(t *testing.T) {
	tests := []struct {
		name        string
		actual      interface{}
		expected    int
		shouldMatch bool
	}{
		// Slices
		{"slice length 0", []int{}, 0, true},
		{"slice length 3", []int{1, 2, 3}, 3, true},
		{"slice wrong length", []int{1, 2}, 3, false},
		{"nil slice", []int(nil), 0, true},

		// Maps
		{"map length 0", map[string]int{}, 0, true},
		{"map length 2", map[string]int{"a": 1, "b": 2}, 2, true},
		{"map wrong length", map[string]int{"a": 1}, 2, false},
		{"nil map", map[string]int(nil), 0, true},

		// Strings
		{"string length 0", "", 0, true},
		{"string length 5", "hello", 5, true},
		{"string wrong length", "hi", 5, false},

		// Arrays
		{"array length 0", [0]int{}, 0, true},
		{"array length 3", [3]int{1, 2, 3}, 3, true},
		{"array wrong length", [2]int{1, 2}, 3, false},

		// Channels
		{"buffered channel", make(chan int, 5), 5, true},
		{"unbuffered channel", make(chan int), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := HaveLength(tt.expected)
			matched, err := matcher.Match(tt.actual)

			assert.NoError(t, err)
			assert.Equal(t, tt.shouldMatch, matched, "HaveLength() match result")

			// Test failure message
			if !matched {
				msg := matcher.FailureMessage(tt.actual)
				assert.NotEmpty(t, msg, "FailureMessage should not be empty")
			}
		})
	}
}

// TestHaveLength_InvalidTypes tests HaveLength with invalid types
func TestHaveLength_InvalidTypes(t *testing.T) {
	tests := []struct {
		name   string
		actual interface{}
	}{
		{"int", 42},
		{"bool", true},
		{"struct", struct{}{}},
		{"pointer", new(int)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := HaveLength(5)
			_, err := matcher.Match(tt.actual)

			assert.Error(t, err, "HaveLength should error for non-collection types")
		})
	}
}

// TestBeNil tests the BeNil matcher
func TestBeNil(t *testing.T) {
	tests := []struct {
		name        string
		actual      interface{}
		shouldMatch bool
	}{
		// Nil values
		{"nil interface", nil, true},
		{"nil pointer", (*int)(nil), true},
		{"nil slice", []int(nil), true},
		{"nil map", map[string]int(nil), true},
		{"nil channel", (chan int)(nil), true},
		{"nil func", (func())(nil), true},

		// Non-nil values
		{"non-nil int", 42, false},
		{"non-nil string", "hello", false},
		{"non-nil bool", true, false},
		{"non-nil pointer", new(int), false},
		{"non-nil slice", []int{}, false},
		{"non-nil map", map[string]int{}, false},
		{"non-nil channel", make(chan int), false},
		{"non-nil struct", struct{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := BeNil()
			matched, err := matcher.Match(tt.actual)

			assert.NoError(t, err)
			assert.Equal(t, tt.shouldMatch, matched, "BeNil() match result")

			// Test failure message
			if !matched {
				msg := matcher.FailureMessage(tt.actual)
				assert.NotEmpty(t, msg, "FailureMessage should not be empty")
			}
		})
	}
}

// TestAssertThat_Success tests AssertThat with passing matchers
func TestAssertThat_Success(t *testing.T) {
	harness := NewHarness(t)
	ct := &ComponentTest{harness: harness}

	tests := []struct {
		name    string
		actual  interface{}
		matcher Matcher
	}{
		{"empty slice", []int{}, BeEmpty()},
		{"length 3", []int{1, 2, 3}, HaveLength(3)},
		{"nil value", nil, BeNil()},
		{"empty string", "", BeEmpty()},
		{"string length", "hello", HaveLength(5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not fail
			ct.AssertThat(tt.actual, tt.matcher)
		})
	}
}

// TestAssertThat_Failure tests AssertThat with failing matchers
func TestAssertThat_Failure(t *testing.T) {
	tests := []struct {
		name    string
		actual  interface{}
		matcher Matcher
	}{
		{"non-empty slice", []int{1, 2, 3}, BeEmpty()},
		{"wrong length", []int{1, 2}, HaveLength(3)},
		{"non-nil value", 42, BeNil()},
		{"non-empty string", "hello", BeEmpty()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock testing.T to capture errors
			mockT := &mockTestingT{}
			harness := &TestHarness{t: mockT}
			ct := &ComponentTest{harness: harness}

			// Should fail
			ct.AssertThat(tt.actual, tt.matcher)

			// Verify error was reported
			assert.True(t, mockT.failed, "AssertThat should call t.Errorf for failing matcher")
			assert.NotEmpty(t, mockT.errors, "Error message should not be empty")
		})
	}
}

// TestMatcher_Composability tests that matchers can be composed
func TestMatcher_Composability(t *testing.T) {
	// Test that matchers work with AssertThat
	harness := NewHarness(t)
	ct := &ComponentTest{harness: harness}

	// Multiple assertions with different matchers
	ct.AssertThat([]int{}, BeEmpty())
	ct.AssertThat([]int{1, 2, 3}, HaveLength(3))
	ct.AssertThat(nil, BeNil())

	// All should pass without errors
}

// TestAssertThat_ErrorPath tests AssertThat with a matcher that returns an error
func TestAssertThat_ErrorPath(t *testing.T) {
	// Create a mock harness with mockTestingT to capture error calls
	mockT := &mockTestingT{}
	harness := &TestHarness{
		t:       mockT,
		refs:    make(map[string]*bubbly.Ref[interface{}]),
		events:  NewEventTracker(),
		cleanup: []func(){},
	}
	ct := &ComponentTest{harness: harness}

	// Create a matcher that always returns an error
	errorMatcher := &struct {
		Matcher
	}{}

	// Override Match method to return error
	errorMatcher.Matcher = MatcherFunc(func(actual interface{}) (bool, error) {
		return false, fmt.Errorf("test matcher error")
	})

	// This should call t.Errorf with the matcher error
	ct.AssertThat("test", errorMatcher)

	assert.True(t, mockT.failed, "AssertThat should fail when matcher returns error")
	assert.Contains(t, mockT.errors[0], "matcher error")
	assert.Contains(t, mockT.errors[0], "test matcher error")
}

// MatcherFunc is a function adapter for Matcher interface
type MatcherFunc func(actual interface{}) (bool, error)

func (f MatcherFunc) Match(actual interface{}) (bool, error) {
	return f(actual)
}

func (f MatcherFunc) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("matcher failed for %v", actual)
}

// TestBeEmpty_ErrorPath tests BeEmpty matcher with invalid types
func TestBeEmpty_ErrorPath(t *testing.T) {
	matcher := BeEmpty()

	// Test with invalid type (should return error)
	invalidTypes := []interface{}{
		42,         // int
		3.14,       // float64
		true,       // bool
		struct{}{}, // struct
	}

	for _, invalid := range invalidTypes {
		t.Run(fmt.Sprintf("invalid_type_%T", invalid), func(t *testing.T) {
			matched, err := matcher.Match(invalid)

			assert.Error(t, err, "should return error for invalid type")
			assert.False(t, matched, "should not match invalid type")
			assert.Contains(t, err.Error(), "BeEmpty matcher expects")
		})
	}
}

// TestHaveLength_ErrorPath tests HaveLength matcher with invalid types
func TestHaveLength_ErrorPath(t *testing.T) {
	matcher := HaveLength(3)

	// Test with invalid type (should return error)
	invalidTypes := []interface{}{
		42,         // int
		3.14,       // float64
		true,       // bool
		struct{}{}, // struct
	}

	for _, invalid := range invalidTypes {
		t.Run(fmt.Sprintf("invalid_type_%T", invalid), func(t *testing.T) {
			matched, err := matcher.Match(invalid)

			assert.Error(t, err, "should return error for invalid type")
			assert.False(t, matched, "should not match invalid type")
			assert.Contains(t, err.Error(), "HaveLength matcher expects")
		})
	}
}
