package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// FocusPane is a test type for focus management.
type FocusPane int

const (
	FocusSidebar FocusPane = iota
	FocusMain
	FocusFooter
	FocusHeader
)

// TestUseFocus_InitialFocusSetCorrectly tests that initial focus is set correctly
func TestUseFocus_InitialFocusSetCorrectly(t *testing.T) {
	tests := []struct {
		name    string
		initial FocusPane
		order   []FocusPane
	}{
		{"first in order", FocusSidebar, []FocusPane{FocusSidebar, FocusMain, FocusFooter}},
		{"middle in order", FocusMain, []FocusPane{FocusSidebar, FocusMain, FocusFooter}},
		{"last in order", FocusFooter, []FocusPane{FocusSidebar, FocusMain, FocusFooter}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			focus := UseFocus(ctx, tt.initial, tt.order)

			assert.NotNil(t, focus, "UseFocus should return non-nil")
			assert.NotNil(t, focus.Current, "Current should not be nil")
			assert.Equal(t, tt.initial, focus.Current.GetTyped(),
				"Initial focus should be %v", tt.initial)
		})
	}
}

// TestUseFocus_NextCyclesThroughOrder tests that Next() cycles through order
func TestUseFocus_NextCyclesThroughOrder(t *testing.T) {
	ctx := createTestContext()
	order := []FocusPane{FocusSidebar, FocusMain, FocusFooter}
	focus := UseFocus(ctx, FocusSidebar, order)

	// Initial state
	assert.Equal(t, FocusSidebar, focus.Current.GetTyped(), "Initial focus")

	// Next should move to Main
	focus.Next()
	assert.Equal(t, FocusMain, focus.Current.GetTyped(), "After first Next()")

	// Next should move to Footer
	focus.Next()
	assert.Equal(t, FocusFooter, focus.Current.GetTyped(), "After second Next()")

	// Next should wrap to Sidebar
	focus.Next()
	assert.Equal(t, FocusSidebar, focus.Current.GetTyped(), "After third Next() - should wrap")
}

// TestUseFocus_PreviousCyclesBackward tests that Previous() cycles backward
func TestUseFocus_PreviousCyclesBackward(t *testing.T) {
	ctx := createTestContext()
	order := []FocusPane{FocusSidebar, FocusMain, FocusFooter}
	focus := UseFocus(ctx, FocusSidebar, order)

	// Previous from first should wrap to last
	focus.Previous()
	assert.Equal(t, FocusFooter, focus.Current.GetTyped(), "Previous from first should wrap to last")

	// Previous should move to Main
	focus.Previous()
	assert.Equal(t, FocusMain, focus.Current.GetTyped(), "After second Previous()")

	// Previous should move to Sidebar
	focus.Previous()
	assert.Equal(t, FocusSidebar, focus.Current.GetTyped(), "After third Previous()")
}

// TestUseFocus_FocusSetsSpecificPane tests that Focus() sets a specific pane
func TestUseFocus_FocusSetsSpecificPane(t *testing.T) {
	ctx := createTestContext()
	order := []FocusPane{FocusSidebar, FocusMain, FocusFooter}
	focus := UseFocus(ctx, FocusSidebar, order)

	// Focus on Footer directly
	focus.Focus(FocusFooter)
	assert.Equal(t, FocusFooter, focus.Current.GetTyped(), "Focus should set to Footer")

	// Focus on Main
	focus.Focus(FocusMain)
	assert.Equal(t, FocusMain, focus.Current.GetTyped(), "Focus should set to Main")

	// Focus on Sidebar
	focus.Focus(FocusSidebar)
	assert.Equal(t, FocusSidebar, focus.Current.GetTyped(), "Focus should set to Sidebar")
}

// TestUseFocus_FocusOnNonExistentPaneIsNoOp tests that Focus() on non-existent pane is no-op
func TestUseFocus_FocusOnNonExistentPaneIsNoOp(t *testing.T) {
	ctx := createTestContext()
	order := []FocusPane{FocusSidebar, FocusMain, FocusFooter}
	focus := UseFocus(ctx, FocusSidebar, order)

	// Focus on Header (not in order) should be no-op
	focus.Focus(FocusHeader)
	assert.Equal(t, FocusSidebar, focus.Current.GetTyped(),
		"Focus on non-existent pane should not change current")
}

// TestUseFocus_IsFocusedReturnsCorrectValue tests that IsFocused() returns correct value
func TestUseFocus_IsFocusedReturnsCorrectValue(t *testing.T) {
	ctx := createTestContext()
	order := []FocusPane{FocusSidebar, FocusMain, FocusFooter}
	focus := UseFocus(ctx, FocusMain, order)

	assert.True(t, focus.IsFocused(FocusMain), "Main should be focused")
	assert.False(t, focus.IsFocused(FocusSidebar), "Sidebar should not be focused")
	assert.False(t, focus.IsFocused(FocusFooter), "Footer should not be focused")

	// Change focus and verify
	focus.Next()
	assert.True(t, focus.IsFocused(FocusFooter), "Footer should now be focused")
	assert.False(t, focus.IsFocused(FocusMain), "Main should no longer be focused")
}

// TestUseFocus_EmptyOrderPanics tests that empty order panics
func TestUseFocus_EmptyOrderPanics(t *testing.T) {
	ctx := createTestContext()

	assert.Panics(t, func() {
		UseFocus(ctx, FocusSidebar, []FocusPane{})
	}, "Empty order should panic")
}

// TestUseFocus_SingleItemOrderStaysFocused tests that single item order stays focused
func TestUseFocus_SingleItemOrderStaysFocused(t *testing.T) {
	ctx := createTestContext()
	order := []FocusPane{FocusMain}
	focus := UseFocus(ctx, FocusMain, order)

	// Initial state
	assert.Equal(t, FocusMain, focus.Current.GetTyped(), "Initial focus")

	// Next should stay on Main
	focus.Next()
	assert.Equal(t, FocusMain, focus.Current.GetTyped(), "Next should stay on single item")

	// Previous should stay on Main
	focus.Previous()
	assert.Equal(t, FocusMain, focus.Current.GetTyped(), "Previous should stay on single item")
}

// TestUseFocus_WorksWithCreateShared tests shared composable pattern
func TestUseFocus_WorksWithCreateShared(t *testing.T) {
	// Create shared instance
	sharedFocus := CreateShared(func(ctx *bubbly.Context) *FocusReturn[FocusPane] {
		return UseFocus(ctx, FocusSidebar, []FocusPane{FocusSidebar, FocusMain, FocusFooter})
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	focus1 := sharedFocus(ctx1)
	focus2 := sharedFocus(ctx2)

	// Both should be the same instance
	focus1.Next()

	assert.Equal(t, FocusMain, focus2.Current.GetTyped(),
		"Shared instance should have same focus state")
}

// TestUseFocus_WithStringType tests generic type with strings
func TestUseFocus_WithStringType(t *testing.T) {
	ctx := createTestContext()
	order := []string{"sidebar", "main", "footer"}
	focus := UseFocus(ctx, "sidebar", order)

	assert.Equal(t, "sidebar", focus.Current.GetTyped(), "Initial focus")

	focus.Next()
	assert.Equal(t, "main", focus.Current.GetTyped(), "After Next()")

	assert.True(t, focus.IsFocused("main"), "main should be focused")
	assert.False(t, focus.IsFocused("sidebar"), "sidebar should not be focused")
}

// TestUseFocus_NextAndPreviousAlternate tests alternating Next and Previous
func TestUseFocus_NextAndPreviousAlternate(t *testing.T) {
	ctx := createTestContext()
	order := []FocusPane{FocusSidebar, FocusMain, FocusFooter}
	focus := UseFocus(ctx, FocusMain, order)

	// Start at Main
	assert.Equal(t, FocusMain, focus.Current.GetTyped())

	// Next to Footer
	focus.Next()
	assert.Equal(t, FocusFooter, focus.Current.GetTyped())

	// Previous back to Main
	focus.Previous()
	assert.Equal(t, FocusMain, focus.Current.GetTyped())

	// Previous to Sidebar
	focus.Previous()
	assert.Equal(t, FocusSidebar, focus.Current.GetTyped())

	// Next back to Main
	focus.Next()
	assert.Equal(t, FocusMain, focus.Current.GetTyped())
}

// TestUseFocus_InitialNotInOrder tests behavior when initial is not in order
func TestUseFocus_InitialNotInOrder(t *testing.T) {
	ctx := createTestContext()
	order := []FocusPane{FocusSidebar, FocusMain, FocusFooter}

	// Header is not in order - should still set it but Next/Previous may behave unexpectedly
	// Per requirements, this is an edge case - we'll set to first in order
	focus := UseFocus(ctx, FocusHeader, order)

	// Should default to first item in order when initial is not found
	assert.Equal(t, FocusSidebar, focus.Current.GetTyped(),
		"When initial not in order, should default to first item")
}
