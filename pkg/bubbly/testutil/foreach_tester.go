package testutil

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// ForEachTester provides utilities for testing ForEach directive list rendering.
// It helps verify that lists render correctly, items update properly, and
// additions/removals work as expected.
//
// This tester is specifically designed for testing components that use the ForEach
// directive. It allows you to:
//   - Assert on item count
//   - Assert on individual item rendering
//   - Test item additions and removals
//   - Verify rendering output
//
// The tester works with a Ref containing a slice of items and tracks the rendered
// output for each item after calling Render().
//
// Example:
//
//	items := []string{"apple", "banana", "cherry"}
//	itemsRef := bubbly.NewRef(items)
//	tester := NewForEachTester(itemsRef)
//
//	// Render items
//	renderFunc := func(item string, index int) string {
//	    return fmt.Sprintf("%d. %s\n", index+1, item)
//	}
//	tester.Render(renderFunc)
//
//	// Assert on count
//	tester.AssertItemCount(t, 3)
//
//	// Assert on individual items
//	tester.AssertItemRendered(t, 0, "1. apple\n")
//	tester.AssertItemRendered(t, 1, "2. banana\n")
//
// Thread Safety:
//
// ForEachTester is thread-safe for concurrent reads and renders using sync.RWMutex.
type ForEachTester struct {
	itemsRef interface{} // *Ref[[]T] - holds the items slice
	rendered []string    // Rendered output for each item
	mu       sync.RWMutex
}

// NewForEachTester creates a new ForEachTester for testing list rendering.
//
// The itemsRef must be a *Ref containing a slice of any type. The tester
// will track the items and their rendered output.
//
// Parameters:
//   - itemsRef: A *Ref[[]T] containing the items to test
//
// Returns:
//   - *ForEachTester: A new tester instance
//
// Example:
//
//	items := []string{"apple", "banana"}
//	itemsRef := bubbly.NewRef(items)
//	tester := NewForEachTester(itemsRef)
func NewForEachTester(itemsRef interface{}) *ForEachTester {
	return &ForEachTester{
		itemsRef: itemsRef,
		rendered: []string{},
	}
}

// Render executes the render function for each item and stores the results.
//
// This method calls the provided render function for each item in the list,
// storing the rendered output for later assertions. The render function
// receives the item and its index.
//
// Parameters:
//   - renderFunc: Function to render each item, receives (item T, index int) and returns string
//
// Example:
//
//	tester.Render(func(item string, index int) string {
//	    return fmt.Sprintf("%d. %s\n", index+1, item)
//	})
func (fet *ForEachTester) Render(renderFunc interface{}) {
	fet.mu.Lock()
	defer fet.mu.Unlock()

	// Get items from ref using reflection
	items := getItemsFromRef(fet.itemsRef)
	if items == nil {
		fet.rendered = []string{}
		return
	}

	// Pre-allocate rendered slice
	fet.rendered = make([]string, len(items))

	// Render each item
	for i, item := range items {
		fet.rendered[i] = callRenderFunc(renderFunc, item, i)
	}
}

// AssertItemCount asserts that the number of items matches the expected count.
//
// Parameters:
//   - t: Testing interface for assertions
//   - expected: Expected number of items
//
// Example:
//
//	tester.AssertItemCount(t, 3)
func (fet *ForEachTester) AssertItemCount(t testingT, expected int) {
	t.Helper()

	fet.mu.RLock()
	defer fet.mu.RUnlock()

	// Get current item count
	items := getItemsFromRef(fet.itemsRef)
	actual := len(items)

	if actual != expected {
		t.Errorf("item count mismatch: expected %d items, got %d", expected, actual)
	}
}

// AssertItemRendered asserts that the item at the given index rendered to the expected output.
//
// This method checks that the rendered output for a specific item matches the expected string.
// You must call Render() before using this assertion.
//
// Parameters:
//   - t: Testing interface for assertions
//   - index: Zero-based index of the item
//   - expected: Expected rendered output for the item
//
// Example:
//
//	tester.Render(renderFunc)
//	tester.AssertItemRendered(t, 0, "1. apple\n")
func (fet *ForEachTester) AssertItemRendered(t testingT, index int, expected string) {
	t.Helper()

	fet.mu.RLock()
	defer fet.mu.RUnlock()

	// Check index bounds
	if index < 0 || index >= len(fet.rendered) {
		t.Errorf("index %d out of bounds (rendered items: %d)", index, len(fet.rendered))
		return
	}

	actual := fet.rendered[index]
	if actual != expected {
		t.Errorf("item %d rendering mismatch:\nexpected: %q\ngot:      %q", index, expected, actual)
	}
}

// GetRendered returns all rendered items as a slice of strings.
//
// Returns:
//   - []string: Slice of rendered output for each item
//
// Example:
//
//	rendered := tester.GetRendered()
//	assert.Equal(t, 3, len(rendered))
func (fet *ForEachTester) GetRendered() []string {
	fet.mu.RLock()
	defer fet.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]string, len(fet.rendered))
	copy(result, fet.rendered)
	return result
}

// GetFullOutput returns the concatenated output of all rendered items.
//
// Returns:
//   - string: All rendered items joined together
//
// Example:
//
//	output := tester.GetFullOutput()
//	assert.Equal(t, "abc", output)
func (fet *ForEachTester) GetFullOutput() string {
	fet.mu.RLock()
	defer fet.mu.RUnlock()

	return strings.Join(fet.rendered, "")
}

// getItemsFromRef extracts the slice from a Ref using reflection.
// Returns empty slice if ref contains empty slice, nil if ref is nil.
func getItemsFromRef(itemsRef interface{}) []interface{} {
	if itemsRef == nil {
		return nil
	}

	// Use reflection to call Get() method on the ref
	refValue := reflect.ValueOf(itemsRef)
	if !refValue.IsValid() {
		return nil
	}

	// Check if pointer is nil
	if refValue.Kind() == reflect.Ptr && refValue.IsNil() {
		return nil
	}

	// Call Get() method
	getMethod := refValue.MethodByName("Get")
	if !getMethod.IsValid() {
		return nil
	}

	// Call Get() and get the result
	results := getMethod.Call(nil)
	if len(results) == 0 {
		return nil
	}

	// Get the slice value
	sliceValue := results[0]
	if !sliceValue.IsValid() {
		return nil
	}

	// If it's an interface, unwrap it to get the actual value
	if sliceValue.Kind() == reflect.Interface {
		sliceValue = sliceValue.Elem()
		if !sliceValue.IsValid() {
			return nil
		}
	}

	// Convert slice to []interface{}
	if sliceValue.Kind() != reflect.Slice {
		return nil
	}

	// Check if slice is nil (only for pointer/interface/slice/map/chan/func)
	if sliceValue.IsNil() {
		return nil
	}

	length := sliceValue.Len()
	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		result[i] = sliceValue.Index(i).Interface()
	}

	return result
}

// callRenderFunc calls the render function with the item and index.
// It handles different function signatures generically using reflection.
func callRenderFunc(renderFunc interface{}, item interface{}, index int) string {
	// Use reflection to call the function
	fnValue := reflect.ValueOf(renderFunc)
	if !fnValue.IsValid() || fnValue.Kind() != reflect.Func {
		// Fallback: return string representation
		return fmt.Sprintf("%v", item)
	}

	// Prepare arguments: item and index
	args := []reflect.Value{
		reflect.ValueOf(item),
		reflect.ValueOf(index),
	}

	// Call the function
	results := fnValue.Call(args)
	if len(results) == 0 {
		return ""
	}

	// Get the string result
	result := results[0]
	if result.Kind() == reflect.String {
		return result.String()
	}

	// Fallback: convert to string
	return fmt.Sprintf("%v", result.Interface())
}
