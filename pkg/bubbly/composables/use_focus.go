package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// FocusReturn is the return value of UseFocus.
// It provides reactive focus state management for multi-pane TUI applications.
type FocusReturn[T comparable] struct {
	// Current is the currently focused pane.
	Current *bubbly.Ref[T]

	// order is the focus cycle order (internal).
	order []T
}

// IsFocused returns true if the given pane is currently focused.
//
// Example:
//
//	if focus.IsFocused(FocusMain) {
//	    // Render with focused styling
//	}
func (f *FocusReturn[T]) IsFocused(pane T) bool {
	return f.Current.GetTyped() == pane
}

// Focus sets focus to the specified pane.
// If the pane is not in the order, this is a no-op.
//
// Example:
//
//	focus.Focus(FocusMain)  // Jump directly to main pane
func (f *FocusReturn[T]) Focus(pane T) {
	// Only set if pane exists in order
	for _, p := range f.order {
		if p == pane {
			f.Current.Set(pane)
			return
		}
	}
	// No-op if pane not in order
}

// Next moves focus to the next pane in order.
// Wraps around to the first pane when at the end.
//
// Example:
//
//	ctx.On("nextFocus", func(_ interface{}) {
//	    focus.Next()
//	})
func (f *FocusReturn[T]) Next() {
	if len(f.order) <= 1 {
		return // Single item or empty - no change
	}

	current := f.Current.GetTyped()
	idx := f.findIndex(current)

	// Move to next, wrap if at end
	nextIdx := (idx + 1) % len(f.order)
	f.Current.Set(f.order[nextIdx])
}

// Previous moves focus to the previous pane in order.
// Wraps around to the last pane when at the beginning.
//
// Example:
//
//	ctx.On("prevFocus", func(_ interface{}) {
//	    focus.Previous()
//	})
func (f *FocusReturn[T]) Previous() {
	if len(f.order) <= 1 {
		return // Single item or empty - no change
	}

	current := f.Current.GetTyped()
	idx := f.findIndex(current)

	// Move to previous, wrap if at start
	prevIdx := idx - 1
	if prevIdx < 0 {
		prevIdx = len(f.order) - 1
	}
	f.Current.Set(f.order[prevIdx])
}

// findIndex returns the index of the given pane in the order.
// Returns 0 if not found (defaults to first item).
func (f *FocusReturn[T]) findIndex(pane T) int {
	for i, p := range f.order {
		if p == pane {
			return i
		}
	}
	return 0 // Default to first if not found
}

// UseFocus creates a focus management composable for multi-pane TUI applications.
// It tracks the currently focused pane and provides methods to cycle through
// a defined focus order.
//
// This composable is essential for building keyboard-navigable TUI applications
// with multiple interactive panes (e.g., sidebar, main content, footer).
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - initial: The initially focused pane
//   - order: The focus cycle order (must not be empty)
//
// Returns:
//   - *FocusReturn[T]: A struct containing reactive focus state and methods
//
// Panics:
//   - If order is empty
//
// Example:
//
//	type FocusPane int
//	const (
//	    FocusSidebar FocusPane = iota
//	    FocusMain
//	    FocusFooter
//	)
//
//	Setup(func(ctx *bubbly.Context) {
//	    focus := composables.UseFocus(ctx, FocusMain, []FocusPane{
//	        FocusSidebar, FocusMain, FocusFooter,
//	    })
//	    ctx.Expose("focus", focus)
//
//	    ctx.On("nextFocus", func(_ interface{}) {
//	        focus.Next()
//	    })
//	    ctx.On("prevFocus", func(_ interface{}) {
//	        focus.Previous()
//	    })
//	}).
//	WithKeyBinding("tab", "nextFocus", "Next pane").
//	WithKeyBinding("shift+tab", "prevFocus", "Previous pane")
//
// Integration with CreateShared:
//
//	var UseSharedFocus = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.FocusReturn[FocusPane] {
//	        return composables.UseFocus(ctx, FocusMain, []FocusPane{
//	            FocusSidebar, FocusMain, FocusFooter,
//	        })
//	    },
//	)
func UseFocus[T comparable](ctx *bubbly.Context, initial T, order []T) *FocusReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseFocus", time.Since(start))
	}()

	// Validate order is not empty
	if len(order) == 0 {
		panic("UseFocus: order must not be empty")
	}

	// Check if initial is in order, default to first if not
	initialInOrder := false
	for _, p := range order {
		if p == initial {
			initialInOrder = true
			break
		}
	}

	// If initial not in order, use first item
	actualInitial := initial
	if !initialInOrder {
		actualInitial = order[0]
	}

	// Create reactive ref with initial value
	current := bubbly.NewRef(actualInitial)

	return &FocusReturn[T]{
		Current: current,
		order:   order,
	}
}
