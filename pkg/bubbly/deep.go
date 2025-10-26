package bubbly

import "reflect"

// DeepCompareFunc is a function type for custom deep comparison.
// It returns true if the old and new values are considered equal.
// This allows users to define their own equality logic for performance-critical paths.
//
// Example:
//
//	func compareUsers(old, new User) bool {
//	    // Only compare specific fields
//	    return old.ID == new.ID && old.Name == new.Name
//	}
type DeepCompareFunc[T any] func(old, new T) bool

// deepEqual performs deep equality comparison between two values.
// It uses reflect.DeepEqual for automatic comparison of nested structures.
//
// This function is used internally by deep watchers to determine if a value
// has actually changed when Set() is called.
//
// Performance note: reflect.DeepEqual can be 10-100x slower than shallow comparison.
// For performance-critical paths, use WithDeepCompare() with a custom comparator.
func deepEqual[T any](a, b T) bool {
	return reflect.DeepEqual(a, b)
}

// hasChanged determines if two values are different based on deep comparison settings.
// If a custom comparator is provided, it uses that. Otherwise, it falls back to
// reflection-based deep equality.
//
// Parameters:
//   - old: The previous value
//   - new: The current value
//   - compareFn: Optional custom comparison function (nil for default reflect.DeepEqual)
//
// Returns true if the values are different (change detected).
func hasChanged[T any](old, new T, compareFn DeepCompareFunc[T]) bool {
	if compareFn != nil {
		// Use custom comparator (returns true if equal)
		return !compareFn(old, new)
	}
	// Use reflection-based comparison (returns true if equal)
	return !deepEqual(old, new)
}
