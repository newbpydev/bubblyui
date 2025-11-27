// Package composables provides shared state and logic for the AI chat demo.
package composables

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// Breakpoint represents a responsive breakpoint.
type Breakpoint string

const (
	// BreakpointSM is for small terminals (<80 cols).
	BreakpointSM Breakpoint = "sm"
	// BreakpointMD is for medium terminals (80-119 cols).
	BreakpointMD Breakpoint = "md"
	// BreakpointLG is for large terminals (120+ cols).
	BreakpointLG Breakpoint = "lg"
)

// MinWidth is the minimum supported terminal width.
const MinWidth = 60

// MinHeight is the minimum supported terminal height.
const MinHeight = 15

// WindowSizeComposable provides reactive window size state.
type WindowSizeComposable struct {
	// Width is the current terminal width.
	Width *bubbly.Ref[int]
	// Height is the current terminal height.
	Height *bubbly.Ref[int]
	// Breakpoint is the current responsive breakpoint.
	Breakpoint *bubbly.Ref[Breakpoint]
	// SidebarVisible indicates if sidebar should be shown.
	SidebarVisible *bubbly.Ref[bool]
	// SidebarWidth is the width of the sidebar.
	SidebarWidth *bubbly.Ref[int]
}

// UseWindowSize creates a window size composable for responsive layouts.
func UseWindowSize(ctx *bubbly.Context) *WindowSizeComposable {
	return &WindowSizeComposable{
		Width:          bubbly.NewRef(80),
		Height:         bubbly.NewRef(24),
		Breakpoint:     bubbly.NewRef(BreakpointMD),
		SidebarVisible: bubbly.NewRef(true),
		SidebarWidth:   bubbly.NewRef(24),
	}
}

// SetSize updates the window size and recalculates responsive values.
func (w *WindowSizeComposable) SetSize(width, height int) {
	// Enforce minimum dimensions
	if width < MinWidth {
		width = MinWidth
	}
	if height < MinHeight {
		height = MinHeight
	}

	w.Width.Set(width)
	w.Height.Set(height)

	// Calculate breakpoint
	var bp Breakpoint
	switch {
	case width >= 120:
		bp = BreakpointLG
	case width >= 80:
		bp = BreakpointMD
	default:
		bp = BreakpointSM
	}
	w.Breakpoint.Set(bp)

	// Sidebar visibility and width based on breakpoint
	switch bp {
	case BreakpointLG:
		w.SidebarVisible.Set(true)
		w.SidebarWidth.Set(28)
	case BreakpointMD:
		w.SidebarVisible.Set(true)
		w.SidebarWidth.Set(24)
	default:
		w.SidebarVisible.Set(false)
		w.SidebarWidth.Set(0)
	}
}

// GetContentWidth returns the available content width.
func (w *WindowSizeComposable) GetContentWidth() int {
	width := w.Width.GetTyped()
	if w.SidebarVisible.GetTyped() {
		return width - w.SidebarWidth.GetTyped()
	}
	return width
}

// GetMessageListHeight returns the height available for the message list.
func (w *WindowSizeComposable) GetMessageListHeight() int {
	height := w.Height.GetTyped()
	// Layout: header (1) + main content + input (3) + footer (1) = height
	// Main content height = height - 5
	// Inside box: border (2) + padding (2) + title (1) + divider (1) = 6
	// Available for messages = height - 5 - 6 = height - 11
	return height - 11
}

// GetMainContentHeight returns the height for the main content area (sidebar + messages).
func (w *WindowSizeComposable) GetMainContentHeight() int {
	height := w.Height.GetTyped()
	// Layout: header (1) + main content + input (3) + footer (1) = height
	return height - 5
}

// UseSharedWindowSize is a singleton composable for window size.
var UseSharedWindowSize = composables.CreateShared(
	func(ctx *bubbly.Context) *WindowSizeComposable {
		return UseWindowSize(ctx)
	},
)
