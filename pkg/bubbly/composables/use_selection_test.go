package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseSelection_InitialSelectionIsZero tests that initial selection is 0 for non-empty list
func TestUseSelection_InitialSelectionIsZero(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c"}
	selection := UseSelection(ctx, items)

	assert.NotNil(t, selection, "UseSelection should return non-nil")
	assert.NotNil(t, selection.SelectedIndex, "SelectedIndex should not be nil")
	assert.Equal(t, 0, selection.SelectedIndex.GetTyped(), "Initial selection should be 0")
}

// TestUseSelection_EmptyListInitialSelection tests that empty list has -1 selection
func TestUseSelection_EmptyListInitialSelection(t *testing.T) {
	ctx := createTestContext()
	items := []string{}
	selection := UseSelection(ctx, items)

	assert.Equal(t, -1, selection.SelectedIndex.GetTyped(), "Empty list should have -1 selection")
}

// TestUseSelection_SelectNextNavigates tests SelectNext navigation
func TestUseSelection_SelectNextNavigates(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		initialIndex  int
		wrap          bool
		expectedIndex int
	}{
		{"next from 0", []string{"a", "b", "c"}, 0, false, 1},
		{"next from middle", []string{"a", "b", "c"}, 1, false, 2},
		{"next at end without wrap", []string{"a", "b", "c"}, 2, false, 2},
		{"next at end with wrap", []string{"a", "b", "c"}, 2, true, 0},
		{"single item no wrap", []string{"a"}, 0, false, 0},
		{"single item with wrap", []string{"a"}, 0, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			var selection *SelectionReturn[string]
			if tt.wrap {
				selection = UseSelection(ctx, tt.items, WithWrap(true))
			} else {
				selection = UseSelection(ctx, tt.items)
			}

			// Set initial index
			selection.Select(tt.initialIndex)

			// Navigate next
			selection.SelectNext()

			assert.Equal(t, tt.expectedIndex, selection.SelectedIndex.GetTyped(),
				"Index should be %d after SelectNext", tt.expectedIndex)
		})
	}
}

// TestUseSelection_SelectPreviousNavigates tests SelectPrevious navigation
func TestUseSelection_SelectPreviousNavigates(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		initialIndex  int
		wrap          bool
		expectedIndex int
	}{
		{"previous from 2", []string{"a", "b", "c"}, 2, false, 1},
		{"previous from middle", []string{"a", "b", "c"}, 1, false, 0},
		{"previous at start without wrap", []string{"a", "b", "c"}, 0, false, 0},
		{"previous at start with wrap", []string{"a", "b", "c"}, 0, true, 2},
		{"single item no wrap", []string{"a"}, 0, false, 0},
		{"single item with wrap", []string{"a"}, 0, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			var selection *SelectionReturn[string]
			if tt.wrap {
				selection = UseSelection(ctx, tt.items, WithWrap(true))
			} else {
				selection = UseSelection(ctx, tt.items)
			}

			// Set initial index
			selection.Select(tt.initialIndex)

			// Navigate previous
			selection.SelectPrevious()

			assert.Equal(t, tt.expectedIndex, selection.SelectedIndex.GetTyped(),
				"Index should be %d after SelectPrevious", tt.expectedIndex)
		})
	}
}

// TestUseSelection_WrapOptionEnablesCircular tests wrap option
func TestUseSelection_WrapOptionEnablesCircular(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c"}
	selection := UseSelection(ctx, items, WithWrap(true))

	// Start at end
	selection.Select(2)

	// Next should wrap to 0
	selection.SelectNext()
	assert.Equal(t, 0, selection.SelectedIndex.GetTyped(), "Should wrap to 0")

	// Previous should wrap to end
	selection.SelectPrevious()
	assert.Equal(t, 2, selection.SelectedIndex.GetTyped(), "Should wrap to 2")
}

// TestUseSelection_SelectedItemComputedCorrectly tests SelectedItem computed
func TestUseSelection_SelectedItemComputedCorrectly(t *testing.T) {
	ctx := createTestContext()
	items := []string{"apple", "banana", "cherry"}
	selection := UseSelection(ctx, items)

	// Initial selection
	assert.Equal(t, "apple", selection.SelectedItem.Get(), "Initial selected item should be 'apple'")

	// Change selection
	selection.Select(1)
	assert.Equal(t, "banana", selection.SelectedItem.Get(), "Selected item should be 'banana'")

	selection.Select(2)
	assert.Equal(t, "cherry", selection.SelectedItem.Get(), "Selected item should be 'cherry'")
}

// TestUseSelection_SelectedItemEmptyList tests SelectedItem with empty list
func TestUseSelection_SelectedItemEmptyList(t *testing.T) {
	ctx := createTestContext()
	items := []string{}
	selection := UseSelection(ctx, items)

	// Should return zero value for empty list
	assert.Equal(t, "", selection.SelectedItem.Get(), "Empty list should return zero value")
}

// TestUseSelection_IsSelectedReturnsCorrectValue tests IsSelected
func TestUseSelection_IsSelectedReturnsCorrectValue(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c"}
	selection := UseSelection(ctx, items)

	// Initial: index 0 is selected
	assert.True(t, selection.IsSelected(0), "Index 0 should be selected")
	assert.False(t, selection.IsSelected(1), "Index 1 should not be selected")
	assert.False(t, selection.IsSelected(2), "Index 2 should not be selected")

	// Change selection
	selection.Select(1)
	assert.False(t, selection.IsSelected(0), "Index 0 should not be selected")
	assert.True(t, selection.IsSelected(1), "Index 1 should be selected")
	assert.False(t, selection.IsSelected(2), "Index 2 should not be selected")
}

// TestUseSelection_SetItemsUpdatesAndAdjustsSelection tests SetItems
func TestUseSelection_SetItemsUpdatesAndAdjustsSelection(t *testing.T) {
	tests := []struct {
		name          string
		initialItems  []string
		initialIndex  int
		newItems      []string
		expectedIndex int
	}{
		{"reduce items clamps selection", []string{"a", "b", "c", "d"}, 3, []string{"a", "b"}, 1},
		{"increase items keeps selection", []string{"a", "b"}, 1, []string{"a", "b", "c", "d"}, 1},
		{"same size keeps selection", []string{"a", "b", "c"}, 1, []string{"x", "y", "z"}, 1},
		{"empty items sets -1", []string{"a", "b", "c"}, 1, []string{}, -1},
		{"from empty to items sets 0", []string{}, -1, []string{"a", "b"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			selection := UseSelection(ctx, tt.initialItems)

			// Set initial index if not empty
			if len(tt.initialItems) > 0 {
				selection.Select(tt.initialIndex)
			}

			// Update items
			selection.SetItems(tt.newItems)

			assert.Equal(t, tt.expectedIndex, selection.SelectedIndex.GetTyped(),
				"Index should be %d after SetItems", tt.expectedIndex)
			assert.Equal(t, len(tt.newItems), len(selection.Items.GetTyped()),
				"Items length should match")
		})
	}
}

// TestUseSelection_EmptyItemsListHandled tests operations on empty list
func TestUseSelection_EmptyItemsListHandled(t *testing.T) {
	ctx := createTestContext()
	items := []string{}
	selection := UseSelection(ctx, items)

	// All operations should be safe
	selection.SelectNext()
	assert.Equal(t, -1, selection.SelectedIndex.GetTyped(), "SelectNext on empty should stay -1")

	selection.SelectPrevious()
	assert.Equal(t, -1, selection.SelectedIndex.GetTyped(), "SelectPrevious on empty should stay -1")

	selection.Select(5)
	assert.Equal(t, -1, selection.SelectedIndex.GetTyped(), "Select on empty should stay -1")

	assert.False(t, selection.IsSelected(0), "IsSelected on empty should return false")
}

// TestUseSelection_SelectClampsToValidRange tests Select clamping
func TestUseSelection_SelectClampsToValidRange(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		selectIndex   int
		expectedIndex int
	}{
		{"valid index", []string{"a", "b", "c"}, 1, 1},
		{"negative clamps to 0", []string{"a", "b", "c"}, -5, 0},
		{"beyond max clamps to last", []string{"a", "b", "c"}, 10, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			selection := UseSelection(ctx, tt.items)

			selection.Select(tt.selectIndex)

			assert.Equal(t, tt.expectedIndex, selection.SelectedIndex.GetTyped(),
				"Index should be clamped to %d", tt.expectedIndex)
		})
	}
}

// TestUseSelection_MultiSelectMode tests multi-select functionality
func TestUseSelection_MultiSelectMode(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c", "d"}
	selection := UseSelection(ctx, items, WithMultiSelect(true))

	// Initially no multi-selections
	assert.Equal(t, 0, len(selection.SelectedIndices.GetTyped()), "Initially no multi-selections")

	// Toggle selection on index 1
	selection.ToggleSelection(1)
	assert.True(t, selection.IsSelected(1), "Index 1 should be selected after toggle")
	assert.Equal(t, 1, len(selection.SelectedIndices.GetTyped()), "Should have 1 selection")

	// Toggle selection on index 2
	selection.ToggleSelection(2)
	assert.True(t, selection.IsSelected(2), "Index 2 should be selected after toggle")
	assert.Equal(t, 2, len(selection.SelectedIndices.GetTyped()), "Should have 2 selections")

	// Toggle off index 1
	selection.ToggleSelection(1)
	assert.False(t, selection.IsSelected(1), "Index 1 should be deselected after toggle")
	assert.Equal(t, 1, len(selection.SelectedIndices.GetTyped()), "Should have 1 selection")

	// Clear all selections
	selection.ClearSelection()
	assert.Equal(t, 0, len(selection.SelectedIndices.GetTyped()), "Should have 0 selections after clear")
}

// TestUseSelection_MultiSelectIsSelectedChecksIndices tests IsSelected in multi-select mode
func TestUseSelection_MultiSelectIsSelectedChecksIndices(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c", "d"}
	selection := UseSelection(ctx, items, WithMultiSelect(true))

	// Select multiple
	selection.ToggleSelection(0)
	selection.ToggleSelection(2)

	assert.True(t, selection.IsSelected(0), "Index 0 should be selected")
	assert.False(t, selection.IsSelected(1), "Index 1 should not be selected")
	assert.True(t, selection.IsSelected(2), "Index 2 should be selected")
	assert.False(t, selection.IsSelected(3), "Index 3 should not be selected")
}

// TestUseSelection_ClearSelectionResets tests ClearSelection
func TestUseSelection_ClearSelectionResets(t *testing.T) {
	t.Run("single select mode", func(t *testing.T) {
		ctx := createTestContext()
		items := []string{"a", "b", "c"}
		selection := UseSelection(ctx, items)

		selection.Select(2)
		assert.Equal(t, 2, selection.SelectedIndex.GetTyped())

		selection.ClearSelection()
		assert.Equal(t, 0, selection.SelectedIndex.GetTyped(), "Should reset to 0")
	})

	t.Run("multi select mode", func(t *testing.T) {
		ctx := createTestContext()
		items := []string{"a", "b", "c"}
		selection := UseSelection(ctx, items, WithMultiSelect(true))

		selection.ToggleSelection(0)
		selection.ToggleSelection(1)
		selection.ToggleSelection(2)
		assert.Equal(t, 3, len(selection.SelectedIndices.GetTyped()))

		selection.ClearSelection()
		assert.Equal(t, 0, len(selection.SelectedIndices.GetTyped()), "Should clear all")
	})
}

// TestUseSelection_WorksWithCreateShared tests shared composable pattern
func TestUseSelection_WorksWithCreateShared(t *testing.T) {
	sharedSelection := CreateShared(func(ctx *bubbly.Context) *SelectionReturn[string] {
		return UseSelection(ctx, []string{"a", "b", "c"})
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	sel1 := sharedSelection(ctx1)
	sel2 := sharedSelection(ctx2)

	// Both should be the same instance
	sel1.Select(2)

	assert.Equal(t, 2, sel2.SelectedIndex.GetTyped(),
		"Shared instance should have same selection state")
}

// TestUseSelection_RefsAreReactive tests that all refs are properly reactive
func TestUseSelection_RefsAreReactive(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c"}
	selection := UseSelection(ctx, items)

	// Verify all refs exist
	assert.NotNil(t, selection.SelectedIndex, "SelectedIndex ref should exist")
	assert.NotNil(t, selection.SelectedItem, "SelectedItem computed should exist")
	assert.NotNil(t, selection.SelectedIndices, "SelectedIndices ref should exist")
	assert.NotNil(t, selection.Items, "Items ref should exist")

	// Verify initial values
	assert.Equal(t, 0, selection.SelectedIndex.GetTyped(), "Initial index")
	assert.Equal(t, "a", selection.SelectedItem.Get(), "Initial item")
	assert.Equal(t, 3, len(selection.Items.GetTyped()), "Items count")
}

// TestUseSelection_MultipleOperations tests a sequence of operations
func TestUseSelection_MultipleOperations(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c", "d", "e"}
	selection := UseSelection(ctx, items, WithWrap(true))

	// Sequence of operations
	selection.SelectNext()
	assert.Equal(t, 1, selection.SelectedIndex.GetTyped())

	selection.SelectNext()
	selection.SelectNext()
	assert.Equal(t, 3, selection.SelectedIndex.GetTyped())

	selection.SelectPrevious()
	assert.Equal(t, 2, selection.SelectedIndex.GetTyped())

	selection.Select(4)
	assert.Equal(t, 4, selection.SelectedIndex.GetTyped())

	// Wrap around
	selection.SelectNext()
	assert.Equal(t, 0, selection.SelectedIndex.GetTyped())

	selection.SelectPrevious()
	assert.Equal(t, 4, selection.SelectedIndex.GetTyped())
}

// TestUseSelection_ToggleSelectionOutOfBounds tests toggle on invalid index
func TestUseSelection_ToggleSelectionOutOfBounds(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c"}
	selection := UseSelection(ctx, items, WithMultiSelect(true))

	// Toggle on invalid indices should be no-op
	selection.ToggleSelection(-1)
	assert.Equal(t, 0, len(selection.SelectedIndices.GetTyped()), "Negative index should be no-op")

	selection.ToggleSelection(10)
	assert.Equal(t, 0, len(selection.SelectedIndices.GetTyped()), "Out of bounds index should be no-op")
}

// TestUseSelection_SetItemsClearsMultiSelect tests that SetItems clears multi-select
func TestUseSelection_SetItemsClearsMultiSelect(t *testing.T) {
	ctx := createTestContext()
	items := []string{"a", "b", "c", "d"}
	selection := UseSelection(ctx, items, WithMultiSelect(true))

	// Select multiple
	selection.ToggleSelection(0)
	selection.ToggleSelection(2)
	assert.Equal(t, 2, len(selection.SelectedIndices.GetTyped()))

	// Set new items
	selection.SetItems([]string{"x", "y"})

	// Multi-select should be cleared
	assert.Equal(t, 0, len(selection.SelectedIndices.GetTyped()), "Multi-select should be cleared")
}
