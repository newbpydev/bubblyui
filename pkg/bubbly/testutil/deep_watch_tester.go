package testutil

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// DeepWatchTester provides utilities for testing deep object watching and nested change detection.
// It allows you to verify that watchers with the Deep option enabled correctly detect changes
// in nested fields, slice elements, and map values.
//
// Deep watching uses reflection-based comparison (reflect.DeepEqual) to detect changes in complex
// structures. This tester helps verify that behavior and provides utilities for modifying nested
// fields by path.
//
// Key Features:
//   - Test deep vs shallow watching behavior
//   - Modify nested fields by path (e.g., "user.profile.age")
//   - Track watch trigger count
//   - Track changed paths
//   - Verify array/map mutations
//
// Example:
//
//	type User struct {
//	    Name    string
//	    Profile struct {
//	        Age int
//	    }
//	}
//
//	user := bubbly.NewRef(User{Name: "John"})
//	watchCount := 0
//
//	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
//	    watchCount++
//	}, bubbly.WithDeep())
//	defer cleanup()
//
//	tester := NewDeepWatchTester(user, &watchCount, true)
//
//	// Modify nested field
//	tester.ModifyNestedField("Profile.Age", 30)
//
//	// Verify watch triggered
//	tester.AssertWatchTriggered(t, 1)
//	tester.AssertPathChanged(t, "Profile.Age")
//
// Thread Safety:
//
// DeepWatchTester is not thread-safe. It should only be used from a single test goroutine.
type DeepWatchTester struct {
	watched      interface{} // The watched Ref (must be *Ref[T])
	watchCount   *int        // Pointer to watch trigger counter
	changedPaths []string    // Paths that were modified
	deep         bool        // Whether deep watching is enabled
}

// NewDeepWatchTester creates a new DeepWatchTester for testing deep object watching.
//
// The tester requires:
//   - A Ref to watch (must be *Ref[T] where T is a struct, slice, or map)
//   - A pointer to an int that tracks watch trigger count
//   - A boolean indicating whether deep watching is enabled
//
// Parameters:
//   - ref: The Ref to watch (must be *Ref[T])
//   - watchCount: Pointer to an int that tracks watch trigger count
//   - deep: Whether deep watching is enabled
//
// Returns:
//   - *DeepWatchTester: A new tester instance
//
// Example:
//
//	user := bubbly.NewRef(User{Name: "John"})
//	watchCount := 0
//
//	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
//	    watchCount++
//	}, bubbly.WithDeep())
//	defer cleanup()
//
//	tester := NewDeepWatchTester(user, &watchCount, true)
func NewDeepWatchTester(ref interface{}, watchCount *int, deep bool) *DeepWatchTester {
	return &DeepWatchTester{
		watched:      ref,
		watchCount:   watchCount,
		changedPaths: []string{},
		deep:         deep,
	}
}

// ModifyNestedField modifies a nested field by path and tracks the change.
//
// The path uses dot notation to navigate nested structures:
//   - "Name" - top-level field
//   - "Profile.Age" - nested struct field
//   - "Tags[0]" - slice element
//   - "Metadata[key]" - map value
//
// This method:
//  1. Gets the current value from the Ref
//  2. Uses reflection to navigate to the nested field
//  3. Sets the new value
//  4. Updates the Ref with the modified value
//  5. Tracks the path in changedPaths
//
// Parameters:
//   - path: Dot-notation path to the field (e.g., "Profile.Age")
//   - value: The new value to set
//
// Example:
//
//	tester.ModifyNestedField("Name", "Jane")
//	tester.ModifyNestedField("Profile.Age", 30)
//	tester.ModifyNestedField("Tags[0]", "admin")
func (dwt *DeepWatchTester) ModifyNestedField(path string, value interface{}) {
	// Track the path
	dwt.changedPaths = append(dwt.changedPaths, path)

	// Get the current value from the Ref
	refValue := reflect.ValueOf(dwt.watched)
	if !refValue.IsValid() || refValue.Kind() != reflect.Ptr {
		return
	}

	// Call Get() method to get current value
	getMethod := refValue.MethodByName("Get")
	if !getMethod.IsValid() {
		return
	}

	results := getMethod.Call(nil)
	if len(results) == 0 {
		return
	}

	currentValue := results[0]

	// Unwrap interface{} if needed
	for currentValue.Kind() == reflect.Interface {
		if currentValue.IsNil() {
			return
		}
		currentValue = currentValue.Elem()
	}

	// Make a deep copy of the current value
	modifiedValue := dwt.deepCopy(currentValue)
	if !modifiedValue.IsValid() {
		return
	}

	// Navigate to the nested field
	// For maps, we need special handling
	if strings.Contains(path, "[") && strings.Contains(path, "]") {
		dwt.setNestedValue(modifiedValue, path, value)
	} else {
		field := dwt.navigateToField(modifiedValue, path)
		if !field.IsValid() || !field.CanSet() {
			return
		}

		// Set the new value
		newValue := reflect.ValueOf(value)
		if newValue.Type().AssignableTo(field.Type()) {
			field.Set(newValue)
		} else if newValue.Type().ConvertibleTo(field.Type()) {
			field.Set(newValue.Convert(field.Type()))
		}
	}

	// Set the modified value back to the Ref
	setMethod := refValue.MethodByName("Set")
	if !setMethod.IsValid() {
		return
	}

	setMethod.Call([]reflect.Value{modifiedValue})
}

// deepCopy creates a deep copy of a value
func (dwt *DeepWatchTester) deepCopy(value reflect.Value) reflect.Value {
	// Handle nil
	if !value.IsValid() {
		return reflect.Value{}
	}

	// Handle pointer
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		newPtr := reflect.New(value.Elem().Type())
		newPtr.Elem().Set(dwt.deepCopy(value.Elem()))
		return newPtr
	}

	// For struct, create new and copy fields
	if value.Kind() == reflect.Struct {
		newValue := reflect.New(value.Type()).Elem()
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			if newValue.Field(i).CanSet() {
				newValue.Field(i).Set(dwt.deepCopy(field))
			}
		}
		return newValue
	}

	// For slice, create new and copy elements
	if value.Kind() == reflect.Slice {
		newSlice := reflect.MakeSlice(value.Type(), value.Len(), value.Cap())
		for i := 0; i < value.Len(); i++ {
			newSlice.Index(i).Set(dwt.deepCopy(value.Index(i)))
		}
		return newSlice
	}

	// For map, create new and copy entries
	if value.Kind() == reflect.Map {
		newMap := reflect.MakeMap(value.Type())
		for _, key := range value.MapKeys() {
			newMap.SetMapIndex(key, dwt.deepCopy(value.MapIndex(key)))
		}
		return newMap
	}

	// For basic types, just return the value
	return value
}

// parseIndexedPath parses a path like "Tags[0]" into field name and index string.
func parseIndexedPath(part string) (fieldName, indexStr string, hasIndex bool) {
	if !strings.Contains(part, "[") || !strings.Contains(part, "]") {
		return part, "", false
	}
	openBracket := strings.Index(part, "[")
	closeBracket := strings.Index(part, "]")
	return part[:openBracket], part[openBracket+1 : closeBracket], true
}

// setValueSafe safely sets a value with type conversion.
func setValueSafe(elem reflect.Value, newValue interface{}) {
	if !elem.CanSet() {
		return
	}
	newVal := reflect.ValueOf(newValue)
	if newVal.Type().AssignableTo(elem.Type()) {
		elem.Set(newVal)
	} else if newVal.Type().ConvertibleTo(elem.Type()) {
		elem.Set(newVal.Convert(elem.Type()))
	}
}

// navigateToIndexedField navigates to a field by name and returns the value.
func navigateToIndexedField(current reflect.Value, fieldName string) (reflect.Value, bool) {
	if fieldName != "" {
		current = current.FieldByName(fieldName)
		if !current.IsValid() {
			return reflect.Value{}, false
		}
	}
	return current, true
}

// handleMapAccess handles map indexed access for setNestedValue.
func handleMapAccess(current reflect.Value, indexStr string, isLast bool, newValue interface{}) (reflect.Value, bool) {
	keyValue := reflect.ValueOf(indexStr)
	if isLast {
		current.SetMapIndex(keyValue, reflect.ValueOf(newValue))
		return reflect.Value{}, true // done
	}
	next := current.MapIndex(keyValue)
	return next, next.IsValid()
}

// handleSliceAccess handles slice/array indexed access for setNestedValue.
func handleSliceAccess(current reflect.Value, indexStr string, isLast bool, newValue interface{}) (reflect.Value, bool) {
	var index int
	_, _ = fmt.Sscanf(indexStr, "%d", &index)
	if index < 0 || index >= current.Len() {
		return reflect.Value{}, false
	}
	if isLast {
		setValueSafe(current.Index(index), newValue)
		return reflect.Value{}, true // done
	}
	return current.Index(index), true
}

// setNestedValue sets a value in a nested structure, handling maps and slices specially
func (dwt *DeepWatchTester) setNestedValue(value reflect.Value, path string, newValue interface{}) {
	parts := strings.Split(path, ".")
	current := value

	for i, part := range parts {
		isLast := i == len(parts)-1
		fieldName, indexStr, hasIndex := parseIndexedPath(part)

		if hasIndex {
			var ok bool
			if current, ok = navigateToIndexedField(current, fieldName); !ok {
				return
			}

			switch current.Kind() {
			case reflect.Map:
				if next, done := handleMapAccess(current, indexStr, isLast, newValue); done || !next.IsValid() {
					return
				} else {
					current = next
				}
			case reflect.Slice, reflect.Array:
				if next, ok := handleSliceAccess(current, indexStr, isLast, newValue); !ok {
					return
				} else if isLast {
					return
				} else {
					current = next
				}
			}
		} else {
			current = current.FieldByName(part)
			if !current.IsValid() {
				return
			}
			if isLast {
				setValueSafe(current, newValue)
				return
			}
		}
	}
}

// unwrapValue unwraps interface and pointer types to get the underlying value.
func unwrapValue(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Interface {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

// navigateIndexedRead handles indexed access (slice/array/map) for reading.
func navigateIndexedRead(current reflect.Value, indexStr string) reflect.Value {
	switch current.Kind() {
	case reflect.Slice, reflect.Array:
		var index int
		_, _ = fmt.Sscanf(indexStr, "%d", &index)
		if index >= 0 && index < current.Len() {
			return current.Index(index)
		}
		return reflect.Value{}
	case reflect.Map:
		keyValue := reflect.ValueOf(indexStr)
		return current.MapIndex(keyValue)
	}
	return current
}

// navigateToField navigates to a nested field using dot notation.
// Supports:
//   - Struct fields: "Profile.Age"
//   - Slice elements: "Tags[0]"
//   - Map values: "Metadata[key]"
func (dwt *DeepWatchTester) navigateToField(value reflect.Value, path string) reflect.Value {
	current := unwrapValue(value)
	if !current.IsValid() {
		return reflect.Value{}
	}

	parts := strings.Split(path, ".")
	for _, part := range parts {
		fieldName, indexStr, hasIndex := parseIndexedPath(part)

		if hasIndex {
			if fieldName != "" {
				current = current.FieldByName(fieldName)
				if !current.IsValid() {
					return reflect.Value{}
				}
			}
			current = navigateIndexedRead(current, indexStr)
			if !current.IsValid() {
				return reflect.Value{}
			}
		} else {
			current = current.FieldByName(part)
			if !current.IsValid() {
				return reflect.Value{}
			}
			current = unwrapValue(current)
			if !current.IsValid() {
				return reflect.Value{}
			}
		}
	}
	return current
}

// AssertWatchTriggered asserts that the watch was triggered the expected number of times.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: The expected trigger count
//
// Example:
//
//	tester.ModifyNestedField("Profile.Age", 30)
//	tester.AssertWatchTriggered(t, 1)
func (dwt *DeepWatchTester) AssertWatchTriggered(t testing.TB, expected int) {
	t.Helper()
	if dwt.watchCount == nil {
		t.Fatal("watch count is nil")
		return
	}

	actual := *dwt.watchCount
	if actual != expected {
		t.Errorf("expected watch to trigger %d times, but triggered %d times", expected, actual)
	}
}

// AssertPathChanged asserts that a specific path was modified.
//
// Parameters:
//   - t: The testing.T instance
//   - path: The path that should have been modified
//
// Example:
//
//	tester.ModifyNestedField("Profile.Age", 30)
//	tester.AssertPathChanged(t, "Profile.Age")
func (dwt *DeepWatchTester) AssertPathChanged(t testing.TB, path string) {
	t.Helper()
	for _, changedPath := range dwt.changedPaths {
		if changedPath == path {
			return
		}
	}
	t.Errorf("expected path %q to be changed, but it was not in changed paths: %v", path, dwt.changedPaths)
}

// GetChangedPaths returns all paths that were modified.
//
// Returns:
//   - []string: List of changed paths
//
// Example:
//
//	paths := tester.GetChangedPaths()
//	assert.Contains(t, paths, "Profile.Age")
func (dwt *DeepWatchTester) GetChangedPaths() []string {
	return dwt.changedPaths
}

// GetWatchCount returns the current watch trigger count.
//
// Returns:
//   - int: The current trigger count
//
// Example:
//
//	count := tester.GetWatchCount()
//	assert.Greater(t, count, 0)
func (dwt *DeepWatchTester) GetWatchCount() int {
	if dwt.watchCount == nil {
		return 0
	}
	return *dwt.watchCount
}

// IsDeepWatching returns whether deep watching is enabled.
//
// Returns:
//   - bool: True if deep watching is enabled
//
// Example:
//
//	if tester.IsDeepWatching() {
//	    // Deep watching behavior
//	}
func (dwt *DeepWatchTester) IsDeepWatching() bool {
	return dwt.deep
}
