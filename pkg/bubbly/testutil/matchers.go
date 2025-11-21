package testutil

import (
	"fmt"
	"reflect"
)

// Matcher is the interface that custom assertion matchers must implement.
// It provides a flexible way to create reusable, composable assertions.
//
// Inspired by Gomega's GomegaMatcher interface, this allows for rich,
// expressive assertions beyond simple equality checks.
//
// Example:
//
//	matcher := BeEmpty()
//	matched, err := matcher.Match([]int{})
//	if err != nil {
//	    // Handle error
//	}
//	if !matched {
//	    msg := matcher.FailureMessage([]int{})
//	    t.Errorf(msg)
//	}
type Matcher interface {
	// Match returns true if the actual value matches the expectation.
	// It returns an error if the matcher cannot be applied to the actual value
	// (e.g., wrong type).
	Match(actual interface{}) (success bool, err error)

	// FailureMessage returns a human-readable message explaining why the match failed.
	// This is called when Match returns false to provide clear test output.
	FailureMessage(actual interface{}) string
}

// AssertThat asserts that the actual value matches the given matcher.
// This is the primary method for using custom matchers in tests.
//
// If the matcher returns an error (e.g., wrong type), it reports the error.
// If the match fails, it reports the matcher's failure message.
//
// Parameters:
//   - actual: The value to test
//   - matcher: The Matcher to apply
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.AssertThat([]int{}, BeEmpty())
//	ct.AssertThat([]int{1, 2, 3}, HaveLength(3))
//	ct.AssertThat(nil, BeNil())
func (ct *ComponentTest) AssertThat(actual interface{}, matcher Matcher) {
	ct.harness.t.Helper()

	// Apply matcher
	matched, err := matcher.Match(actual)
	if err != nil {
		ct.harness.t.Errorf("matcher error: %v", err)
		return
	}

	// Check if match succeeded
	if !matched {
		msg := matcher.FailureMessage(actual)
		ct.harness.t.Errorf("%s", msg)
	}
}

// emptyMatcher matches empty collections (slices, maps, strings, arrays, channels).
type emptyMatcher struct{}

// BeEmpty returns a matcher that succeeds if the actual value is empty.
// It works with slices, maps, strings, arrays, and channels.
//
// For slices, maps, and channels, nil values are considered empty.
// For strings, the empty string "" is considered empty.
// For arrays, only zero-length arrays are considered empty.
//
// Returns an error if the actual value is not a collection type.
//
// Example:
//
//	ct.AssertThat([]int{}, BeEmpty())           // passes
//	ct.AssertThat([]int{1, 2, 3}, BeEmpty())    // fails
//	ct.AssertThat("", BeEmpty())                // passes
//	ct.AssertThat("hello", BeEmpty())           // fails
//	ct.AssertThat(map[string]int{}, BeEmpty())  // passes
func BeEmpty() Matcher {
	return &emptyMatcher{}
}

func (m *emptyMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return true, nil
	}

	val := reflect.ValueOf(actual)
	kind := val.Kind()

	switch kind {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Array, reflect.Chan:
		return val.Len() == 0, nil
	default:
		return false, fmt.Errorf("BeEmpty matcher expects a slice, map, string, array, or channel, got %T", actual)
	}
}

func (m *emptyMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto be empty", actual)
}

// lengthMatcher matches collections with a specific length.
type lengthMatcher struct {
	expected int
}

// HaveLength returns a matcher that succeeds if the actual value has the expected length.
// It works with slices, maps, strings, arrays, and channels.
//
// For slices, maps, and channels, nil values have length 0.
// For strings, it checks the number of bytes (not runes).
// For channels, it checks the buffer capacity.
//
// Returns an error if the actual value is not a collection type.
//
// Parameters:
//   - expected: The expected length
//
// Example:
//
//	ct.AssertThat([]int{1, 2, 3}, HaveLength(3))     // passes
//	ct.AssertThat([]int{1, 2}, HaveLength(3))        // fails
//	ct.AssertThat("hello", HaveLength(5))            // passes
//	ct.AssertThat(map[string]int{"a": 1}, HaveLength(1)) // passes
func HaveLength(expected int) Matcher {
	return &lengthMatcher{expected: expected}
}

func (m *lengthMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return m.expected == 0, nil
	}

	val := reflect.ValueOf(actual)
	kind := val.Kind()

	switch kind {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Array:
		return val.Len() == m.expected, nil
	case reflect.Chan:
		// For channels, check capacity (not current length)
		return val.Cap() == m.expected, nil
	default:
		return false, fmt.Errorf("HaveLength matcher expects a slice, map, string, array, or channel, got %T", actual)
	}
}

func (m *lengthMatcher) FailureMessage(actual interface{}) string {
	actualLen := 0
	if actual != nil {
		val := reflect.ValueOf(actual)
		kind := val.Kind()
		switch kind {
		case reflect.Slice, reflect.Map, reflect.String, reflect.Array:
			actualLen = val.Len()
		case reflect.Chan:
			actualLen = val.Cap()
		}
	}
	return fmt.Sprintf("Expected\n\t%#v\nto have length %d, but has length %d", actual, m.expected, actualLen)
}

// nilMatcher matches nil values.
type nilMatcher struct{}

// BeNil returns a matcher that succeeds if the actual value is nil.
// It works with any type that can be nil: pointers, slices, maps, channels,
// functions, and interfaces.
//
// For non-nillable types (int, string, bool, struct), it always returns false.
//
// Example:
//
//	ct.AssertThat(nil, BeNil())                  // passes
//	ct.AssertThat((*int)(nil), BeNil())          // passes
//	ct.AssertThat([]int(nil), BeNil())           // passes
//	ct.AssertThat(42, BeNil())                   // fails
//	ct.AssertThat("hello", BeNil())              // fails
func BeNil() Matcher {
	return &nilMatcher{}
}

func (m *nilMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return true, nil
	}

	// Check if the value is a nil pointer, slice, map, channel, or function
	val := reflect.ValueOf(actual)
	kind := val.Kind()

	switch kind {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return val.IsNil(), nil
	default:
		// Non-nillable types are never nil
		return false, nil
	}
}

func (m *nilMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto be nil", actual)
}
