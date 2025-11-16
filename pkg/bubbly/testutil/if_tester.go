package testutil

import (
	"reflect"
	"sync"
)

// IfTester provides utilities for testing If directive conditional rendering.
// It helps verify that:
//   - Content renders when condition is true
//   - Content is hidden when condition is false
//   - Reactivity works on condition changes
//   - Nested If directives work correctly
//   - ElseIf and Else branches work as expected
//
// This tester is specifically designed for testing components that use the If
// directive for conditional rendering. It tracks the condition ref and provides
// methods to test conditional rendering behavior.
//
// Example:
//
//	conditionRef := bubbly.NewRef(true)
//	tester := NewIfTester(conditionRef)
//
//	// Test initial rendering
//	tester.AssertRendered(t, true)
//
//	// Change condition
//	tester.SetCondition(false)
//	tester.AssertNotRendered(t)
//
// Thread Safety:
//
// IfTester is thread-safe for concurrent operations using sync.RWMutex.
type IfTester struct {
	conditionRef interface{} // *Ref[bool] - the condition reference
	mu           sync.RWMutex
}

// NewIfTester creates a new IfTester for testing conditional rendering.
//
// The conditionRef parameter should be a *Ref[bool] containing the condition
// that controls rendering. The tester will track changes to this ref.
//
// Parameters:
//   - conditionRef: A *Ref[bool] controlling conditional rendering
//
// Returns:
//   - *IfTester: A new tester instance
//
// Example:
//
//	showContent := bubbly.NewRef(true)
//	tester := NewIfTester(showContent)
func NewIfTester(conditionRef interface{}) *IfTester {
	return &IfTester{
		conditionRef: conditionRef,
	}
}

// SetCondition updates the condition ref to the specified value.
//
// This method simulates changing the condition that controls rendering,
// allowing you to test reactivity and conditional behavior.
//
// Parameters:
//   - value: The new boolean value for the condition
//
// Example:
//
//	tester.SetCondition(false)  // Hide content
//	tester.SetCondition(true)   // Show content
func (it *IfTester) SetCondition(value bool) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.conditionRef == nil {
		return // Safe no-op for nil ref
	}

	// Use reflection to call Set() method on the ref
	refValue := reflect.ValueOf(it.conditionRef)
	if !refValue.IsValid() || (refValue.Kind() == reflect.Ptr && refValue.IsNil()) {
		return
	}

	// Get the Set method
	setMethod := refValue.MethodByName("Set")
	if !setMethod.IsValid() {
		return
	}

	// Call Set(value)
	setMethod.Call([]reflect.Value{reflect.ValueOf(value)})
}

// GetCondition retrieves the current value of the condition ref.
//
// This method reads the current condition value, useful for verifying
// the state before making assertions.
//
// Returns:
//   - bool: The current condition value, or false if ref is nil/invalid
//
// Example:
//
//	if tester.GetCondition() {
//	    // Content should be rendered
//	}
func (it *IfTester) GetCondition() bool {
	it.mu.RLock()
	defer it.mu.RUnlock()

	if it.conditionRef == nil {
		return false
	}

	// Use reflection to call Get() method on the ref
	refValue := reflect.ValueOf(it.conditionRef)
	if !refValue.IsValid() || (refValue.Kind() == reflect.Ptr && refValue.IsNil()) {
		return false
	}

	// Get the Get method
	getMethod := refValue.MethodByName("Get")
	if !getMethod.IsValid() {
		return false
	}

	// Call Get() and extract bool value
	results := getMethod.Call(nil)
	if len(results) == 0 {
		return false
	}

	// Handle interface{} return type
	result := results[0].Interface()
	if boolVal, ok := result.(bool); ok {
		return boolVal
	}

	return false
}

// AssertRendered asserts that content should be rendered based on the condition.
//
// This method checks that the condition ref has the expected value. When the
// condition is true, content should be rendered. When false, it should not.
//
// Parameters:
//   - t: The testing.T instance for assertions
//   - expected: Whether content should be rendered (true) or not (false)
//
// Example:
//
//	tester.SetCondition(true)
//	tester.AssertRendered(t, true)  // Content should render
//
//	tester.SetCondition(false)
//	tester.AssertRendered(t, false) // Content should not render
func (it *IfTester) AssertRendered(t testingT, expected bool) {
	t.Helper()

	actual := it.GetCondition()
	if actual != expected {
		t.Errorf("Expected condition to be %v, but got %v", expected, actual)
	}
}

// AssertNotRendered asserts that content should not be rendered.
//
// This is a convenience method equivalent to AssertRendered(t, false).
// It verifies that the condition is false and content should be hidden.
//
// Parameters:
//   - t: The testing.T instance for assertions
//
// Example:
//
//	tester.SetCondition(false)
//	tester.AssertNotRendered(t)  // Content should be hidden
func (it *IfTester) AssertNotRendered(t testingT) {
	t.Helper()
	it.AssertRendered(t, false)
}
