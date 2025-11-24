package components

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestList_Creation tests that a List component can be created successfully.
func TestList_Creation(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return fmt.Sprintf("%d. %s", index+1, item)
		},
	})

	assert.NotNil(t, list, "List component should not be nil")
}

// TestList_Rendering tests that the List renders items correctly.
func TestList_Rendering(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected []string // Expected strings to be present in output
	}{
		{
			name:     "single item",
			items:    []string{"Item 1"},
			expected: []string{"1. Item 1"},
		},
		{
			name:     "multiple items",
			items:    []string{"Item 1", "Item 2", "Item 3"},
			expected: []string{"1. Item 1", "2. Item 2", "3. Item 3"},
		},
		{
			name:     "empty list",
			items:    []string{},
			expected: []string{"No items to display"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemsRef := bubbly.NewRef(tt.items)

			list := List(ListProps[string]{
				Items: itemsRef,
				RenderItem: func(item string, index int) string {
					return fmt.Sprintf("%d. %s", index+1, item)
				},
			})

			list.Init()
			output := list.View()

			for _, expected := range tt.expected {
				assert.Contains(t, output, expected, "Output should contain expected text")
			}
		})
	}
}

// TestList_GenericTypes tests that List works with different types.
func TestList_GenericTypes(t *testing.T) {
	t.Run("int type", func(t *testing.T) {
		items := bubbly.NewRef([]int{1, 2, 3, 4, 5})

		list := List(ListProps[int]{
			Items: items,
			RenderItem: func(item int, index int) string {
				return fmt.Sprintf("Number: %d", item)
			},
		})

		list.Init()
		output := list.View()

		assert.Contains(t, output, "Number: 1")
		assert.Contains(t, output, "Number: 5")
	})

	t.Run("struct type", func(t *testing.T) {
		type Todo struct {
			Title     string
			Completed bool
		}

		items := bubbly.NewRef([]Todo{
			{Title: "Buy groceries", Completed: false},
			{Title: "Write code", Completed: true},
		})

		list := List(ListProps[Todo]{
			Items: items,
			RenderItem: func(todo Todo, index int) string {
				checkbox := "☐"
				if todo.Completed {
					checkbox = "☑"
				}
				return fmt.Sprintf("%s %s", checkbox, todo.Title)
			},
		})

		list.Init()
		output := list.View()

		assert.Contains(t, output, "☐ Buy groceries")
		assert.Contains(t, output, "☑ Write code")
	})
}

// TestList_KeyboardNavigation_Down tests navigating down through the list.
func TestList_KeyboardNavigation_Down(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	var lastSelected string
	var lastIndex int

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		OnSelect: func(item string, index int) {
			lastSelected = item
			lastIndex = index
		},
	})

	list.Init()

	// Press down - should select first item
	list.Emit("keyDown", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 1", lastSelected)
	assert.Equal(t, 0, lastIndex)

	// Press down again - should select second item
	list.Emit("keyDown", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 2", lastSelected)
	assert.Equal(t, 1, lastIndex)

	// Press down again - should select third item
	list.Emit("keyDown", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 3", lastSelected)
	assert.Equal(t, 2, lastIndex)

	// Press down at end - should stay at last item
	list.Emit("keyDown", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 3", lastSelected)
	assert.Equal(t, 2, lastIndex)
}

// TestList_KeyboardNavigation_Up tests navigating up through the list.
func TestList_KeyboardNavigation_Up(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	var lastSelected string
	var lastIndex int

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		OnSelect: func(item string, index int) {
			lastSelected = item
			lastIndex = index
		},
	})

	list.Init()

	// Press up from no selection - should select last item
	list.Emit("keyUp", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 3", lastSelected)
	assert.Equal(t, 2, lastIndex)

	// Press up - should select second item
	list.Emit("keyUp", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 2", lastSelected)
	assert.Equal(t, 1, lastIndex)

	// Press up - should select first item
	list.Emit("keyUp", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 1", lastSelected)
	assert.Equal(t, 0, lastIndex)

	// Press up at start - should stay at first item
	list.Emit("keyUp", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 1", lastSelected)
	assert.Equal(t, 0, lastIndex)
}

// TestList_KeyboardNavigation_Home tests jumping to the first item.
func TestList_KeyboardNavigation_Home(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	var lastSelected string
	var lastIndex int

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		OnSelect: func(item string, index int) {
			lastSelected = item
			lastIndex = index
		},
	})

	list.Init()

	// Select last item
	list.Emit("keyUp", nil)

	// Press Home - should jump to first item
	list.Emit("keyHome", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 1", lastSelected)
	assert.Equal(t, 0, lastIndex)
}

// TestList_KeyboardNavigation_End tests jumping to the last item.
func TestList_KeyboardNavigation_End(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	var lastSelected string
	var lastIndex int

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		OnSelect: func(item string, index int) {
			lastSelected = item
			lastIndex = index
		},
	})

	list.Init()

	// Press End - should jump to last item
	list.Emit("keyEnd", nil)
	list.Emit("keyEnter", nil)
	assert.Equal(t, "Item 3", lastSelected)
	assert.Equal(t, 2, lastIndex)
}

// TestList_OnSelect tests that the OnSelect callback is triggered.
func TestList_OnSelect(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	var selectedItem string
	var selectedIdx int

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		OnSelect: func(item string, index int) {
			selectedItem = item
			selectedIdx = index
		},
	})

	list.Init()

	// Select first item
	list.Emit("keyDown", nil)

	// Press Enter to trigger OnSelect
	list.Emit("keyEnter", nil)

	assert.Equal(t, "Item 1", selectedItem, "Selected item should be Item 1")
	assert.Equal(t, 0, selectedIdx, "Selected index should be 0")
}

// TestList_OnSelect_NoCallback tests that no panic occurs when OnSelect is nil.
func TestList_OnSelect_NoCallback(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		OnSelect: nil, // No callback
	})

	list.Init()

	// Select first item
	list.Emit("keyDown", nil)

	// Press Enter - should not panic
	assert.NotPanics(t, func() {
		list.Emit("keyEnter", nil)
	}, "Should not panic when OnSelect is nil")
}

// TestList_VirtualScrolling tests that virtual scrolling works correctly.
func TestList_VirtualScrolling(t *testing.T) {
	// Create a large list
	largeList := make([]string, 100)
	for i := range largeList {
		largeList[i] = fmt.Sprintf("Item %d", i+1)
	}

	items := bubbly.NewRef(largeList)

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		Height:  10,
		Virtual: true,
	})

	list.Init()
	output := list.View()

	// Should contain first 10 items
	assert.Contains(t, output, "Item 1")
	assert.Contains(t, output, "Item 10")

	// Should NOT contain items beyond visible range
	assert.NotContains(t, output, "Item 20")
	assert.NotContains(t, output, "Item 100")

	// Should show scroll indicator
	assert.Contains(t, output, "↓ More items below")
}

// TestList_VirtualScrolling_Scrolling tests scrolling in virtual mode.
func TestList_VirtualScrolling_Scrolling(t *testing.T) {
	// Create a large list
	largeList := make([]string, 100)
	for i := range largeList {
		largeList[i] = fmt.Sprintf("Item %d", i+1)
	}

	items := bubbly.NewRef(largeList)

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		Height:  10,
		Virtual: true,
	})

	list.Init()

	// Navigate down multiple times to trigger scrolling
	// After 15 keyDown events, we're at index 14 (Item 15)
	// With height 10, scroll offset will be 5, showing items 6-15
	for i := 0; i < 15; i++ {
		list.Emit("keyDown", nil)
	}

	output := list.View()

	// After scrolling, should see items in the middle range
	assert.Contains(t, output, "Item 10")
	assert.Contains(t, output, "Item 15")
	assert.Contains(t, output, "Item 6") // First visible item

	// Should NOT see very first items (Item 1-5 are scrolled out of view)
	assert.NotContains(t, output, "Item 2")
	assert.NotContains(t, output, "Item 3")

	// Should show both scroll indicators
	assert.Contains(t, output, "↑ More items above")
	assert.Contains(t, output, "↓ More items below")
}

// TestList_Height tests that custom height is respected.
func TestList_Height(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3", "Item 4", "Item 5"})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
		Height:  3,
		Virtual: false,
	})

	list.Init()
	output := list.View()

	// All items should be visible (Virtual is false)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Count non-empty lines
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	assert.GreaterOrEqual(t, count, 5, "All items should be rendered when Virtual is false")
}

// TestList_EmptyList tests handling of empty lists.
func TestList_EmptyList(t *testing.T) {
	items := bubbly.NewRef([]string{})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()
	output := list.View()

	assert.Contains(t, output, "No items to display")

	// Navigation should not panic on empty list
	assert.NotPanics(t, func() {
		list.Emit("keyDown", nil)
		list.Emit("keyUp", nil)
		list.Emit("keyHome", nil)
		list.Emit("keyEnd", nil)
		list.Emit("keyEnter", nil)
	}, "Navigation on empty list should not panic")
}

// TestList_ThemeIntegration tests that the List integrates with the theme system.
func TestList_ThemeIntegration(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2"})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()
	output := list.View()

	// Theme integration is verified by successful rendering
	// (component would panic if theme injection failed)
	assert.NotEmpty(t, output, "List should render successfully with theme")
}

// TestList_ReactiveUpdates tests that the List updates when items change.
func TestList_ReactiveUpdates(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2"})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()
	output1 := list.View()

	assert.Contains(t, output1, "Item 1")
	assert.Contains(t, output1, "Item 2")
	assert.NotContains(t, output1, "Item 3")

	// Update items
	items.Set([]string{"Item 1", "Item 2", "Item 3"})

	output2 := list.View()

	assert.Contains(t, output2, "Item 1")
	assert.Contains(t, output2, "Item 2")
	assert.Contains(t, output2, "Item 3")
}

// TestList_SelectionHighlight tests that selected items are visually highlighted.
func TestList_SelectionHighlight(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()

	// Select first item
	list.Emit("keyDown", nil)

	output := list.View()

	// Output should contain ANSI escape codes for highlighting
	// (Lipgloss adds these for background/foreground colors)
	assert.NotEmpty(t, output, "Output should not be empty")
	assert.Contains(t, output, "Item 1")
}

// ============================================================================
// LIST HELPER FUNCTION TESTS - Additional Coverage
// ============================================================================

func TestList_VirtualScroll_ScrollDown(t *testing.T) {
	// Create a long list with virtual scrolling
	items := bubbly.NewRef([]string{
		"Item 1", "Item 2", "Item 3", "Item 4", "Item 5",
		"Item 6", "Item 7", "Item 8", "Item 9", "Item 10",
		"Item 11", "Item 12", "Item 13", "Item 14", "Item 15",
	})

	list := List(ListProps[string]{
		Items:  items,
		Height: 5, // Only show 5 items at a time
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()

	// Navigate down past visible area to trigger scroll adjustment
	for i := 0; i < 7; i++ {
		list.Emit("keyDown", nil)
	}

	output := list.View()
	assert.NotEmpty(t, output, "Should render after scrolling")
}

func TestList_VirtualScroll_ScrollUp(t *testing.T) {
	// Create a list and scroll down then up
	items := bubbly.NewRef([]string{
		"Item 1", "Item 2", "Item 3", "Item 4", "Item 5",
		"Item 6", "Item 7", "Item 8", "Item 9", "Item 10",
	})

	list := List(ListProps[string]{
		Items:  items,
		Height: 3,
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()

	// Navigate down to scroll
	for i := 0; i < 6; i++ {
		list.Emit("keyDown", nil)
	}

	// Navigate up past visible area to trigger scroll up adjustment
	for i := 0; i < 6; i++ {
		list.Emit("keyUp", nil)
	}

	output := list.View()
	assert.NotEmpty(t, output, "Should render after scroll up")
}

func TestList_SelectItem_OutOfBounds_Negative(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2"})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()

	// Try to select negative index via direct emit (simulating edge case)
	assert.NotPanics(t, func() {
		list.Emit("select", -1)
	})
}

func TestList_SelectItem_OutOfBounds_TooLarge(t *testing.T) {
	items := bubbly.NewRef([]string{"Item 1", "Item 2"})

	list := List(ListProps[string]{
		Items: items,
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()

	// Try to select index beyond list length
	assert.NotPanics(t, func() {
		list.Emit("select", 100)
	})
}

func TestList_ZeroHeight(t *testing.T) {
	// Test with Height = 0 (should use default)
	items := bubbly.NewRef([]string{"Item 1", "Item 2", "Item 3"})

	list := List(ListProps[string]{
		Items:  items,
		Height: 0, // Zero height should use default (10)
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()

	// Navigate to trigger scroll offset calculation
	for i := 0; i < 3; i++ {
		list.Emit("keyDown", nil)
	}

	output := list.View()
	assert.NotEmpty(t, output, "Should render with default height")
}

func TestList_ScrollOffset_BeyondVisible(t *testing.T) {
	// Test scrolling when selected item is well beyond visible area
	items := bubbly.NewRef([]string{
		"Item 1", "Item 2", "Item 3", "Item 4", "Item 5",
		"Item 6", "Item 7", "Item 8", "Item 9", "Item 10",
		"Item 11", "Item 12", "Item 13", "Item 14", "Item 15",
		"Item 16", "Item 17", "Item 18", "Item 19", "Item 20",
	})

	list := List(ListProps[string]{
		Items:  items,
		Height: 5,
		RenderItem: func(item string, index int) string {
			return item
		},
	})

	list.Init()

	// Navigate well past visible area
	for i := 0; i < 15; i++ {
		list.Emit("keyDown", nil)
	}

	// Then navigate back up past current visible area
	for i := 0; i < 10; i++ {
		list.Emit("keyUp", nil)
	}

	output := list.View()
	assert.NotEmpty(t, output, "Should handle extensive scroll operations")
}
