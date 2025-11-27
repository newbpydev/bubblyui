package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// Breakpoint represents responsive breakpoints for terminal layouts.
// These follow a mobile-first approach similar to CSS frameworks.
type Breakpoint string

const (
	// BreakpointXS represents extra small terminals (<60 cols).
	BreakpointXS Breakpoint = "xs"
	// BreakpointSM represents small terminals (60-79 cols).
	BreakpointSM Breakpoint = "sm"
	// BreakpointMD represents medium terminals (80-119 cols).
	BreakpointMD Breakpoint = "md"
	// BreakpointLG represents large terminals (120-159 cols).
	BreakpointLG Breakpoint = "lg"
	// BreakpointXL represents extra large terminals (160+ cols).
	BreakpointXL Breakpoint = "xl"
)

// Default breakpoint thresholds (in columns).
const (
	defaultBreakpointSM = 60
	defaultBreakpointMD = 80
	defaultBreakpointLG = 120
	defaultBreakpointXL = 160
)

// Default terminal dimensions.
const (
	defaultWidth  = 80
	defaultHeight = 24
)

// Default sidebar width for content calculations.
const defaultSidebarWidth = 30

// BreakpointConfig allows custom breakpoint thresholds.
// Each value represents the minimum width (in columns) for that breakpoint.
type BreakpointConfig struct {
	// XS is the minimum width for extra small (default: 0).
	XS int
	// SM is the minimum width for small (default: 60).
	SM int
	// MD is the minimum width for medium (default: 80).
	MD int
	// LG is the minimum width for large (default: 120).
	LG int
	// XL is the minimum width for extra large (default: 160).
	XL int
}

// windowSizeConfig holds internal configuration for UseWindowSize.
type windowSizeConfig struct {
	breakpoints  BreakpointConfig
	minWidth     int
	minHeight    int
	sidebarWidth int
}

// defaultWindowSizeConfig returns the default configuration.
func defaultWindowSizeConfig() windowSizeConfig {
	return windowSizeConfig{
		breakpoints: BreakpointConfig{
			XS: 0,
			SM: defaultBreakpointSM,
			MD: defaultBreakpointMD,
			LG: defaultBreakpointLG,
			XL: defaultBreakpointXL,
		},
		minWidth:     0,
		minHeight:    0,
		sidebarWidth: defaultSidebarWidth,
	}
}

// WindowSizeOption configures UseWindowSize.
type WindowSizeOption func(*windowSizeConfig)

// WithBreakpoints sets custom breakpoint thresholds.
//
// Example:
//
//	ws := UseWindowSize(ctx, WithBreakpoints(BreakpointConfig{
//	    XS: 0, SM: 40, MD: 60, LG: 80, XL: 100,
//	}))
func WithBreakpoints(config BreakpointConfig) WindowSizeOption {
	return func(c *windowSizeConfig) {
		c.breakpoints = config
	}
}

// WithMinDimensions sets minimum width and height.
// Values below these minimums will be clamped.
//
// Example:
//
//	ws := UseWindowSize(ctx, WithMinDimensions(40, 10))
func WithMinDimensions(minWidth, minHeight int) WindowSizeOption {
	return func(c *windowSizeConfig) {
		c.minWidth = minWidth
		c.minHeight = minHeight
	}
}

// WithSidebarWidth sets sidebar width for content calculation.
// This affects GetContentWidth() when sidebar is visible.
//
// Example:
//
//	ws := UseWindowSize(ctx, WithSidebarWidth(25))
func WithSidebarWidth(width int) WindowSizeOption {
	return func(c *windowSizeConfig) {
		c.sidebarWidth = width
	}
}

// WindowSizeReturn is the return value of UseWindowSize.
// It provides reactive terminal dimensions and responsive layout helpers.
type WindowSizeReturn struct {
	// Width is the current terminal width in columns.
	Width *bubbly.Ref[int]

	// Height is the current terminal height in rows.
	Height *bubbly.Ref[int]

	// Breakpoint is the current responsive breakpoint.
	Breakpoint *bubbly.Ref[Breakpoint]

	// SidebarVisible indicates if sidebar should be visible.
	// True for MD, LG, XL breakpoints; false for XS, SM.
	SidebarVisible *bubbly.Ref[bool]

	// GridColumns is the recommended number of grid columns.
	// XS=1, SM=2, MD=2, LG=3, XL=4.
	GridColumns *bubbly.Ref[int]

	// config holds internal configuration
	config windowSizeConfig
}

// SetSize updates the window dimensions and recalculates derived values.
// Dimensions are clamped to configured minimums.
//
// Example:
//
//	ws.SetSize(120, 40)  // Handle tea.WindowSizeMsg
func (w *WindowSizeReturn) SetSize(width, height int) {
	// Clamp to minimums
	if width < w.config.minWidth {
		width = w.config.minWidth
	}
	if height < w.config.minHeight {
		height = w.config.minHeight
	}

	// Update dimensions
	w.Width.Set(width)
	w.Height.Set(height)

	// Calculate and update derived values
	bp := w.calculateBreakpoint(width)
	w.Breakpoint.Set(bp)
	w.SidebarVisible.Set(w.isSidebarVisible(bp))
	w.GridColumns.Set(w.calculateGridColumns(bp))
}

// GetContentWidth returns available content width (accounting for sidebar).
// If sidebar is visible, subtracts sidebar width from total width.
//
// Example:
//
//	contentWidth := ws.GetContentWidth()
//	// Use contentWidth to size main content area
func (w *WindowSizeReturn) GetContentWidth() int {
	width := w.Width.GetTyped()
	if w.SidebarVisible.GetTyped() {
		return width - w.config.sidebarWidth
	}
	return width
}

// GetCardWidth returns optimal card width for current grid.
// Calculated as total width divided by grid columns.
//
// Example:
//
//	cardWidth := ws.GetCardWidth()
//	// Use cardWidth to size cards in a grid layout
func (w *WindowSizeReturn) GetCardWidth() int {
	width := w.Width.GetTyped()
	columns := w.GridColumns.GetTyped()
	if columns <= 0 {
		return width
	}
	return width / columns
}

// calculateBreakpoint determines the breakpoint for a given width.
func (w *WindowSizeReturn) calculateBreakpoint(width int) Breakpoint {
	bp := w.config.breakpoints
	switch {
	case width >= bp.XL:
		return BreakpointXL
	case width >= bp.LG:
		return BreakpointLG
	case width >= bp.MD:
		return BreakpointMD
	case width >= bp.SM:
		return BreakpointSM
	default:
		return BreakpointXS
	}
}

// isSidebarVisible determines if sidebar should be visible for a breakpoint.
// Sidebar is visible for MD, LG, XL breakpoints.
func (w *WindowSizeReturn) isSidebarVisible(bp Breakpoint) bool {
	switch bp {
	case BreakpointMD, BreakpointLG, BreakpointXL:
		return true
	default:
		return false
	}
}

// calculateGridColumns determines grid columns for a breakpoint.
func (w *WindowSizeReturn) calculateGridColumns(bp Breakpoint) int {
	switch bp {
	case BreakpointXL:
		return 4
	case BreakpointLG:
		return 3
	case BreakpointMD, BreakpointSM:
		return 2
	default:
		return 1
	}
}

// UseWindowSize creates a window size composable for responsive layouts.
// It tracks terminal dimensions and calculates responsive breakpoints,
// sidebar visibility, and grid column recommendations.
//
// This composable is essential for building responsive TUI applications
// that adapt to different terminal sizes.
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - opts: Optional configuration (breakpoints, min dimensions, sidebar width)
//
// Returns:
//   - *WindowSizeReturn: A struct containing reactive dimensions and helpers
//
// Default Breakpoints:
//   - XS: <60 cols (1 grid column, no sidebar)
//   - SM: 60-79 cols (2 grid columns, no sidebar)
//   - MD: 80-119 cols (2 grid columns, sidebar visible)
//   - LG: 120-159 cols (3 grid columns, sidebar visible)
//   - XL: 160+ cols (4 grid columns, sidebar visible)
//
// Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    ws := composables.UseWindowSize(ctx)
//	    ctx.Expose("windowSize", ws)
//
//	    ctx.On("resize", func(data interface{}) {
//	        if size, ok := data.(map[string]int); ok {
//	            ws.SetSize(size["width"], size["height"])
//	        }
//	    })
//	})
//
// Custom Configuration:
//
//	ws := UseWindowSize(ctx,
//	    WithBreakpoints(BreakpointConfig{XS: 0, SM: 40, MD: 60, LG: 80, XL: 100}),
//	    WithMinDimensions(40, 10),
//	    WithSidebarWidth(25),
//	)
//
// Integration with tea.WindowSizeMsg:
//
//	WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
//	    if wmsg, ok := msg.(tea.WindowSizeMsg); ok {
//	        comp.Emit("resize", map[string]int{
//	            "width": wmsg.Width, "height": wmsg.Height,
//	        })
//	    }
//	    return nil
//	})
func UseWindowSize(ctx *bubbly.Context, opts ...WindowSizeOption) *WindowSizeReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseWindowSize", time.Since(start))
	}()

	// Apply options
	config := defaultWindowSizeConfig()
	for _, opt := range opts {
		opt(&config)
	}

	// Create reactive refs with defaults
	width := bubbly.NewRef(defaultWidth)
	height := bubbly.NewRef(defaultHeight)
	breakpoint := bubbly.NewRef(BreakpointMD) // 80 cols = MD
	sidebarVisible := bubbly.NewRef(true)     // MD shows sidebar
	gridColumns := bubbly.NewRef(2)           // MD = 2 columns

	ws := &WindowSizeReturn{
		Width:          width,
		Height:         height,
		Breakpoint:     breakpoint,
		SidebarVisible: sidebarVisible,
		GridColumns:    gridColumns,
		config:         config,
	}

	// Initialize derived values based on default dimensions
	ws.SetSize(defaultWidth, defaultHeight)

	return ws
}
