package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseHistory_InitialStateSetCorrectly tests that initial state is set correctly
func TestUseHistory_InitialStateSetCorrectly(t *testing.T) {
	tests := []struct {
		name    string
		initial int
		maxSize int
	}{
		{"initial zero", 0, 10},
		{"initial positive", 42, 10},
		{"initial negative", -10, 10},
		{"max size 1", 5, 1},
		{"max size 100", 5, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			history := UseHistory(ctx, tt.initial, tt.maxSize)

			assert.NotNil(t, history, "UseHistory should return non-nil")
			assert.NotNil(t, history.Current, "Current should not be nil")
			assert.NotNil(t, history.CanUndo, "CanUndo should not be nil")
			assert.NotNil(t, history.CanRedo, "CanRedo should not be nil")
			assert.Equal(t, tt.initial, history.Current.GetTyped(),
				"Initial value should be %d", tt.initial)
		})
	}
}

// TestUseHistory_PushAddsToHistory tests that Push adds to history
func TestUseHistory_PushAddsToHistory(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		pushVals []int
		expected int
	}{
		{"push one value", 0, []int{10}, 10},
		{"push multiple values", 0, []int{10, 20, 30}, 30},
		{"push same value", 5, []int{5, 5, 5}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			history := UseHistory(ctx, tt.initial, 10)

			for _, val := range tt.pushVals {
				history.Push(val)
			}

			assert.Equal(t, tt.expected, history.Current.GetTyped(),
				"Current should be %d after pushes", tt.expected)
		})
	}
}

// TestUseHistory_UndoRevertsState tests that Undo reverts state
func TestUseHistory_UndoRevertsState(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Push some values
	history.Push(10)
	history.Push(20)
	history.Push(30)

	// Current should be 30
	assert.Equal(t, 30, history.Current.GetTyped())

	// Undo once
	history.Undo()
	assert.Equal(t, 20, history.Current.GetTyped(), "After first undo, should be 20")

	// Undo again
	history.Undo()
	assert.Equal(t, 10, history.Current.GetTyped(), "After second undo, should be 10")

	// Undo again
	history.Undo()
	assert.Equal(t, 0, history.Current.GetTyped(), "After third undo, should be initial 0")
}

// TestUseHistory_RedoRestoresState tests that Redo restores state
func TestUseHistory_RedoRestoresState(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Push some values
	history.Push(10)
	history.Push(20)
	history.Push(30)

	// Undo twice
	history.Undo()
	history.Undo()
	assert.Equal(t, 10, history.Current.GetTyped())

	// Redo once
	history.Redo()
	assert.Equal(t, 20, history.Current.GetTyped(), "After first redo, should be 20")

	// Redo again
	history.Redo()
	assert.Equal(t, 30, history.Current.GetTyped(), "After second redo, should be 30")
}

// TestUseHistory_CanUndoCanRedoComputedCorrectly tests CanUndo/CanRedo computed values
func TestUseHistory_CanUndoCanRedoComputedCorrectly(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Initially, cannot undo or redo
	assert.False(t, history.CanUndo.GetTyped(), "Initially cannot undo")
	assert.False(t, history.CanRedo.GetTyped(), "Initially cannot redo")

	// Push a value
	history.Push(10)
	assert.True(t, history.CanUndo.GetTyped(), "After push, can undo")
	assert.False(t, history.CanRedo.GetTyped(), "After push, cannot redo")

	// Undo
	history.Undo()
	assert.False(t, history.CanUndo.GetTyped(), "After undo to initial, cannot undo")
	assert.True(t, history.CanRedo.GetTyped(), "After undo, can redo")

	// Redo
	history.Redo()
	assert.True(t, history.CanUndo.GetTyped(), "After redo, can undo")
	assert.False(t, history.CanRedo.GetTyped(), "After redo to end, cannot redo")
}

// TestUseHistory_PushClearsRedoStack tests that Push clears redo stack
func TestUseHistory_PushClearsRedoStack(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Push values
	history.Push(10)
	history.Push(20)
	history.Push(30)

	// Undo twice
	history.Undo()
	history.Undo()
	assert.Equal(t, 10, history.Current.GetTyped())
	assert.True(t, history.CanRedo.GetTyped(), "Can redo after undo")

	// Push new value - should clear redo stack
	history.Push(100)
	assert.Equal(t, 100, history.Current.GetTyped())
	assert.False(t, history.CanRedo.GetTyped(), "Cannot redo after push (redo stack cleared)")

	// Verify we can still undo to the previous state
	history.Undo()
	assert.Equal(t, 10, history.Current.GetTyped(), "Undo should go to 10, not 20")
}

// TestUseHistory_MaxSizeEnforced tests that max size is enforced (drops oldest)
func TestUseHistory_MaxSizeEnforced(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 3) // Max 3 items in history

	// Push more than max size
	history.Push(10)
	history.Push(20)
	history.Push(30)
	history.Push(40) // Should drop oldest (0)

	assert.Equal(t, 40, history.Current.GetTyped())

	// Undo all the way
	history.Undo() // 30
	history.Undo() // 20
	history.Undo() // 10 (oldest kept)

	// Should not be able to undo further (0 was dropped)
	assert.Equal(t, 10, history.Current.GetTyped(), "Oldest should be 10 (0 was dropped)")
	assert.False(t, history.CanUndo.GetTyped(), "Cannot undo past max size limit")
}

// TestUseHistory_ClearEmptiesHistory tests that Clear empties history
func TestUseHistory_ClearEmptiesHistory(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Push some values
	history.Push(10)
	history.Push(20)
	history.Push(30)

	// Clear history
	history.Clear()

	// Current should remain at last value
	assert.Equal(t, 30, history.Current.GetTyped(), "Current should remain at last value after clear")

	// Cannot undo or redo
	assert.False(t, history.CanUndo.GetTyped(), "Cannot undo after clear")
	assert.False(t, history.CanRedo.GetTyped(), "Cannot redo after clear")
}

// TestUseHistory_UndoAtStartIsNoOp tests that Undo at start is no-op
func TestUseHistory_UndoAtStartIsNoOp(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 42, 10)

	// Try to undo when nothing to undo
	history.Undo()
	assert.Equal(t, 42, history.Current.GetTyped(), "Undo at start should be no-op")

	// Push and undo back to start
	history.Push(100)
	history.Undo()
	assert.Equal(t, 42, history.Current.GetTyped())

	// Try to undo again
	history.Undo()
	assert.Equal(t, 42, history.Current.GetTyped(), "Undo past start should be no-op")
}

// TestUseHistory_RedoAtEndIsNoOp tests that Redo at end is no-op
func TestUseHistory_RedoAtEndIsNoOp(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Try to redo when nothing to redo
	history.Redo()
	assert.Equal(t, 0, history.Current.GetTyped(), "Redo at end should be no-op")

	// Push value
	history.Push(10)
	assert.Equal(t, 10, history.Current.GetTyped())

	// Try to redo (already at end)
	history.Redo()
	assert.Equal(t, 10, history.Current.GetTyped(), "Redo at end should be no-op")
}

// TestUseHistory_WorksWithDifferentTypes tests generic type support
func TestUseHistory_WorksWithDifferentTypes(t *testing.T) {
	t.Run("string type", func(t *testing.T) {
		ctx := createTestContext()
		history := UseHistory(ctx, "initial", 10)

		history.Push("second")
		history.Push("third")
		assert.Equal(t, "third", history.Current.GetTyped())

		history.Undo()
		assert.Equal(t, "second", history.Current.GetTyped())

		history.Redo()
		assert.Equal(t, "third", history.Current.GetTyped())
	})

	t.Run("struct type", func(t *testing.T) {
		type State struct {
			Name  string
			Count int
		}

		ctx := createTestContext()
		history := UseHistory(ctx, State{Name: "initial", Count: 0}, 10)

		history.Push(State{Name: "second", Count: 1})
		history.Push(State{Name: "third", Count: 2})
		assert.Equal(t, State{Name: "third", Count: 2}, history.Current.GetTyped())

		history.Undo()
		assert.Equal(t, State{Name: "second", Count: 1}, history.Current.GetTyped())
	})

	t.Run("slice type", func(t *testing.T) {
		ctx := createTestContext()
		history := UseHistory(ctx, []int{1, 2, 3}, 10)

		history.Push([]int{4, 5, 6})
		assert.Equal(t, []int{4, 5, 6}, history.Current.GetTyped())

		history.Undo()
		assert.Equal(t, []int{1, 2, 3}, history.Current.GetTyped())
	})
}

// TestUseHistory_WorksWithCreateShared tests shared composable pattern
func TestUseHistory_WorksWithCreateShared(t *testing.T) {
	// Create shared history
	sharedHistory := CreateShared(func(ctx *bubbly.Context) *HistoryReturn[int] {
		return UseHistory(ctx, 0, 10)
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	history1 := sharedHistory(ctx1)
	history2 := sharedHistory(ctx2)

	// Both should be the same instance
	history1.Push(10)
	assert.Equal(t, 10, history1.Current.GetTyped())
	assert.Equal(t, 10, history2.Current.GetTyped(), "Shared history should reflect same state")
}

// TestUseHistory_CurrentIsReactive tests that Current ref is reactive
func TestUseHistory_CurrentIsReactive(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Track changes to Current
	changeCount := 0
	bubbly.Watch(history.Current, func(newVal, oldVal int) {
		changeCount++
	})

	// Push should trigger watcher
	history.Push(10)
	assert.Equal(t, 1, changeCount, "Push should trigger watcher")

	// Undo should trigger watcher
	history.Undo()
	assert.Equal(t, 2, changeCount, "Undo should trigger watcher")

	// Redo should trigger watcher
	history.Redo()
	assert.Equal(t, 3, changeCount, "Redo should trigger watcher")
}

// TestUseHistory_MultipleUndoRedo tests complex undo/redo sequences
func TestUseHistory_MultipleUndoRedo(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Build history
	history.Push(1)
	history.Push(2)
	history.Push(3)
	history.Push(4)
	history.Push(5)

	// Undo 3 times
	history.Undo()
	history.Undo()
	history.Undo()
	assert.Equal(t, 2, history.Current.GetTyped())

	// Redo 2 times
	history.Redo()
	history.Redo()
	assert.Equal(t, 4, history.Current.GetTyped())

	// Push new value (clears redo)
	history.Push(100)
	assert.Equal(t, 100, history.Current.GetTyped())
	assert.False(t, history.CanRedo.GetTyped())

	// Undo should go to 4
	history.Undo()
	assert.Equal(t, 4, history.Current.GetTyped())
}

// TestUseHistory_MaxSizeOne tests edge case with max size of 1
func TestUseHistory_MaxSizeOne(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 1)

	// Push a value
	history.Push(10)
	assert.Equal(t, 10, history.Current.GetTyped())

	// With max size 1, we can undo once (past has 1 entry: the initial value)
	assert.True(t, history.CanUndo.GetTyped(), "Can undo once with max size 1")

	// Undo to initial
	history.Undo()
	assert.Equal(t, 0, history.Current.GetTyped())
	assert.False(t, history.CanUndo.GetTyped(), "Cannot undo past initial")

	// Push two values - oldest should be dropped
	history.Push(10)
	history.Push(20)
	assert.Equal(t, 20, history.Current.GetTyped())

	// Can only undo once (to 10, not to 0)
	assert.True(t, history.CanUndo.GetTyped(), "Can undo once")
	history.Undo()
	assert.Equal(t, 10, history.Current.GetTyped())
	assert.False(t, history.CanUndo.GetTyped(), "Cannot undo further (0 was dropped)")
}

// TestUseHistory_EmptyHistoryAfterClear tests behavior after clear
func TestUseHistory_EmptyHistoryAfterClear(t *testing.T) {
	ctx := createTestContext()
	history := UseHistory(ctx, 0, 10)

	// Push and undo to create redo stack
	history.Push(10)
	history.Push(20)
	history.Undo()

	// Clear
	history.Clear()

	// Push new value
	history.Push(100)
	assert.Equal(t, 100, history.Current.GetTyped())
	assert.True(t, history.CanUndo.GetTyped(), "Can undo after push following clear")

	// Undo should go to value at clear time (10, since we undid 20)
	history.Undo()
	assert.Equal(t, 10, history.Current.GetTyped())
}
