package testutil

import (
	"reflect"
	"sync"
)

// ShowTester provides utilities for testing Show directive visibility toggling.
// It helps verify that:
//   - Content visibility toggles correctly
//   - Visibility state changes reactively
//   - Styling/markers are applied correctly
//   - Difference from If directive (visibility vs DOM presence)
//
// This tester is specifically designed for testing components that use the Show
// directive for visibility control. It tracks the visibility ref and provides
// methods to test visibility toggling behavior.
//
// # Show vs If Directive
//
// The Show directive differs from the If directive in how it handles hidden content:
//   - Show: Keeps element in output, toggles visibility (optional [Hidden] marker)
//   - If: Removes element from output completely when condition is false
//
// Use Show when you want to:
//   - Toggle visibility frequently
//   - Preserve element state while hidden
//   - Apply terminal transitions/animations
//
// Use If when you want to:
//   - Conditionally render different content
//   - Remove elements completely from output
//   - Save rendering performance for complex content
//
// Example:
//
//	visibleRef := bubbly.NewRef(true)
//	tester := NewShowTester(visibleRef)
//
//	// Test initial visibility
//	tester.AssertVisible(t, true)
//
//	// Change visibility
//	tester.SetVisible(false)
//	tester.AssertHidden(t)
//
// Thread Safety:
//
// ShowTester is thread-safe for concurrent operations using sync.RWMutex.
type ShowTester struct {
	visibleRef interface{} // *Ref[bool] - the visibility reference
	mu         sync.RWMutex
}

// NewShowTester creates a new ShowTester for testing visibility toggling.
//
// The visibleRef parameter should be a *Ref[bool] containing the visibility
// state that controls rendering. The tester will track changes to this ref.
//
// Parameters:
//   - visibleRef: A *Ref[bool] controlling visibility
//
// Returns:
//   - *ShowTester: A new tester instance
//
// Example:
//
//	showContent := bubbly.NewRef(true)
//	tester := NewShowTester(showContent)
func NewShowTester(visibleRef interface{}) *ShowTester {
	return &ShowTester{
		visibleRef: visibleRef,
	}
}

// SetVisible updates the visibility ref to the specified value.
//
// This method simulates changing the visibility state that controls rendering,
// allowing you to test reactivity and visibility toggling behavior.
//
// Parameters:
//   - value: The new boolean value for visibility
//
// Example:
//
//	tester.SetVisible(false)  // Hide content
//	tester.SetVisible(true)   // Show content
func (st *ShowTester) SetVisible(value bool) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.visibleRef == nil {
		return // Safe no-op for nil ref
	}

	// Use reflection to call Set() method on the ref
	refValue := reflect.ValueOf(st.visibleRef)
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

// GetVisible retrieves the current value of the visibility ref.
//
// This method reads the current visibility state, useful for verifying
// the state before making assertions.
//
// Returns:
//   - bool: The current visibility value, or false if ref is nil/invalid
//
// Example:
//
//	if tester.GetVisible() {
//	    // Content should be visible
//	}
func (st *ShowTester) GetVisible() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if st.visibleRef == nil {
		return false
	}

	// Use reflection to call Get() method on the ref
	refValue := reflect.ValueOf(st.visibleRef)
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

// AssertVisible asserts that content visibility matches the expected state.
//
// This method checks that the visibility ref has the expected value. When the
// visibility is true, content should be shown. When false, it should be hidden
// (either removed or marked with [Hidden] depending on transition mode).
//
// Parameters:
//   - t: The testing.T instance for assertions
//   - expected: Whether content should be visible (true) or hidden (false)
//
// Example:
//
//	tester.SetVisible(true)
//	tester.AssertVisible(t, true)  // Content should be visible
//
//	tester.SetVisible(false)
//	tester.AssertVisible(t, false) // Content should be hidden
func (st *ShowTester) AssertVisible(t testingT, expected bool) {
	t.Helper()

	actual := st.GetVisible()
	if actual != expected {
		t.Errorf("Expected visibility to be %v, but got %v", expected, actual)
	}
}

// AssertHidden asserts that content should be hidden.
//
// This is a convenience method equivalent to AssertVisible(t, false).
// It verifies that the visibility is false and content should be hidden.
//
// Parameters:
//   - t: The testing.T instance for assertions
//
// Example:
//
//	tester.SetVisible(false)
//	tester.AssertHidden(t)  // Content should be hidden
func (st *ShowTester) AssertHidden(t testingT) {
	t.Helper()
	st.AssertVisible(t, false)
}
