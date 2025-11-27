// Package composables provides shared state and logic for the responsive layouts example.
package composables

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// Breakpoint represents a responsive breakpoint.
type Breakpoint string

const (
	// BreakpointXS is for very small terminals (<60 cols).
	BreakpointXS Breakpoint = "xs"
	// BreakpointSM is for small terminals (60-79 cols).
	BreakpointSM Breakpoint = "sm"
	// BreakpointMD is for medium terminals (80-119 cols).
	BreakpointMD Breakpoint = "md"
	// BreakpointLG is for large terminals (120-159 cols).
	BreakpointLG Breakpoint = "lg"
	// BreakpointXL is for extra large terminals (160+ cols).
	BreakpointXL Breakpoint = "xl"
)

// BreakpointWidths defines the minimum width for each breakpoint.
var BreakpointWidths = map[Breakpoint]int{
	BreakpointXS: 0,
	BreakpointSM: 60,
	BreakpointMD: 80,
	BreakpointLG: 120,
	BreakpointXL: 160,
}

// MinWidth is the minimum supported terminal width.
const MinWidth = 60

// MinHeight is the minimum supported terminal height.
const MinHeight = 20

// WindowSizeComposable provides reactive window size state.
type WindowSizeComposable struct {
	// Width is the current terminal width.
	Width *bubbly.Ref[int]
	// Height is the current terminal height.
	Height *bubbly.Ref[int]
	// Breakpoint is the current responsive breakpoint.
	Breakpoint *bubbly.Ref[Breakpoint]
	// SidebarVisible indicates if sidebar should be shown (based on width).
	SidebarVisible *bubbly.Ref[bool]
	// GridColumns is the recommended number of grid columns.
	GridColumns *bubbly.Ref[int]
}

// UseWindowSize creates a window size composable for responsive layouts.
func UseWindowSize(ctx *bubbly.Context) *WindowSizeComposable {
	// Initialize with reasonable defaults (will be updated on first WindowSizeMsg)
	width := bubbly.NewRef(80)
	height := bubbly.NewRef(24)
	breakpoint := bubbly.NewRef(BreakpointMD)
	sidebarVisible := bubbly.NewRef(true)
	gridColumns := bubbly.NewRef(3)

	return &WindowSizeComposable{
		Width:          width,
		Height:         height,
		Breakpoint:     breakpoint,
		SidebarVisible: sidebarVisible,
		GridColumns:    gridColumns,
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
	bp := w.calculateBreakpoint(width)
	w.Breakpoint.Set(bp)

	// Calculate sidebar visibility (hide on small screens)
	w.SidebarVisible.Set(width >= BreakpointWidths[BreakpointMD])

	// Calculate grid columns based on width
	w.GridColumns.Set(w.calculateGridColumns(width))
}

// calculateBreakpoint determines the breakpoint for a given width.
func (w *WindowSizeComposable) calculateBreakpoint(width int) Breakpoint {
	switch {
	case width >= BreakpointWidths[BreakpointXL]:
		return BreakpointXL
	case width >= BreakpointWidths[BreakpointLG]:
		return BreakpointLG
	case width >= BreakpointWidths[BreakpointMD]:
		return BreakpointMD
	case width >= BreakpointWidths[BreakpointSM]:
		return BreakpointSM
	default:
		return BreakpointXS
	}
}

// calculateGridColumns determines optimal grid columns for width.
func (w *WindowSizeComposable) calculateGridColumns(width int) int {
	switch {
	case width >= 160:
		return 5
	case width >= 120:
		return 4
	case width >= 80:
		return 3
	case width >= 60:
		return 2
	default:
		return 1
	}
}

// GetContentWidth returns the available content width (accounting for sidebar).
func (w *WindowSizeComposable) GetContentWidth() int {
	width := w.Width.GetTyped()
	if w.SidebarVisible.GetTyped() {
		// Sidebar takes ~22 chars (20 content + 2 border)
		return width - 22
	}
	return width - 2 // Just borders
}

// GetCardWidth returns the optimal card width for the current grid.
func (w *WindowSizeComposable) GetCardWidth() int {
	contentWidth := w.GetContentWidth()
	cols := w.GridColumns.GetTyped()
	// Account for gaps between cards (1 char each)
	gaps := cols - 1
	return (contentWidth - gaps) / cols
}

// UseSharedWindowSize is a singleton composable for window size across all components.
var UseSharedWindowSize = composables.CreateShared(
	func(ctx *bubbly.Context) *WindowSizeComposable {
		return UseWindowSize(ctx)
	},
)
