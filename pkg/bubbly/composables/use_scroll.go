package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// ScrollReturn is the return value of UseScroll.
// It provides reactive scroll state management for viewport scrolling in TUI applications.
type ScrollReturn struct {
	// Offset is the current scroll position (0-indexed).
	Offset *bubbly.Ref[int]

	// MaxOffset is the maximum valid scroll offset.
	MaxOffset *bubbly.Ref[int]

	// VisibleCount is the number of visible items in the viewport.
	VisibleCount *bubbly.Ref[int]

	// TotalItems is the total number of items in the list.
	TotalItems *bubbly.Ref[int]
}

// ScrollUp moves the scroll position up by one.
// Respects the lower bound (0).
//
// Example:
//
//	ctx.On("scrollUp", func(_ interface{}) {
//	    scroll.ScrollUp()
//	})
func (s *ScrollReturn) ScrollUp() {
	current := s.Offset.GetTyped()
	if current > 0 {
		s.Offset.Set(current - 1)
	}
}

// ScrollDown moves the scroll position down by one.
// Respects the upper bound (MaxOffset).
//
// Example:
//
//	ctx.On("scrollDown", func(_ interface{}) {
//	    scroll.ScrollDown()
//	})
func (s *ScrollReturn) ScrollDown() {
	current := s.Offset.GetTyped()
	maxOffset := s.MaxOffset.GetTyped()
	if current < maxOffset {
		s.Offset.Set(current + 1)
	}
}

// ScrollTo moves to a specific offset (clamped to valid range).
// Negative values are clamped to 0, values beyond MaxOffset are clamped to MaxOffset.
//
// Example:
//
//	scroll.ScrollTo(50)  // Jump to offset 50
func (s *ScrollReturn) ScrollTo(offset int) {
	maxOffset := s.MaxOffset.GetTyped()

	// Clamp to valid range
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	s.Offset.Set(offset)
}

// ScrollToTop scrolls to the beginning (offset 0).
//
// Example:
//
//	ctx.On("home", func(_ interface{}) {
//	    scroll.ScrollToTop()
//	})
func (s *ScrollReturn) ScrollToTop() {
	s.Offset.Set(0)
}

// ScrollToBottom scrolls to the end (MaxOffset).
//
// Example:
//
//	ctx.On("end", func(_ interface{}) {
//	    scroll.ScrollToBottom()
//	})
func (s *ScrollReturn) ScrollToBottom() {
	s.Offset.Set(s.MaxOffset.GetTyped())
}

// PageUp scrolls up by visible count.
// Respects the lower bound (0).
//
// Example:
//
//	ctx.On("pageUp", func(_ interface{}) {
//	    scroll.PageUp()
//	})
func (s *ScrollReturn) PageUp() {
	current := s.Offset.GetTyped()
	visibleCount := s.VisibleCount.GetTyped()

	newOffset := current - visibleCount
	if newOffset < 0 {
		newOffset = 0
	}

	s.Offset.Set(newOffset)
}

// PageDown scrolls down by visible count.
// Respects the upper bound (MaxOffset).
//
// Example:
//
//	ctx.On("pageDown", func(_ interface{}) {
//	    scroll.PageDown()
//	})
func (s *ScrollReturn) PageDown() {
	current := s.Offset.GetTyped()
	visibleCount := s.VisibleCount.GetTyped()
	maxOffset := s.MaxOffset.GetTyped()

	newOffset := current + visibleCount
	if newOffset > maxOffset {
		newOffset = maxOffset
	}

	s.Offset.Set(newOffset)
}

// IsAtTop returns true if scrolled to top (offset is 0).
//
// Example:
//
//	if scroll.IsAtTop() {
//	    // Disable up arrow
//	}
func (s *ScrollReturn) IsAtTop() bool {
	return s.Offset.GetTyped() == 0
}

// IsAtBottom returns true if scrolled to bottom (offset equals MaxOffset).
//
// Example:
//
//	if scroll.IsAtBottom() {
//	    // Disable down arrow
//	}
func (s *ScrollReturn) IsAtBottom() bool {
	return s.Offset.GetTyped() >= s.MaxOffset.GetTyped()
}

// SetTotalItems updates the total item count and recalculates max offset.
// If current offset is beyond new max, it is clamped.
//
// Example:
//
//	// After loading more items
//	scroll.SetTotalItems(newCount)
func (s *ScrollReturn) SetTotalItems(count int) {
	if count < 0 {
		count = 0
	}

	s.TotalItems.Set(count)
	s.recalculateMaxOffset()

	// Clamp current offset if necessary
	s.ScrollTo(s.Offset.GetTyped())
}

// SetVisibleCount updates visible count and recalculates max offset.
// This is useful when the viewport size changes (e.g., terminal resize).
//
// Example:
//
//	// After terminal resize
//	scroll.SetVisibleCount(newHeight - headerHeight)
func (s *ScrollReturn) SetVisibleCount(count int) {
	if count < 0 {
		count = 0
	}

	s.VisibleCount.Set(count)
	s.recalculateMaxOffset()

	// Clamp current offset if necessary
	s.ScrollTo(s.Offset.GetTyped())
}

// recalculateMaxOffset updates MaxOffset based on TotalItems and VisibleCount.
// MaxOffset = max(0, TotalItems - VisibleCount)
func (s *ScrollReturn) recalculateMaxOffset() {
	total := s.TotalItems.GetTyped()
	visible := s.VisibleCount.GetTyped()

	maxOffset := total - visible
	if maxOffset < 0 {
		maxOffset = 0
	}

	s.MaxOffset.Set(maxOffset)
}

// UseScroll creates a scroll management composable for viewport scrolling.
// It tracks scroll offset and provides methods for navigation within a scrollable list.
//
// This composable is essential for building scrollable TUI components like lists,
// tables, and log viewers.
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - totalItems: The total number of items in the list
//   - visibleCount: The number of items visible in the viewport
//
// Returns:
//   - *ScrollReturn: A struct containing reactive scroll state and methods
//
// Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    items := []string{"Item 1", "Item 2", ..., "Item 100"}
//	    visibleCount := 10
//
//	    scroll := composables.UseScroll(ctx, len(items), visibleCount)
//	    ctx.Expose("scroll", scroll)
//
//	    ctx.On("scrollUp", func(_ interface{}) {
//	        scroll.ScrollUp()
//	    })
//	    ctx.On("scrollDown", func(_ interface{}) {
//	        scroll.ScrollDown()
//	    })
//	    ctx.On("pageUp", func(_ interface{}) {
//	        scroll.PageUp()
//	    })
//	    ctx.On("pageDown", func(_ interface{}) {
//	        scroll.PageDown()
//	    })
//	}).
//	WithMultiKeyBindings("scrollUp", "Scroll up", "up", "k").
//	WithMultiKeyBindings("scrollDown", "Scroll down", "down", "j").
//	WithKeyBinding("pgup", "pageUp", "Page up").
//	WithKeyBinding("pgdown", "pageDown", "Page down")
//
// Integration with CreateShared:
//
//	var UseSharedScroll = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.ScrollReturn {
//	        return composables.UseScroll(ctx, 100, 10)
//	    },
//	)
//
// Rendering visible items:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    scroll := ctx.Get("scroll").(*composables.ScrollReturn)
//	    offset := scroll.Offset.GetTyped()
//	    visible := scroll.VisibleCount.GetTyped()
//
//	    // Render only visible items
//	    visibleItems := items[offset:min(offset+visible, len(items))]
//	    // ... render visibleItems
//	})
func UseScroll(ctx *bubbly.Context, totalItems, visibleCount int) *ScrollReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseScroll", time.Since(start))
	}()

	// Ensure non-negative values
	if totalItems < 0 {
		totalItems = 0
	}
	if visibleCount < 0 {
		visibleCount = 0
	}

	// Calculate initial max offset
	maxOffset := totalItems - visibleCount
	if maxOffset < 0 {
		maxOffset = 0
	}

	// Create reactive refs
	offset := bubbly.NewRef(0)
	maxOffsetRef := bubbly.NewRef(maxOffset)
	visibleCountRef := bubbly.NewRef(visibleCount)
	totalItemsRef := bubbly.NewRef(totalItems)

	return &ScrollReturn{
		Offset:       offset,
		MaxOffset:    maxOffsetRef,
		VisibleCount: visibleCountRef,
		TotalItems:   totalItemsRef,
	}
}
