package testutil

import (
	"reflect"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CustomComparatorTester provides utilities for testing custom equality comparators
// in change detection. It allows you to verify that custom comparators are used correctly
// and track comparison invocations for performance testing.
//
// Custom comparators enable logical equality checks instead of identity checks,
// which is useful for:
//   - Comparing structs by specific fields (e.g., ID only)
//   - Ignoring large fields for performance
//   - Implementing domain-specific equality logic
//   - Testing array/slice equality by length or content
//
// Key Features:
//   - Track comparison function invocations
//   - Verify custom comparator is used (not default equality)
//   - Test logical equality vs identity
//   - Performance optimization verification
//
// Example:
//
//	type User struct {
//	    ID   int
//	    Name string
//	}
//
//	// Custom comparator that only compares ID
//	compareByID := func(a, b interface{}) bool {
//	    userA, okA := a.(User)
//	    userB, okB := b.(User)
//	    if !okA || !okB {
//	        return false
//	    }
//	    return userA.ID == userB.ID
//	}
//
//	ref := bubbly.NewRef[interface{}](User{ID: 1, Name: "Alice"})
//	tester := NewCustomComparatorTester(ref, compareByID)
//
//	// Change name but keep ID same
//	tester.SetValue(User{ID: 1, Name: "Bob"})
//
//	// Should not have changed according to custom comparator
//	tester.AssertChanged(t, false)
//	tester.AssertComparisons(t, 1)
//
// Thread Safety:
//
// CustomComparatorTester is not thread-safe. It should only be used from a single test goroutine.
type CustomComparatorTester struct {
	ref        *bubbly.Ref[interface{}]    // The ref being tested
	comparator func(a, b interface{}) bool // Custom comparator function
	compared   int                         // Number of comparisons performed
	changed    bool                        // Whether last comparison detected change
}

// NewCustomComparatorTester creates a new CustomComparatorTester for testing custom equality comparators.
//
// The tester wraps a Ref and a custom comparator function, tracking how many times
// the comparator is invoked and whether changes are detected.
//
// Parameters:
//   - ref: The Ref to test (must be *Ref[interface{}])
//   - comparator: Custom comparison function that returns true if values are equal
//
// Returns:
//   - *CustomComparatorTester: A new tester instance
//
// Example:
//
//	compareByID := func(a, b interface{}) bool {
//	    userA, okA := a.(User)
//	    userB, okB := b.(User)
//	    if !okA || !okB {
//	        return false
//	    }
//	    return userA.ID == userB.ID
//	}
//
//	ref := bubbly.NewRef[interface{}](User{ID: 1, Name: "Alice"})
//	tester := NewCustomComparatorTester(ref, compareByID)
func NewCustomComparatorTester(ref *bubbly.Ref[interface{}], comparator func(a, b interface{}) bool) *CustomComparatorTester {
	return &CustomComparatorTester{
		ref:        ref,
		comparator: comparator,
		compared:   0,
		changed:    false,
	}
}

// SetValue sets a new value on the ref and tracks whether the comparator detected a change.
//
// This method:
//  1. Gets the current value from the ref
//  2. Invokes the custom comparator to check equality
//  3. Increments the comparison counter
//  4. Updates the changed flag based on comparator result
//  5. Sets the new value on the ref
//
// The changed flag is set to true if the comparator returns false (values are different),
// and false if the comparator returns true (values are equal).
//
// Parameters:
//   - value: The new value to set
//
// Example:
//
//	tester.SetValue(User{ID: 1, Name: "Bob"})
//	tester.AssertChanged(t, false) // Same ID, different name
func (cct *CustomComparatorTester) SetValue(value interface{}) {
	// Get current value
	currentValue := cct.ref.Get()

	// Invoke comparator and track
	cct.compared++
	areEqual := cct.comparator(currentValue, value)
	cct.changed = !areEqual // Changed if not equal

	// Set new value
	cct.ref.Set(value)
}

// AssertComparisons asserts that the comparator was invoked the expected number of times.
//
// This is useful for verifying that:
//   - The custom comparator is actually being used
//   - Comparisons happen at the right times
//   - Performance optimizations are working (fewer comparisons)
//
// Parameters:
//   - t: The testing.T instance
//   - expected: The expected number of comparisons
//
// Example:
//
//	tester.SetValue(value1)
//	tester.SetValue(value2)
//	tester.AssertComparisons(t, 2) // Two SetValue calls = 2 comparisons
func (cct *CustomComparatorTester) AssertComparisons(t testing.TB, expected int) {
	t.Helper()
	if cct.compared != expected {
		t.Errorf("expected %d comparisons, but got %d", expected, cct.compared)
	}
}

// AssertChanged asserts whether the last SetValue call detected a change.
//
// This verifies that the custom comparator's logic is working correctly:
//   - Returns false when values are logically equal (no change)
//   - Returns true when values are different (change detected)
//
// Note: This reflects the result of the LAST SetValue call only.
// Each SetValue resets the changed flag based on that comparison.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: Whether a change should have been detected
//
// Example:
//
//	tester.SetValue(User{ID: 1, Name: "Alice"})
//	tester.AssertChanged(t, false) // Same ID = no change
//
//	tester.SetValue(User{ID: 2, Name: "Alice"})
//	tester.AssertChanged(t, true) // Different ID = change
func (cct *CustomComparatorTester) AssertChanged(t testing.TB, expected bool) {
	t.Helper()
	if cct.changed != expected {
		t.Errorf("expected changed=%v, but got changed=%v", expected, cct.changed)
	}
}

// GetComparisonCount returns the total number of comparisons performed.
//
// Returns:
//   - int: The number of times the comparator was invoked
//
// Example:
//
//	count := tester.GetComparisonCount()
//	assert.Equal(t, 3, count)
func (cct *CustomComparatorTester) GetComparisonCount() int {
	return cct.compared
}

// WasChanged returns whether the last comparison detected a change.
//
// Returns:
//   - bool: True if the last SetValue detected a change
//
// Example:
//
//	tester.SetValue(newValue)
//	if tester.WasChanged() {
//	    // Value changed
//	}
func (cct *CustomComparatorTester) WasChanged() bool {
	return cct.changed
}

// GetCurrentValue returns the current value stored in the ref.
//
// Returns:
//   - interface{}: The current value
//
// Example:
//
//	current := tester.GetCurrentValue()
//	assert.Equal(t, expectedValue, current)
func (cct *CustomComparatorTester) GetCurrentValue() interface{} {
	return cct.ref.Get()
}

// ResetCounters resets the comparison counter and changed flag.
//
// This is useful for testing multiple scenarios in the same test
// without creating new tester instances.
//
// Example:
//
//	tester.SetValue(value1)
//	tester.AssertComparisons(t, 1)
//
//	tester.ResetCounters()
//
//	tester.SetValue(value2)
//	tester.AssertComparisons(t, 1) // Counter reset
func (cct *CustomComparatorTester) ResetCounters() {
	cct.compared = 0
	cct.changed = false
}

// CompareValues directly invokes the comparator without setting values.
//
// This is useful for testing the comparator function in isolation
// without modifying the ref.
//
// Parameters:
//   - a: First value to compare
//   - b: Second value to compare
//
// Returns:
//   - bool: True if the comparator considers the values equal
//
// Example:
//
//	areEqual := tester.CompareValues(value1, value2)
//	assert.True(t, areEqual)
func (cct *CustomComparatorTester) CompareValues(a, b interface{}) bool {
	cct.compared++
	return cct.comparator(a, b)
}

// VerifyComparatorBehavior verifies that the comparator behaves correctly
// for a set of test cases.
//
// This is a helper method for comprehensive comparator testing.
//
// Parameters:
//   - t: The testing.T instance
//   - testCases: Map of test case name to [valueA, valueB, expectedEqual]
//
// Example:
//
//	tester.VerifyComparatorBehavior(t, map[string][3]interface{}{
//	    "same ID": {User{ID: 1}, User{ID: 1}, true},
//	    "different ID": {User{ID: 1}, User{ID: 2}, false},
//	})
func (cct *CustomComparatorTester) VerifyComparatorBehavior(t *testing.T, testCases map[string][3]interface{}) {
	t.Helper()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			valueA := tc[0]
			valueB := tc[1]
			expectedEqual := tc[2].(bool)

			actualEqual := cct.CompareValues(valueA, valueB)

			if actualEqual != expectedEqual {
				t.Errorf("comparator(%v, %v): expected %v, got %v",
					valueA, valueB, expectedEqual, actualEqual)
			}
		})
	}
}

// AssertComparatorType verifies that the comparator is of the expected type.
//
// This is useful for ensuring type safety in tests.
//
// Parameters:
//   - t: The testing.T instance
//
// Example:
//
//	tester.AssertComparatorType(t)
func (cct *CustomComparatorTester) AssertComparatorType(t testing.TB) {
	t.Helper()
	if cct.comparator == nil {
		t.Error("comparator is nil")
		return
	}

	comparatorType := reflect.TypeOf(cct.comparator)
	if comparatorType.Kind() != reflect.Func {
		t.Errorf("comparator is not a function, got %v", comparatorType.Kind())
		return
	}

	// Verify function signature: func(interface{}, interface{}) bool
	if comparatorType.NumIn() != 2 {
		t.Errorf("comparator should take 2 parameters, got %d", comparatorType.NumIn())
	}
	if comparatorType.NumOut() != 1 {
		t.Errorf("comparator should return 1 value, got %d", comparatorType.NumOut())
	}
	if comparatorType.NumOut() > 0 && comparatorType.Out(0).Kind() != reflect.Bool {
		t.Errorf("comparator should return bool, got %v", comparatorType.Out(0).Kind())
	}
}
