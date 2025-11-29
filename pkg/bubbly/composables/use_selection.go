package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// selectionConfig holds configuration for UseSelection.
type selectionConfig struct {
	wrap        bool
	multiSelect bool
}

// SelectionOption configures UseSelection.
type SelectionOption func(*selectionConfig)

// WithWrap enables wrapping at list boundaries.
// When enabled, navigating past the last item wraps to the first,
// and navigating before the first item wraps to the last.
//
// Example:
//
//	selection := UseSelection(ctx, items, WithWrap(true))
//	selection.Select(len(items) - 1)  // At last item
//	selection.SelectNext()             // Wraps to index 0
func WithWrap(wrap bool) SelectionOption {
	return func(c *selectionConfig) {
		c.wrap = wrap
	}
}

// WithMultiSelect enables multi-selection mode.
// In this mode, ToggleSelection can be used to select/deselect multiple items,
// and IsSelected checks the SelectedIndices list.
//
// Example:
//
//	selection := UseSelection(ctx, items, WithMultiSelect(true))
//	selection.ToggleSelection(0)  // Select first item
//	selection.ToggleSelection(2)  // Select third item
//	selection.IsSelected(0)       // true
//	selection.IsSelected(1)       // false
func WithMultiSelect(multi bool) SelectionOption {
	return func(c *selectionConfig) {
		c.multiSelect = multi
	}
}

// SelectionReturn is the return value of UseSelection.
// It provides reactive selection state management for lists and tables in TUI applications.
type SelectionReturn[T any] struct {
	// SelectedIndex is the currently selected index (cursor position).
	// Returns -1 for empty lists.
	SelectedIndex *bubbly.Ref[int]

	// SelectedItem is the currently selected item (computed from index and items).
	// Returns zero value for empty lists or invalid index.
	SelectedItem *bubbly.Computed[T]

	// SelectedIndices is for multi-select mode.
	// Contains all selected indices when WithMultiSelect(true) is used.
	SelectedIndices *bubbly.Ref[[]int]

	// Items is the list of selectable items.
	Items *bubbly.Ref[[]T]

	// config holds the selection configuration.
	config selectionConfig
}

// Select sets the selection to a specific index.
// The index is clamped to valid range [0, len(items)-1].
// For empty lists, this is a no-op.
//
// Example:
//
//	selection.Select(5)  // Jump to index 5
func (s *SelectionReturn[T]) Select(index int) {
	items := s.Items.GetTyped()
	if len(items) == 0 {
		return // No-op for empty list
	}

	// Clamp to valid range
	if index < 0 {
		index = 0
	}
	if index >= len(items) {
		index = len(items) - 1
	}

	s.SelectedIndex.Set(index)
}

// SelectNext moves selection to the next item.
// Without wrap: stops at the last item.
// With wrap: cycles to the first item.
//
// Example:
//
//	ctx.On("selectNext", func(_ interface{}) {
//	    selection.SelectNext()
//	})
func (s *SelectionReturn[T]) SelectNext() {
	items := s.Items.GetTyped()
	if len(items) == 0 {
		return // No-op for empty list
	}

	current := s.SelectedIndex.GetTyped()
	if current < 0 {
		current = 0
	}

	next := current + 1
	if next >= len(items) {
		if s.config.wrap {
			next = 0
		} else {
			next = len(items) - 1
		}
	}

	s.SelectedIndex.Set(next)
}

// SelectPrevious moves selection to the previous item.
// Without wrap: stops at the first item.
// With wrap: cycles to the last item.
//
// Example:
//
//	ctx.On("selectPrevious", func(_ interface{}) {
//	    selection.SelectPrevious()
//	})
func (s *SelectionReturn[T]) SelectPrevious() {
	items := s.Items.GetTyped()
	if len(items) == 0 {
		return // No-op for empty list
	}

	current := s.SelectedIndex.GetTyped()
	if current < 0 {
		current = 0
	}

	prev := current - 1
	if prev < 0 {
		if s.config.wrap {
			prev = len(items) - 1
		} else {
			prev = 0
		}
	}

	s.SelectedIndex.Set(prev)
}

// IsSelected returns true if the index is selected.
// In single-select mode, checks against SelectedIndex.
// In multi-select mode, checks against SelectedIndices.
//
// Example:
//
//	if selection.IsSelected(i) {
//	    // Render with selected styling
//	}
func (s *SelectionReturn[T]) IsSelected(index int) bool {
	items := s.Items.GetTyped()
	if len(items) == 0 || index < 0 || index >= len(items) {
		return false
	}

	if s.config.multiSelect {
		// Check in SelectedIndices
		for _, idx := range s.SelectedIndices.GetTyped() {
			if idx == index {
				return true
			}
		}
		return false
	}

	// Single-select mode
	return s.SelectedIndex.GetTyped() == index
}

// ToggleSelection toggles selection at index (multi-select mode).
// If the index is already selected, it is deselected.
// If the index is not selected, it is added to the selection.
// This is a no-op in single-select mode or for invalid indices.
//
// Example:
//
//	ctx.On("toggle", func(_ interface{}) {
//	    selection.ToggleSelection(selection.SelectedIndex.GetTyped())
//	})
func (s *SelectionReturn[T]) ToggleSelection(index int) {
	if !s.config.multiSelect {
		return // No-op in single-select mode
	}

	items := s.Items.GetTyped()
	if len(items) == 0 || index < 0 || index >= len(items) {
		return // Invalid index
	}

	indices := s.SelectedIndices.GetTyped()

	// Check if already selected
	for i, idx := range indices {
		if idx == index {
			// Remove from selection
			newIndices := make([]int, 0, len(indices)-1)
			newIndices = append(newIndices, indices[:i]...)
			newIndices = append(newIndices, indices[i+1:]...)
			s.SelectedIndices.Set(newIndices)
			return
		}
	}

	// Add to selection
	newIndices := make([]int, len(indices)+1)
	copy(newIndices, indices)
	newIndices[len(indices)] = index
	s.SelectedIndices.Set(newIndices)
}

// ClearSelection clears all selections.
// In single-select mode, resets to index 0.
// In multi-select mode, clears the SelectedIndices list.
//
// Example:
//
//	ctx.On("clearSelection", func(_ interface{}) {
//	    selection.ClearSelection()
//	})
func (s *SelectionReturn[T]) ClearSelection() {
	items := s.Items.GetTyped()

	if s.config.multiSelect {
		s.SelectedIndices.Set([]int{})
	}

	// Reset cursor to 0 (or -1 for empty list)
	if len(items) == 0 {
		s.SelectedIndex.Set(-1)
	} else {
		s.SelectedIndex.Set(0)
	}
}

// SetItems updates the items list and adjusts selection.
// If the current selection is beyond the new list bounds, it is clamped.
// Multi-select indices are cleared when items change.
//
// Example:
//
//	// After filtering items
//	selection.SetItems(filteredItems)
func (s *SelectionReturn[T]) SetItems(items []T) {
	s.Items.Set(items)

	// Clear multi-select
	if s.config.multiSelect {
		s.SelectedIndices.Set([]int{})
	}

	// Adjust selection
	if len(items) == 0 {
		s.SelectedIndex.Set(-1)
	} else {
		current := s.SelectedIndex.GetTyped()
		if current < 0 {
			// Was empty, now has items
			s.SelectedIndex.Set(0)
		} else if current >= len(items) {
			// Clamp to last item
			s.SelectedIndex.Set(len(items) - 1)
		}
	}
}

// UseSelection creates a selection management composable for lists and tables.
// It tracks the selected index and provides methods for navigation and selection.
//
// This composable is essential for building interactive lists, tables, and menus
// in TUI applications.
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - items: The list of selectable items
//   - opts: Optional configuration (WithWrap, WithMultiSelect)
//
// Returns:
//   - *SelectionReturn[T]: A struct containing reactive selection state and methods
//
// Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    items := []Item{...}
//
//	    selection := composables.UseSelection(ctx, items,
//	        composables.WithWrap(true))
//	    ctx.Expose("selection", selection)
//
//	    ctx.On("selectNext", func(_ interface{}) {
//	        selection.SelectNext()
//	    })
//	    ctx.On("selectPrevious", func(_ interface{}) {
//	        selection.SelectPrevious()
//	    })
//	    ctx.On("select", func(_ interface{}) {
//	        item := selection.SelectedItem.Get()
//	        // Handle selection
//	    })
//	}).
//	WithMultiKeyBindings("selectNext", "Next item", "down", "j").
//	WithMultiKeyBindings("selectPrevious", "Previous item", "up", "k").
//	WithKeyBinding("enter", "select", "Select item")
//
// Integration with CreateShared:
//
//	var UseSharedSelection = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.SelectionReturn[Item] {
//	        return composables.UseSelection(ctx, items)
//	    },
//	)
//
// Multi-select mode:
//
//	selection := composables.UseSelection(ctx, items,
//	    composables.WithMultiSelect(true))
//
//	ctx.On("toggle", func(_ interface{}) {
//	    selection.ToggleSelection(selection.SelectedIndex.GetTyped())
//	})
func UseSelection[T any](ctx *bubbly.Context, items []T, opts ...SelectionOption) *SelectionReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseSelection", time.Since(start))
	}()

	// Apply options
	config := selectionConfig{
		wrap:        false,
		multiSelect: false,
	}
	for _, opt := range opts {
		opt(&config)
	}

	// Determine initial selection
	initialIndex := 0
	if len(items) == 0 {
		initialIndex = -1
	}

	// Create reactive refs
	selectedIndex := bubbly.NewRef(initialIndex)
	selectedIndices := bubbly.NewRef([]int{})
	itemsRef := bubbly.NewRef(items)

	// Create computed for selected item
	selectedItem := bubbly.NewComputed(func() T {
		var zero T
		idx := selectedIndex.GetTyped()
		currentItems := itemsRef.GetTyped()

		if idx < 0 || idx >= len(currentItems) {
			return zero
		}

		return currentItems[idx]
	})

	return &SelectionReturn[T]{
		SelectedIndex:   selectedIndex,
		SelectedItem:    selectedItem,
		SelectedIndices: selectedIndices,
		Items:           itemsRef,
		config:          config,
	}
}
