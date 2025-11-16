package testutil

import (
	"reflect"
	"strconv"
	"sync"
)

// BindTester provides utilities for testing Bind directive two-way data binding.
// It helps verify that:
//   - Ref changes update the bound element
//   - Element changes update the Ref
//   - Two-way binding works correctly
//   - Type conversions work as expected
//
// This tester is specifically designed for testing components that use the Bind
// directive for two-way data binding. It simulates user input changes and verifies
// that the Ref is updated correctly with proper type conversion.
//
// Example:
//
//	nameRef := bubbly.NewRef("")
//	tester := NewBindTester(nameRef)
//
//	// Simulate user typing in input
//	tester.TriggerElementChange("John Doe")
//
//	// Assert ref was updated
//	tester.AssertRefUpdated(t, "John Doe")
//
// Thread Safety:
//
// BindTester is thread-safe for concurrent operations using sync.RWMutex.
type BindTester struct {
	ref interface{} // *Ref[T] - the bound reference
	mu  sync.RWMutex
}

// NewBindTester creates a new BindTester for testing two-way data binding.
//
// The ref parameter should be a *Ref[T] containing the value to bind.
// The tester will track changes to the ref and simulate element changes.
//
// Parameters:
//   - ref: A *Ref[T] to test binding with (can be nil for safe no-op behavior)
//
// Returns:
//   - *BindTester: A new tester instance
//
// Example:
//
//	nameRef := bubbly.NewRef("Alice")
//	tester := NewBindTester(nameRef)
func NewBindTester(ref interface{}) *BindTester {
	return &BindTester{
		ref: ref,
	}
}

// TriggerElementChange simulates a user changing the input element value.
//
// This method simulates what happens when a user types in an input field or
// changes a form element. It updates the bound Ref with the new value, performing
// type conversion if necessary.
//
// Type Conversion:
//   - If value is a string and Ref is int/float/bool, converts the string
//   - If value matches Ref type, sets directly
//   - If conversion fails, sets to zero value
//
// Parameters:
//   - value: The new value from the element (typically string from user input)
//
// Example:
//
//	// String input
//	tester.TriggerElementChange("new value")
//
//	// Numeric input (as string)
//	tester.TriggerElementChange("42")
//
//	// Direct value
//	tester.TriggerElementChange(42)
func (bt *BindTester) TriggerElementChange(value interface{}) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.ref == nil {
		return // Safe no-op for nil ref
	}

	// Use reflection to call Set() method on the ref
	refValue := reflect.ValueOf(bt.ref)
	if !refValue.IsValid() || refValue.Kind() == reflect.Ptr && refValue.IsNil() {
		return
	}

	// Get the Set method
	setMethod := refValue.MethodByName("Set")
	if !setMethod.IsValid() {
		return
	}

	// Get the current ref type to determine conversion
	getMethod := refValue.MethodByName("Get")
	if !getMethod.IsValid() {
		return
	}

	// Call Get() to determine the type
	results := getMethod.Call(nil)
	if len(results) == 0 {
		return
	}

	currentValue := results[0].Interface()
	targetType := reflect.TypeOf(currentValue)

	// Convert value to target type
	convertedValue := convertToType(value, targetType)

	// Call Set() with converted value
	setMethod.Call([]reflect.Value{reflect.ValueOf(convertedValue)})
}

// AssertRefUpdated asserts that the Ref value matches the expected value.
//
// This method verifies that the Ref has been updated to the expected value,
// typically after calling TriggerElementChange(). It uses reflect.DeepEqual
// for comparison to handle all Go types correctly.
//
// Parameters:
//   - t: Testing interface for assertions
//   - expected: Expected value of the Ref
//
// Example:
//
//	tester.TriggerElementChange("updated")
//	tester.AssertRefUpdated(t, "updated")
func (bt *BindTester) AssertRefUpdated(t testingT, expected interface{}) {
	t.Helper()

	bt.mu.RLock()
	defer bt.mu.RUnlock()

	if bt.ref == nil {
		t.Errorf("ref is nil, cannot assert value")
		return
	}

	// Get current value from ref
	actual := bt.GetCurrentValue()

	// Compare using reflect.DeepEqual
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("ref value mismatch:\nexpected: %v (%T)\ngot:      %v (%T)",
			expected, expected, actual, actual)
	}
}

// GetCurrentValue returns the current value of the bound Ref.
//
// This is a convenience method to access the current Ref value without
// needing to call Get() directly on the Ref.
//
// Returns:
//   - interface{}: Current value of the Ref, or nil if ref is nil
//
// Example:
//
//	current := tester.GetCurrentValue()
//	assert.Equal(t, "expected", current)
func (bt *BindTester) GetCurrentValue() interface{} {
	// Note: This method is called with lock already held by AssertRefUpdated
	// For external calls, we need to acquire the lock
	if bt.ref == nil {
		return nil
	}

	// Use reflection to call Get() method
	refValue := reflect.ValueOf(bt.ref)
	if !refValue.IsValid() || refValue.Kind() == reflect.Ptr && refValue.IsNil() {
		return nil
	}

	getMethod := refValue.MethodByName("Get")
	if !getMethod.IsValid() {
		return nil
	}

	results := getMethod.Call(nil)
	if len(results) == 0 {
		return nil
	}

	return results[0].Interface()
}

// convertToType converts a value to the target type, handling common conversions.
// This mimics the type conversion behavior of the Bind directive.
func convertToType(value interface{}, targetType reflect.Type) interface{} {
	// If value is nil, return zero value of target type
	if value == nil {
		return reflect.Zero(targetType).Interface()
	}

	valueType := reflect.TypeOf(value)

	// If types match, return as-is
	if valueType == targetType {
		return value
	}

	// Handle string to other type conversions
	if valueType.Kind() == reflect.String {
		strValue := value.(string)

		switch targetType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.Atoi(strValue)
			if err != nil {
				return reflect.Zero(targetType).Interface()
			}
			return reflect.ValueOf(intVal).Convert(targetType).Interface()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal, err := strconv.ParseUint(strValue, 10, 64)
			if err != nil {
				return reflect.Zero(targetType).Interface()
			}
			return reflect.ValueOf(uintVal).Convert(targetType).Interface()

		case reflect.Float32, reflect.Float64:
			floatVal, err := strconv.ParseFloat(strValue, 64)
			if err != nil {
				return reflect.Zero(targetType).Interface()
			}
			return reflect.ValueOf(floatVal).Convert(targetType).Interface()

		case reflect.Bool:
			if strValue == "true" || strValue == "1" {
				return true
			}
			return false

		case reflect.String:
			return strValue
		}
	}

	// Try direct conversion if possible
	valueReflect := reflect.ValueOf(value)
	if valueReflect.Type().ConvertibleTo(targetType) {
		return valueReflect.Convert(targetType).Interface()
	}

	// Fallback: return zero value
	return reflect.Zero(targetType).Interface()
}
