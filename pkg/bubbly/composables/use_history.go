package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// HistoryReturn is the return value of UseHistory.
// It provides undo/redo state management with a configurable maximum history size.
//
// The history is implemented as two stacks:
//   - past: States that can be undone (going back in time)
//   - future: States that can be redone (going forward after undo)
//
// When Push is called, the new state is added and the redo stack is cleared.
// When the history exceeds maxSize, the oldest entries are dropped.
type HistoryReturn[T any] struct {
	// Current is the current state value.
	// This is a reactive ref that can be watched for changes.
	Current *bubbly.Ref[T]

	// CanUndo indicates if undo is available.
	// This is a computed value that returns true when there are past states.
	CanUndo *bubbly.Computed[bool]

	// CanRedo indicates if redo is available.
	// This is a computed value that returns true when there are future states.
	CanRedo *bubbly.Computed[bool]

	// mu protects past and future slices
	mu sync.Mutex

	// past holds states that can be undone (stack, most recent at end)
	past []T

	// future holds states that can be redone (stack, most recent at end)
	future []T

	// maxSize is the maximum number of history entries
	maxSize int

	// pastLen and futureLen are refs for computed values to track
	pastLen   *bubbly.Ref[int]
	futureLen *bubbly.Ref[int]
}

// Push adds a new state to history (clears redo stack).
// The current state is pushed to the past stack, and the new value becomes current.
// If the history exceeds maxSize, the oldest entry is dropped.
//
// Example:
//
//	history := UseHistory(ctx, "initial", 10)
//	history.Push("second")  // Can now undo to "initial"
//	history.Push("third")   // Can now undo to "second"
func (h *HistoryReturn[T]) Push(value T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Get current value and add to past
	current := h.Current.GetTyped()
	h.past = append(h.past, current)

	// Enforce max size by dropping oldest entries
	// maxSize represents the maximum number of entries in the past stack
	// This allows undoing up to maxSize times
	for len(h.past) > h.maxSize {
		h.past = h.past[1:] // Drop oldest
	}

	// Clear future (redo stack) when pushing new value
	h.future = nil

	// Update reactive refs for computed values
	h.pastLen.Set(len(h.past))
	h.futureLen.Set(0)

	// Set new current value
	h.Current.Set(value)
}

// Undo reverts to previous state.
// The current state is pushed to the future stack (for redo), and the most recent
// past state becomes current. If there is no past state, this is a no-op.
//
// Example:
//
//	history := UseHistory(ctx, 0, 10)
//	history.Push(10)
//	history.Push(20)
//	history.Undo()  // Current is now 10
//	history.Undo()  // Current is now 0 (initial)
func (h *HistoryReturn[T]) Undo() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if we can undo
	if len(h.past) == 0 {
		return // No-op
	}

	// Get current value and push to future
	current := h.Current.GetTyped()
	h.future = append(h.future, current)

	// Pop from past and set as current
	lastIdx := len(h.past) - 1
	prev := h.past[lastIdx]
	h.past = h.past[:lastIdx]

	// Update reactive refs for computed values
	h.pastLen.Set(len(h.past))
	h.futureLen.Set(len(h.future))

	// Set previous value as current
	h.Current.Set(prev)
}

// Redo restores next state.
// The current state is pushed to the past stack, and the most recent future state
// becomes current. If there is no future state, this is a no-op.
//
// Example:
//
//	history := UseHistory(ctx, 0, 10)
//	history.Push(10)
//	history.Push(20)
//	history.Undo()  // Current is 10
//	history.Redo()  // Current is 20 again
func (h *HistoryReturn[T]) Redo() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if we can redo
	if len(h.future) == 0 {
		return // No-op
	}

	// Get current value and push to past
	current := h.Current.GetTyped()
	h.past = append(h.past, current)

	// Pop from future and set as current
	lastIdx := len(h.future) - 1
	next := h.future[lastIdx]
	h.future = h.future[:lastIdx]

	// Update reactive refs for computed values
	h.pastLen.Set(len(h.past))
	h.futureLen.Set(len(h.future))

	// Set next value as current
	h.Current.Set(next)
}

// Clear clears all history.
// The current value remains unchanged, but all past and future states are removed.
// After Clear, CanUndo and CanRedo will both return false.
//
// Example:
//
//	history := UseHistory(ctx, 0, 10)
//	history.Push(10)
//	history.Push(20)
//	history.Clear()
//	history.CanUndo.GetTyped()  // false
//	history.CanRedo.GetTyped()  // false
func (h *HistoryReturn[T]) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Clear both stacks
	h.past = nil
	h.future = nil

	// Update reactive refs for computed values
	h.pastLen.Set(0)
	h.futureLen.Set(0)
}

// UseHistory creates an undo/redo history composable.
// It provides a simple API for managing state history with configurable maximum size.
//
// This composable is useful for:
//   - Text editor undo/redo
//   - Form state history
//   - Drawing application undo
//   - Game state snapshots
//   - Any application requiring state time-travel
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - initial: The initial state value
//   - maxSize: Maximum number of history entries (including current)
//
// Returns:
//   - *HistoryReturn[T]: A struct containing the reactive current value and computed CanUndo/CanRedo
//
// Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    // Create history for text content
//	    history := composables.UseHistory(ctx, "", 50)
//	    ctx.Expose("history", history)
//
//	    // Save current state
//	    ctx.On("save", func(data interface{}) {
//	        text := data.(string)
//	        history.Push(text)
//	    })
//
//	    // Undo/redo handlers
//	    ctx.On("undo", func(_ interface{}) {
//	        history.Undo()
//	    })
//
//	    ctx.On("redo", func(_ interface{}) {
//	        history.Redo()
//	    })
//	}).
//	WithKeyBinding("ctrl+z", "undo", "Undo").
//	WithKeyBinding("ctrl+y", "redo", "Redo")
//
// Template usage:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    history := ctx.Get("history").(*composables.HistoryReturn[string])
//	    content := history.Current.GetTyped()
//
//	    undoStatus := "Undo: disabled"
//	    if history.CanUndo.GetTyped() {
//	        undoStatus = "Undo: available"
//	    }
//
//	    return fmt.Sprintf("%s\n%s", content, undoStatus)
//	})
//
// Integration with CreateShared:
//
//	var UseSharedHistory = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.HistoryReturn[AppState] {
//	        return composables.UseHistory(ctx, AppState{}, 100)
//	    },
//	)
func UseHistory[T any](ctx *bubbly.Context, initial T, maxSize int) *HistoryReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseHistory", time.Since(start))
	}()

	// Ensure maxSize is at least 1
	if maxSize < 1 {
		maxSize = 1
	}

	// Create reactive refs for tracking history lengths (for computed values)
	pastLen := bubbly.NewRef(0)
	futureLen := bubbly.NewRef(0)

	// Create current value ref
	current := bubbly.NewRef(initial)

	// Create computed values for CanUndo and CanRedo
	canUndo := bubbly.NewComputed(func() bool {
		return pastLen.GetTyped() > 0
	})

	canRedo := bubbly.NewComputed(func() bool {
		return futureLen.GetTyped() > 0
	})

	return &HistoryReturn[T]{
		Current:   current,
		CanUndo:   canUndo,
		CanRedo:   canRedo,
		past:      nil,
		future:    nil,
		maxSize:   maxSize,
		pastLen:   pastLen,
		futureLen: futureLen,
	}
}
