package testutil

import (
	"fmt"
	"reflect"
)

// nilValueStr is the string representation for nil values in error messages.
const nilValueStr = "<nil>"

// AssertRefEquals asserts that a ref's value equals the expected value.
// It uses reflect.DeepEqual for comparison, which works for all Go types including
// slices, maps, and structs.
//
// If the assertion fails, it reports an error via t.Errorf with a clear message
// showing the ref name, expected value, and actual value.
//
// Parameters:
//   - name: The name of the ref to check
//   - expected: The expected value
//
// Panics:
//   - If the ref with the given name doesn't exist
//
// Example:
//
//	ct := harness.Mount(createCounter())
//	ct.AssertRefEquals("count", 0)  // Passes if count is 0
//	ct.AssertRefEquals("count", 42) // Fails if count is not 42
func (ct *ComponentTest) AssertRefEquals(name string, expected interface{}) {
	ct.harness.t.Helper()

	// Get actual value (will panic if ref doesn't exist)
	actual := ct.state.GetRefValue(name)

	// Compare using reflect.DeepEqual
	if !reflect.DeepEqual(actual, expected) {
		ct.harness.t.Errorf("ref %q: expected %v, got %v", name, formatValue(expected), formatValue(actual))
	}
}

// AssertRefChanged asserts that a ref's value has changed from the initial value.
// It uses reflect.DeepEqual to check if the current value differs from the initial value.
//
// If the assertion fails (value hasn't changed), it reports an error via t.Errorf
// with a message showing the ref name and the unchanged value.
//
// Parameters:
//   - name: The name of the ref to check
//   - initial: The initial value to compare against
//
// Panics:
//   - If the ref with the given name doesn't exist
//
// Example:
//
//	ct := harness.Mount(createCounter())
//	initialValue := ct.state.GetRefValue("count")
//	// ... trigger some action ...
//	ct.AssertRefChanged("count", initialValue) // Passes if count changed
func (ct *ComponentTest) AssertRefChanged(name string, initial interface{}) {
	ct.harness.t.Helper()

	// Get current value (will panic if ref doesn't exist)
	actual := ct.state.GetRefValue(name)

	// Check if value changed
	if reflect.DeepEqual(actual, initial) {
		ct.harness.t.Errorf("ref %q: expected change from %v", name, formatValue(initial))
	}
}

// AssertRefType asserts that a ref's value has the expected type.
// It uses reflect.TypeOf to get the actual type and compares it with the expected type.
//
// If the assertion fails, it reports an error via t.Errorf with a message
// showing the ref name, expected type, and actual type.
//
// Parameters:
//   - name: The name of the ref to check
//   - expectedType: The expected reflect.Type (use reflect.TypeOf(value) to get it)
//
// Panics:
//   - If the ref with the given name doesn't exist
//
// Example:
//
//	ct := harness.Mount(createCounter())
//	ct.AssertRefType("count", reflect.TypeOf(0))      // Passes if count is int
//	ct.AssertRefType("name", reflect.TypeOf(""))      // Passes if name is string
//	ct.AssertRefType("items", reflect.TypeOf([]int{})) // Passes if items is []int
func (ct *ComponentTest) AssertRefType(name string, expectedType reflect.Type) {
	ct.harness.t.Helper()

	// Get actual value (will panic if ref doesn't exist)
	actual := ct.state.GetRefValue(name)

	// Get actual type
	actualType := reflect.TypeOf(actual)

	// Compare types
	if actualType != expectedType {
		ct.harness.t.Errorf("ref %q: expected type %v, got %v",
			name, formatTypeName(expectedType), formatTypeName(actualType))
	}
}

// formatValue formats a value for display in error messages.
// It handles nil values and uses fmt.Sprintf for other values.
func formatValue(v interface{}) string {
	if v == nil {
		return nilValueStr
	}
	// Use %v for general formatting, %q for strings
	if _, ok := v.(string); ok {
		return fmt.Sprintf("%q", v)
	}
	return fmt.Sprintf("%v", v)
}

// formatTypeName formats a type name for display in error messages.
// It handles nil types specially.
func formatTypeName(t reflect.Type) string {
	if t == nil {
		return nilValueStr
	}
	return t.String()
}
