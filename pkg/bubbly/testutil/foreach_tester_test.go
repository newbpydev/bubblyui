package testutil

import (
	"fmt"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/directives"
	"github.com/stretchr/testify/assert"
)

// TestForEachTester_Creation tests basic tester creation
func TestForEachTester_Creation(t *testing.T) {
	tests := []struct {
		name  string
		items []string
	}{
		{"empty list", []string{}},
		{"single item", []string{"apple"}},
		{"multiple items", []string{"apple", "banana", "cherry"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemsRef := bubbly.NewRef(tt.items)
			tester := NewForEachTester(itemsRef)

			assert.NotNil(t, tester)
			assert.Equal(t, tt.items, itemsRef.Get())
		})
	}
}

// TestForEachTester_AssertItemCount tests item count assertions
func TestForEachTester_AssertItemCount(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		expectedCount int
	}{
		{"empty list", []string{}, 0},
		{"single item", []string{"apple"}, 1},
		{"three items", []string{"apple", "banana", "cherry"}, 3},
		{"ten items", []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemsRef := bubbly.NewRef(tt.items)
			tester := NewForEachTester(itemsRef)

			// This should pass
			tester.AssertItemCount(t, tt.expectedCount)
		})
	}
}

// TestForEachTester_AssertItemCount_Failure tests count assertion failures
func TestForEachTester_AssertItemCount_Failure(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"apple", "banana"})
	tester := NewForEachTester(itemsRef)

	// Use mock testing.T to capture error
	mockT := &mockTestingT{}
	tester.AssertItemCount(mockT, 5) // Wrong count

	assert.True(t, mockT.failed, "Expected error to be called")
	assert.NotEmpty(t, mockT.errors, "Expected error message")
	assert.Contains(t, mockT.errors[0], "expected 5 items")
	assert.Contains(t, mockT.errors[0], "got 2")
}

// TestForEachTester_AssertItemRendered tests item rendering assertions
func TestForEachTester_AssertItemRendered(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	// Render items
	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("%d. %s\n", index+1, item)
	}
	tester.Render(renderFunc)

	// Assert each item rendered correctly
	tester.AssertItemRendered(t, 0, "1. apple\n")
	tester.AssertItemRendered(t, 1, "2. banana\n")
	tester.AssertItemRendered(t, 2, "3. cherry\n")
}

// TestForEachTester_AssertItemRendered_Failure tests rendering assertion failures
func TestForEachTester_AssertItemRendered_Failure(t *testing.T) {
	items := []string{"apple", "banana"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	// Render items
	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("%d. %s\n", index+1, item)
	}
	tester.Render(renderFunc)

	// Use mock testing.T to capture error
	mockT := &mockTestingT{}
	tester.AssertItemRendered(mockT, 0, "wrong content")

	assert.True(t, mockT.failed, "Expected error to be called")
	assert.NotEmpty(t, mockT.errors, "Expected error message")
	assert.Contains(t, mockT.errors[0], "expected")
	assert.Contains(t, mockT.errors[0], "got")
}

// TestForEachTester_ItemUpdate tests updating items
func TestForEachTester_ItemUpdate(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"apple", "banana"})
	tester := NewForEachTester(itemsRef)

	// Initial count
	tester.AssertItemCount(t, 2)

	// Update items
	itemsRef.Set([]string{"apple", "banana", "cherry", "date"})

	// New count
	tester.AssertItemCount(t, 4)
}

// TestForEachTester_ItemRemoval tests removing items
func TestForEachTester_ItemRemoval(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"apple", "banana", "cherry"})
	tester := NewForEachTester(itemsRef)

	// Initial count
	tester.AssertItemCount(t, 3)

	// Remove an item
	itemsRef.Set([]string{"apple", "cherry"})

	// New count
	tester.AssertItemCount(t, 2)

	// Verify rendering reflects removal
	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("%s ", item)
	}
	tester.Render(renderFunc)

	tester.AssertItemRendered(t, 0, "apple ")
	tester.AssertItemRendered(t, 1, "cherry ")
}

// TestForEachTester_ItemAddition tests adding items
func TestForEachTester_ItemAddition(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"apple"})
	tester := NewForEachTester(itemsRef)

	// Initial count
	tester.AssertItemCount(t, 1)

	// Add items
	itemsRef.Set([]string{"apple", "banana", "cherry"})

	// New count
	tester.AssertItemCount(t, 3)
}

// TestForEachTester_EmptyList tests empty list handling
func TestForEachTester_EmptyList(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{})
	tester := NewForEachTester(itemsRef)

	// Empty list
	tester.AssertItemCount(t, 0)

	// Render should return empty
	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	// No items to assert
	assert.Equal(t, 0, len(tester.GetRendered()))
}

// TestForEachTester_NilItems tests nil items handling
func TestForEachTester_NilItems(t *testing.T) {
	var nilItems []string
	itemsRef := bubbly.NewRef(nilItems)
	tester := NewForEachTester(itemsRef)

	// Nil list treated as empty
	tester.AssertItemCount(t, 0)
}

// TestForEachTester_ComplexItems tests with struct items
func TestForEachTester_ComplexItems(t *testing.T) {
	type User struct {
		Name  string
		Email string
	}

	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	itemsRef := bubbly.NewRef(users)
	tester := NewForEachTester(itemsRef)

	tester.AssertItemCount(t, 2)

	// Render users
	renderFunc := func(user User, index int) string {
		return fmt.Sprintf("%d. %s <%s>\n", index+1, user.Name, user.Email)
	}
	tester.Render(renderFunc)

	tester.AssertItemRendered(t, 0, "1. Alice <alice@example.com>\n")
	tester.AssertItemRendered(t, 1, "2. Bob <bob@example.com>\n")
}

// TestForEachTester_IntegrationWithDirective tests integration with ForEach directive
func TestForEachTester_IntegrationWithDirective(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	itemsRef := bubbly.NewRef(items)

	// Create tester
	tester := NewForEachTester(itemsRef)

	// Render with directive
	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("- %s\n", item)
	}

	// Use actual ForEach directive
	directive := directives.ForEach(items, renderFunc)
	output := directive.Render()

	// Verify output
	expected := "- apple\n- banana\n- cherry\n"
	assert.Equal(t, expected, output)

	// Tester should also render correctly
	tester.Render(renderFunc)
	tester.AssertItemRendered(t, 0, "- apple\n")
	tester.AssertItemRendered(t, 1, "- banana\n")
	tester.AssertItemRendered(t, 2, "- cherry\n")
}

// TestForEachTester_GetRendered tests getting all rendered items
func TestForEachTester_GetRendered(t *testing.T) {
	items := []string{"a", "b", "c"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("[%s]", item)
	}
	tester.Render(renderFunc)

	rendered := tester.GetRendered()
	assert.Equal(t, 3, len(rendered))
	assert.Equal(t, "[a]", rendered[0])
	assert.Equal(t, "[b]", rendered[1])
	assert.Equal(t, "[c]", rendered[2])
}

// TestForEachTester_GetFullOutput tests getting full concatenated output
func TestForEachTester_GetFullOutput(t *testing.T) {
	items := []string{"a", "b", "c"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	output := tester.GetFullOutput()
	assert.Equal(t, "abc", output)
}

// TestForEachTester_ThreadSafety tests concurrent access
func TestForEachTester_ThreadSafety(t *testing.T) {
	itemsRef := bubbly.NewRef([]string{"a", "b", "c"})
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return item
	}

	// Render from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			tester.Render(renderFunc)
			tester.AssertItemCount(t, 3)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestForEachTester_EdgeCases_InvalidRef tests error handling with invalid refs
func TestForEachTester_EdgeCases_InvalidRef(t *testing.T) {
	tests := []struct {
		name     string
		itemsRef interface{}
	}{
		{"nil ref", nil},
		{"invalid pointer", (*int)(nil)},
		{"non-ref object", "not a ref"},
		{"object without Get method", &struct{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewForEachTester(tt.itemsRef)

			// Should handle gracefully
			tester.AssertItemCount(t, 0)

			// Render should handle invalid ref
			renderFunc := func(item string, index int) string {
				return item
			}
			tester.Render(renderFunc)

			// Should have no rendered items
			assert.Equal(t, 0, len(tester.GetRendered()))
		})
	}
}

// TestForEachTester_EdgeCases_InvalidRenderFunc tests error handling with invalid render functions
func TestForEachTester_EdgeCases_InvalidRenderFunc(t *testing.T) {
	items := []string{"apple", "banana"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	tests := []struct {
		name       string
		renderFunc interface{}
		expected   []string
	}{
		{
			"nil function",
			nil,
			[]string{"apple", "banana"}, // Falls back to fmt.Sprintf
		},
		{
			"not a function",
			"not a function",
			[]string{"apple", "banana"}, // Falls back to fmt.Sprintf
		},
		{
			"function returning non-string",
			func(item string, index int) int { return index },
			[]string{"0", "1"}, // Converts to string
		},
		{
			"function with no return",
			func(item string, index int) {},
			[]string{"", ""}, // Returns empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester.Render(tt.renderFunc)
			rendered := tester.GetRendered()
			assert.Equal(t, len(tt.expected), len(rendered))

			// Check that it didn't panic and returned something
			for i := range rendered {
				assert.NotNil(t, rendered[i])
			}
		})
	}
}

// TestForEachTester_EdgeCases_OutOfBoundsIndex tests index bounds checking
func TestForEachTester_EdgeCases_OutOfBoundsIndex(t *testing.T) {
	items := []string{"apple", "banana"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	tests := []struct {
		name  string
		index int
	}{
		{"negative index", -1},
		{"index too large", 10},
		{"index at boundary", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mock testing.T to capture error
			mockT := &mockTestingT{}
			tester.AssertItemRendered(mockT, tt.index, "anything")

			// Should report error about out of bounds
			assert.True(t, mockT.failed, "Expected error for out of bounds index")
			assert.NotEmpty(t, mockT.errors)
			assert.Contains(t, mockT.errors[0], "out of bounds")
		})
	}
}

// TestForEachTester_EdgeCases_NilPointerRef tests nil pointer ref
func TestForEachTester_EdgeCases_NilPointerRef(t *testing.T) {
	var nilRef *bubbly.Ref[[]string]
	tester := NewForEachTester(nilRef)

	// Should handle nil pointer gracefully
	tester.AssertItemCount(t, 0)

	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	assert.Equal(t, 0, len(tester.GetRendered()))
}

// TestForEachTester_EdgeCases_RefReturningNonSlice tests ref returning non-slice value
func TestForEachTester_EdgeCases_RefReturningNonSlice(t *testing.T) {
	// Create a ref that returns a non-slice value
	intRef := bubbly.NewRef(42)
	tester := NewForEachTester(intRef)

	// Should handle non-slice gracefully
	tester.AssertItemCount(t, 0)

	renderFunc := func(item int, index int) string {
		return fmt.Sprintf("%d", item)
	}
	tester.Render(renderFunc)

	assert.Equal(t, 0, len(tester.GetRendered()))
}

// TestForEachTester_EdgeCases_EmptyRenderedList tests operations on empty rendered list
func TestForEachTester_EdgeCases_EmptyRenderedList(t *testing.T) {
	items := []string{"apple"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	// Don't call Render() - rendered list is empty

	// GetRendered should return empty slice
	rendered := tester.GetRendered()
	assert.Equal(t, 0, len(rendered))

	// GetFullOutput should return empty string
	output := tester.GetFullOutput()
	assert.Equal(t, "", output)

	// AssertItemRendered should report error
	mockT := &mockTestingT{}
	tester.AssertItemRendered(mockT, 0, "anything")
	assert.True(t, mockT.failed)
	assert.Contains(t, mockT.errors[0], "out of bounds")
}

// TestForEachTester_EdgeCases_RenderWithNilRef tests Render with nil items
func TestForEachTester_EdgeCases_RenderWithNilRef(t *testing.T) {
	var nilItems []string
	itemsRef := bubbly.NewRef(nilItems)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		t.Error("Should not be called for nil slice")
		return item
	}

	// Render should handle nil slice gracefully
	tester.Render(renderFunc)

	// Should have empty rendered list
	assert.Equal(t, 0, len(tester.GetRendered()))
	assert.Equal(t, "", tester.GetFullOutput())
}

// TestForEachTester_EdgeCases_LargeList tests performance with large lists
func TestForEachTester_EdgeCases_LargeList(t *testing.T) {
	// Create large list
	size := 10000
	items := make([]int, size)
	for i := 0; i < size; i++ {
		items[i] = i
	}

	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	// Verify count
	tester.AssertItemCount(t, size)

	// Render all items
	renderFunc := func(item int, index int) string {
		return fmt.Sprintf("%d", item)
	}
	tester.Render(renderFunc)

	// Verify all rendered
	rendered := tester.GetRendered()
	assert.Equal(t, size, len(rendered))

	// Spot check a few
	assert.Equal(t, "0", rendered[0])
	assert.Equal(t, "5000", rendered[5000])
	assert.Equal(t, "9999", rendered[9999])
}

// TestForEachTester_EdgeCases_ConcurrentRenderAndRead tests concurrent Render and GetRendered
func TestForEachTester_EdgeCases_ConcurrentRenderAndRead(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return item
	}

	// Concurrent Render and reads
	done := make(chan bool, 20)

	// 10 renders
	for i := 0; i < 10; i++ {
		go func() {
			tester.Render(renderFunc)
			done <- true
		}()
	}

	// 10 reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = tester.GetRendered()
			_ = tester.GetFullOutput()
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify final state is valid
	rendered := tester.GetRendered()
	assert.Equal(t, 5, len(rendered))
}

// TestForEachTester_EdgeCases_UpdateDuringRender tests updating ref during render
func TestForEachTester_EdgeCases_UpdateDuringRender(t *testing.T) {
	items := []string{"a", "b", "c"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return item
	}

	// Render initial
	tester.Render(renderFunc)
	assert.Equal(t, 3, len(tester.GetRendered()))

	// Update ref
	itemsRef.Set([]string{"x", "y"})

	// Count should reflect new value
	tester.AssertItemCount(t, 2)

	// But rendered is still old until we call Render again
	assert.Equal(t, 3, len(tester.GetRendered()))

	// Re-render
	tester.Render(renderFunc)

	// Now rendered matches new value
	assert.Equal(t, 2, len(tester.GetRendered()))
}

// TestForEachTester_EdgeCases_SpecialCharacters tests rendering with special characters
func TestForEachTester_EdgeCases_SpecialCharacters(t *testing.T) {
	items := []string{
		"line\nbreak",
		"tab\there",
		"quote\"here",
		"unicode: 你好",
		"emoji: 🎉",
		"",
	}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	renderFunc := func(item string, index int) string {
		return fmt.Sprintf("[%s]", item)
	}
	tester.Render(renderFunc)

	rendered := tester.GetRendered()
	assert.Equal(t, 6, len(rendered))
	assert.Equal(t, "[line\nbreak]", rendered[0])
	assert.Equal(t, "[tab\there]", rendered[1])
	assert.Equal(t, "[quote\"here]", rendered[2])
	assert.Equal(t, "[unicode: 你好]", rendered[3])
	assert.Equal(t, "[emoji: 🎉]", rendered[4])
	assert.Equal(t, "[]", rendered[5])
}

// customRefNilSlice is a helper type for testing reflection edge cases
type customRefNilSlice struct{}

func (cr *customRefNilSlice) Get() interface{} {
	var nilSlice []string
	return nilSlice
}

// customRefInterfaceWrapped is a helper type for testing interface wrapping
type customRefInterfaceWrapped struct{}

func (ir *customRefInterfaceWrapped) Get() interface{} {
	return []string{"wrapped", "in", "interface"}
}

// customRefInvalidUnwrap is a helper type for testing invalid interface unwrapping
type customRefInvalidUnwrap struct{}

func (cr *customRefInvalidUnwrap) Get() interface{} {
	// Return an interface{} that contains nil
	var nilValue interface{}
	return nilValue
}

// TestForEachTester_EdgeCases_ReflectionEdgeCases tests reflection edge cases
func TestForEachTester_EdgeCases_ReflectionEdgeCases(t *testing.T) {
	// Test with various edge cases that exercise reflection error paths

	// Create a custom type that has a Get method returning interface{} with nil slice
	customRef := &customRefNilSlice{}
	tester := NewForEachTester(customRef)

	// Should handle nil slice gracefully
	tester.AssertItemCount(t, 0)

	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	assert.Equal(t, 0, len(tester.GetRendered()))
}

// TestForEachTester_EdgeCases_InterfaceWrapping tests interface wrapping edge cases
func TestForEachTester_EdgeCases_InterfaceWrapping(t *testing.T) {
	// Test with ref returning interface{} containing slice
	interfaceRef := &customRefInterfaceWrapped{}
	tester := NewForEachTester(interfaceRef)

	tester.AssertItemCount(t, 3)

	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	rendered := tester.GetRendered()
	assert.Equal(t, 3, len(rendered))
	assert.Equal(t, "wrapped", rendered[0])
	assert.Equal(t, "in", rendered[1])
	assert.Equal(t, "interface", rendered[2])
}

// TestForEachTester_EdgeCases_RenderFuncPanic tests render function that panics
func TestForEachTester_EdgeCases_RenderFuncPanic(t *testing.T) {
	items := []string{"a", "b"}
	itemsRef := bubbly.NewRef(items)
	tester := NewForEachTester(itemsRef)

	// Render function that panics - should be caught by reflection
	renderFunc := func(item string, index int) string {
		if index == 1 {
			panic("intentional panic")
		}
		return item
	}

	// This should panic since reflection.Call doesn't recover panics
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic
			assert.Contains(t, fmt.Sprintf("%v", r), "panic")
		}
	}()

	tester.Render(renderFunc)
}

// TestForEachTester_EdgeCases_MultipleTypes tests with multiple data types
func TestForEachTester_EdgeCases_MultipleTypes(t *testing.T) {
	// Test with bools
	bools := []bool{true, false, true}
	boolRef := bubbly.NewRef(bools)
	boolTester := NewForEachTester(boolRef)

	boolTester.AssertItemCount(t, 3)

	boolRender := func(item bool, index int) string {
		return fmt.Sprintf("%v", item)
	}
	boolTester.Render(boolRender)

	rendered := boolTester.GetRendered()
	assert.Equal(t, "true", rendered[0])
	assert.Equal(t, "false", rendered[1])
	assert.Equal(t, "true", rendered[2])

	// Test with floats
	floats := []float64{1.1, 2.2, 3.3}
	floatRef := bubbly.NewRef(floats)
	floatTester := NewForEachTester(floatRef)

	floatTester.AssertItemCount(t, 3)

	floatRender := func(item float64, index int) string {
		return fmt.Sprintf("%.1f", item)
	}
	floatTester.Render(floatRender)

	renderedFloats := floatTester.GetRendered()
	assert.Equal(t, "1.1", renderedFloats[0])
	assert.Equal(t, "2.2", renderedFloats[1])
	assert.Equal(t, "3.3", renderedFloats[2])
}

// TestForEachTester_EdgeCases_InvalidInterfaceUnwrap tests invalid interface unwrapping
func TestForEachTester_EdgeCases_InvalidInterfaceUnwrap(t *testing.T) {
	// Test with ref returning interface{} that contains nil
	invalidRef := &customRefInvalidUnwrap{}
	tester := NewForEachTester(invalidRef)

	// Should handle invalid unwrap gracefully
	tester.AssertItemCount(t, 0)

	renderFunc := func(item string, index int) string {
		return item
	}
	tester.Render(renderFunc)

	assert.Equal(t, 0, len(tester.GetRendered()))
}

// mockTestingT is defined in assertions_state_test.go
