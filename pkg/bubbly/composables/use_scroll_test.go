package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseScroll_InitialOffsetIsZero tests that initial offset is 0
func TestUseScroll_InitialOffsetIsZero(t *testing.T) {
	ctx := createTestContext()
	scroll := UseScroll(ctx, 100, 10)

	assert.NotNil(t, scroll, "UseScroll should return non-nil")
	assert.NotNil(t, scroll.Offset, "Offset should not be nil")
	assert.Equal(t, 0, scroll.Offset.GetTyped(), "Initial offset should be 0")
}

// TestUseScroll_ScrollUpRespectsBounds tests that ScrollUp respects bounds
func TestUseScroll_ScrollUpRespectsBounds(t *testing.T) {
	tests := []struct {
		name           string
		totalItems     int
		visibleCount   int
		initialOffset  int
		expectedOffset int
	}{
		{"scroll up from middle", 100, 10, 50, 49},
		{"scroll up from 1", 100, 10, 1, 0},
		{"scroll up from 0 stays at 0", 100, 10, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			scroll := UseScroll(ctx, tt.totalItems, tt.visibleCount)

			// Set initial offset
			scroll.ScrollTo(tt.initialOffset)

			// Scroll up
			scroll.ScrollUp()

			assert.Equal(t, tt.expectedOffset, scroll.Offset.GetTyped(),
				"Offset should be %d after ScrollUp", tt.expectedOffset)
		})
	}
}

// TestUseScroll_ScrollDownRespectsBounds tests that ScrollDown respects bounds
func TestUseScroll_ScrollDownRespectsBounds(t *testing.T) {
	tests := []struct {
		name           string
		totalItems     int
		visibleCount   int
		initialOffset  int
		expectedOffset int
	}{
		{"scroll down from 0", 100, 10, 0, 1},
		{"scroll down from middle", 100, 10, 50, 51},
		{"scroll down at max stays at max", 100, 10, 90, 90},
		{"scroll down near max", 100, 10, 89, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			scroll := UseScroll(ctx, tt.totalItems, tt.visibleCount)

			// Set initial offset
			scroll.ScrollTo(tt.initialOffset)

			// Scroll down
			scroll.ScrollDown()

			assert.Equal(t, tt.expectedOffset, scroll.Offset.GetTyped(),
				"Offset should be %d after ScrollDown", tt.expectedOffset)
		})
	}
}

// TestUseScroll_ScrollToClampsToValidRange tests that ScrollTo clamps to valid range
func TestUseScroll_ScrollToClampsToValidRange(t *testing.T) {
	tests := []struct {
		name           string
		totalItems     int
		visibleCount   int
		scrollTo       int
		expectedOffset int
	}{
		{"scroll to valid position", 100, 10, 50, 50},
		{"scroll to 0", 100, 10, 0, 0},
		{"scroll to max", 100, 10, 90, 90},
		{"scroll to negative clamps to 0", 100, 10, -10, 0},
		{"scroll beyond max clamps to max", 100, 10, 200, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			scroll := UseScroll(ctx, tt.totalItems, tt.visibleCount)

			scroll.ScrollTo(tt.scrollTo)

			assert.Equal(t, tt.expectedOffset, scroll.Offset.GetTyped(),
				"Offset should be clamped to %d", tt.expectedOffset)
		})
	}
}

// TestUseScroll_ScrollToTopAndBottom tests ScrollToTop and ScrollToBottom
func TestUseScroll_ScrollToTopAndBottom(t *testing.T) {
	ctx := createTestContext()
	scroll := UseScroll(ctx, 100, 10)

	// Start at middle
	scroll.ScrollTo(50)
	assert.Equal(t, 50, scroll.Offset.GetTyped(), "Should be at 50")

	// Scroll to top
	scroll.ScrollToTop()
	assert.Equal(t, 0, scroll.Offset.GetTyped(), "ScrollToTop should set offset to 0")

	// Scroll to bottom
	scroll.ScrollToBottom()
	assert.Equal(t, 90, scroll.Offset.GetTyped(), "ScrollToBottom should set offset to max")
}

// TestUseScroll_PageUpAndPageDown tests PageUp and PageDown
func TestUseScroll_PageUpAndPageDown(t *testing.T) {
	tests := []struct {
		name           string
		totalItems     int
		visibleCount   int
		initialOffset  int
		action         string
		expectedOffset int
	}{
		{"page down from 0", 100, 10, 0, "down", 10},
		{"page down from middle", 100, 10, 50, "down", 60},
		{"page down near end clamps to max", 100, 10, 85, "down", 90},
		{"page up from middle", 100, 10, 50, "up", 40},
		{"page up from 5 clamps to 0", 100, 10, 5, "up", 0},
		{"page up from 0 stays at 0", 100, 10, 0, "up", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			scroll := UseScroll(ctx, tt.totalItems, tt.visibleCount)

			scroll.ScrollTo(tt.initialOffset)

			if tt.action == "down" {
				scroll.PageDown()
			} else {
				scroll.PageUp()
			}

			assert.Equal(t, tt.expectedOffset, scroll.Offset.GetTyped(),
				"Offset should be %d after %s", tt.expectedOffset, tt.action)
		})
	}
}

// TestUseScroll_IsAtTopAndBottom tests IsAtTop and IsAtBottom
func TestUseScroll_IsAtTopAndBottom(t *testing.T) {
	ctx := createTestContext()
	scroll := UseScroll(ctx, 100, 10)

	// At top
	assert.True(t, scroll.IsAtTop(), "Should be at top initially")
	assert.False(t, scroll.IsAtBottom(), "Should not be at bottom initially")

	// Move to middle
	scroll.ScrollTo(50)
	assert.False(t, scroll.IsAtTop(), "Should not be at top when at 50")
	assert.False(t, scroll.IsAtBottom(), "Should not be at bottom when at 50")

	// Move to bottom
	scroll.ScrollToBottom()
	assert.False(t, scroll.IsAtTop(), "Should not be at top when at bottom")
	assert.True(t, scroll.IsAtBottom(), "Should be at bottom")
}

// TestUseScroll_SetTotalItemsRecalculatesMaxOffset tests SetTotalItems
func TestUseScroll_SetTotalItemsRecalculatesMaxOffset(t *testing.T) {
	ctx := createTestContext()
	scroll := UseScroll(ctx, 100, 10)

	// Initial max offset should be 90 (100 - 10)
	assert.Equal(t, 90, scroll.MaxOffset.GetTyped(), "Initial max offset should be 90")

	// Change total items
	scroll.SetTotalItems(50)
	assert.Equal(t, 40, scroll.MaxOffset.GetTyped(), "Max offset should be 40 after SetTotalItems(50)")
	assert.Equal(t, 50, scroll.TotalItems.GetTyped(), "TotalItems should be 50")

	// If current offset is beyond new max, it should be clamped
	scroll.ScrollTo(50) // This should clamp to 40
	assert.Equal(t, 40, scroll.Offset.GetTyped(), "Offset should be clamped to new max")
}

// TestUseScroll_SetVisibleCountRecalculatesMaxOffset tests SetVisibleCount
func TestUseScroll_SetVisibleCountRecalculatesMaxOffset(t *testing.T) {
	ctx := createTestContext()
	scroll := UseScroll(ctx, 100, 10)

	// Initial max offset should be 90 (100 - 10)
	assert.Equal(t, 90, scroll.MaxOffset.GetTyped(), "Initial max offset should be 90")

	// Change visible count
	scroll.SetVisibleCount(20)
	assert.Equal(t, 80, scroll.MaxOffset.GetTyped(), "Max offset should be 80 after SetVisibleCount(20)")
	assert.Equal(t, 20, scroll.VisibleCount.GetTyped(), "VisibleCount should be 20")
}

// TestUseScroll_EmptyListHandled tests handling of empty list (0 items)
func TestUseScroll_EmptyListHandled(t *testing.T) {
	ctx := createTestContext()
	scroll := UseScroll(ctx, 0, 10)

	assert.Equal(t, 0, scroll.Offset.GetTyped(), "Offset should be 0 for empty list")
	assert.Equal(t, 0, scroll.MaxOffset.GetTyped(), "MaxOffset should be 0 for empty list")
	assert.Equal(t, 0, scroll.TotalItems.GetTyped(), "TotalItems should be 0")

	// Operations should be safe on empty list
	scroll.ScrollDown()
	assert.Equal(t, 0, scroll.Offset.GetTyped(), "ScrollDown on empty list should stay at 0")

	scroll.ScrollUp()
	assert.Equal(t, 0, scroll.Offset.GetTyped(), "ScrollUp on empty list should stay at 0")

	scroll.PageDown()
	assert.Equal(t, 0, scroll.Offset.GetTyped(), "PageDown on empty list should stay at 0")

	scroll.PageUp()
	assert.Equal(t, 0, scroll.Offset.GetTyped(), "PageUp on empty list should stay at 0")

	assert.True(t, scroll.IsAtTop(), "Empty list should be at top")
	assert.True(t, scroll.IsAtBottom(), "Empty list should be at bottom")
}

// TestUseScroll_VisibleCountGreaterThanTotal tests when visible count >= total items
func TestUseScroll_VisibleCountGreaterThanTotal(t *testing.T) {
	tests := []struct {
		name         string
		totalItems   int
		visibleCount int
	}{
		{"visible equals total", 10, 10},
		{"visible greater than total", 10, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			scroll := UseScroll(ctx, tt.totalItems, tt.visibleCount)

			assert.Equal(t, 0, scroll.Offset.GetTyped(), "Offset should be 0")
			assert.Equal(t, 0, scroll.MaxOffset.GetTyped(), "MaxOffset should be 0 when visible >= total")

			// All items visible, so at top and bottom
			assert.True(t, scroll.IsAtTop(), "Should be at top")
			assert.True(t, scroll.IsAtBottom(), "Should be at bottom")

			// Scrolling should have no effect
			scroll.ScrollDown()
			assert.Equal(t, 0, scroll.Offset.GetTyped(), "ScrollDown should have no effect")
		})
	}
}

// TestUseScroll_WorksWithCreateShared tests shared composable pattern
func TestUseScroll_WorksWithCreateShared(t *testing.T) {
	sharedScroll := CreateShared(func(ctx *bubbly.Context) *ScrollReturn {
		return UseScroll(ctx, 100, 10)
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	scroll1 := sharedScroll(ctx1)
	scroll2 := sharedScroll(ctx2)

	// Both should be the same instance
	scroll1.ScrollTo(50)

	assert.Equal(t, 50, scroll2.Offset.GetTyped(),
		"Shared instance should have same scroll state")
}

// TestUseScroll_RefsAreReactive tests that all refs are properly reactive
func TestUseScroll_RefsAreReactive(t *testing.T) {
	ctx := createTestContext()
	scroll := UseScroll(ctx, 100, 10)

	// Verify all refs exist and are initialized
	assert.NotNil(t, scroll.Offset, "Offset ref should exist")
	assert.NotNil(t, scroll.MaxOffset, "MaxOffset ref should exist")
	assert.NotNil(t, scroll.VisibleCount, "VisibleCount ref should exist")
	assert.NotNil(t, scroll.TotalItems, "TotalItems ref should exist")

	// Verify initial values
	assert.Equal(t, 0, scroll.Offset.GetTyped(), "Initial offset")
	assert.Equal(t, 90, scroll.MaxOffset.GetTyped(), "Initial max offset")
	assert.Equal(t, 10, scroll.VisibleCount.GetTyped(), "Initial visible count")
	assert.Equal(t, 100, scroll.TotalItems.GetTyped(), "Initial total items")
}

// TestUseScroll_MultipleScrollOperations tests a sequence of scroll operations
func TestUseScroll_MultipleScrollOperations(t *testing.T) {
	ctx := createTestContext()
	scroll := UseScroll(ctx, 100, 10)

	// Sequence of operations
	scroll.ScrollDown()
	assert.Equal(t, 1, scroll.Offset.GetTyped())

	scroll.ScrollDown()
	scroll.ScrollDown()
	assert.Equal(t, 3, scroll.Offset.GetTyped())

	scroll.PageDown()
	assert.Equal(t, 13, scroll.Offset.GetTyped())

	scroll.ScrollUp()
	assert.Equal(t, 12, scroll.Offset.GetTyped())

	scroll.PageUp()
	assert.Equal(t, 2, scroll.Offset.GetTyped())

	scroll.ScrollToTop()
	assert.Equal(t, 0, scroll.Offset.GetTyped())

	scroll.ScrollToBottom()
	assert.Equal(t, 90, scroll.Offset.GetTyped())
}
