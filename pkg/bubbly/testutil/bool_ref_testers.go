package testutil

import (
	"reflect"
	"sync"
)

// BoolRefTester provides common utilities for testing boolean refs.
// It encapsulates the reflection-based ref manipulation used by both
// IfTester and ShowTester.
//
// Thread Safety:
//
// BoolRefTester is thread-safe for concurrent operations using sync.RWMutex.
type BoolRefTester struct {
	ref interface{} // *Ref[bool] - the boolean reference
	mu  sync.RWMutex
}

// NewBoolRefTester creates a new BoolRefTester for testing boolean refs.
func NewBoolRefTester(ref interface{}) *BoolRefTester {
	return &BoolRefTester{ref: ref}
}

// SetValue updates the ref to the specified boolean value using reflection.
func (brt *BoolRefTester) SetValue(value bool) {
	brt.mu.Lock()
	defer brt.mu.Unlock()

	if brt.ref == nil {
		return
	}

	refValue := reflect.ValueOf(brt.ref)
	if !refValue.IsValid() || (refValue.Kind() == reflect.Ptr && refValue.IsNil()) {
		return
	}

	setMethod := refValue.MethodByName("Set")
	if !setMethod.IsValid() {
		return
	}

	setMethod.Call([]reflect.Value{reflect.ValueOf(value)})
}

// GetValue retrieves the current value of the ref using reflection.
func (brt *BoolRefTester) GetValue() bool {
	brt.mu.RLock()
	defer brt.mu.RUnlock()

	if brt.ref == nil {
		return false
	}

	refValue := reflect.ValueOf(brt.ref)
	if !refValue.IsValid() || (refValue.Kind() == reflect.Ptr && refValue.IsNil()) {
		return false
	}

	getMethod := refValue.MethodByName("Get")
	if !getMethod.IsValid() {
		return false
	}

	results := getMethod.Call(nil)
	if len(results) == 0 {
		return false
	}

	if boolVal, ok := results[0].Interface().(bool); ok {
		return boolVal
	}
	return false
}

// AssertValue asserts that the ref value matches the expected state.
func (brt *BoolRefTester) AssertValue(t testingT, expected bool, valueDesc string) {
	t.Helper()
	if actual := brt.GetValue(); actual != expected {
		t.Errorf("Expected %s to be %v, but got %v", valueDesc, expected, actual)
	}
}

// IfTester provides utilities for testing If directive conditional rendering.
// It tracks a boolean condition ref and provides methods to test conditional
// rendering behavior through SetCondition, GetCondition, and assertion methods.
//
// Thread Safety: IfTester is thread-safe for concurrent operations.
type IfTester struct{ base *BoolRefTester }

// NewIfTester creates a new IfTester for testing conditional rendering.
// The conditionRef parameter should be a *Ref[bool] controlling rendering.
func NewIfTester(conditionRef interface{}) *IfTester {
	return &IfTester{base: NewBoolRefTester(conditionRef)}
}

// SetCondition updates the condition ref to the specified value.
func (it *IfTester) SetCondition(value bool) { it.base.SetValue(value) }

// GetCondition retrieves the current value of the condition ref.
func (it *IfTester) GetCondition() bool { return it.base.GetValue() }

// AssertRendered asserts that content rendering state matches expected.
func (it *IfTester) AssertRendered(t testingT, expected bool) {
	t.Helper()
	it.base.AssertValue(t, expected, "condition")
}

// AssertNotRendered asserts that content should not be rendered.
func (it *IfTester) AssertNotRendered(t testingT) {
	t.Helper()
	it.AssertRendered(t, false)
}

// ShowTester provides utilities for testing Show directive visibility toggling.
// Unlike IfTester (which removes content), ShowTester tracks visibility state
// where content may remain in output but be marked as hidden.
//
// Thread Safety: ShowTester is thread-safe for concurrent operations.
type ShowTester struct{ base *BoolRefTester }

// NewShowTester creates a new ShowTester for testing visibility toggling.
// The visibleRef parameter should be a *Ref[bool] controlling visibility.
func NewShowTester(visibleRef interface{}) *ShowTester {
	return &ShowTester{base: NewBoolRefTester(visibleRef)}
}

// SetVisible updates the visibility ref to the specified value.
func (st *ShowTester) SetVisible(value bool) { st.base.SetValue(value) }

// GetVisible retrieves the current value of the visibility ref.
func (st *ShowTester) GetVisible() bool { return st.base.GetValue() }

// AssertVisible asserts that content visibility matches the expected state.
func (st *ShowTester) AssertVisible(t testingT, expected bool) {
	t.Helper()
	st.base.AssertValue(t, expected, "visibility")
}

// AssertHidden asserts that content should be hidden.
func (st *ShowTester) AssertHidden(t testingT) {
	t.Helper()
	st.AssertVisible(t, false)
}
