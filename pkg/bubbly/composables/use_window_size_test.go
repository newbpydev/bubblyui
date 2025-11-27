package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseWindowSize_BreakpointCalculation tests that breakpoints are calculated correctly
func TestUseWindowSize_BreakpointCalculation(t *testing.T) {
	tests := []struct {
		name               string
		width              int
		height             int
		expectedBreakpoint Breakpoint
	}{
		{"xs breakpoint - 40 cols", 40, 24, BreakpointXS},
		{"xs breakpoint - 59 cols", 59, 24, BreakpointXS},
		{"sm breakpoint - 60 cols", 60, 24, BreakpointSM},
		{"sm breakpoint - 79 cols", 79, 24, BreakpointSM},
		{"md breakpoint - 80 cols", 80, 24, BreakpointMD},
		{"md breakpoint - 119 cols", 119, 24, BreakpointMD},
		{"lg breakpoint - 120 cols", 120, 24, BreakpointLG},
		{"lg breakpoint - 159 cols", 159, 24, BreakpointLG},
		{"xl breakpoint - 160 cols", 160, 24, BreakpointXL},
		{"xl breakpoint - 200 cols", 200, 24, BreakpointXL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			ws := UseWindowSize(ctx)

			ws.SetSize(tt.width, tt.height)

			assert.Equal(t, tt.expectedBreakpoint, ws.Breakpoint.GetTyped(),
				"Breakpoint should be %s for width %d", tt.expectedBreakpoint, tt.width)
		})
	}
}

// TestUseWindowSize_SetSizeUpdatesAllDerivedValues tests that SetSize updates all derived values
func TestUseWindowSize_SetSizeUpdatesAllDerivedValues(t *testing.T) {
	ctx := createTestContext()
	ws := UseWindowSize(ctx)

	// Initial state should be defaults (80x24)
	assert.Equal(t, 80, ws.Width.GetTyped(), "Initial width")
	assert.Equal(t, 24, ws.Height.GetTyped(), "Initial height")

	// Set new size
	ws.SetSize(160, 40)

	assert.Equal(t, 160, ws.Width.GetTyped(), "Width should be updated")
	assert.Equal(t, 40, ws.Height.GetTyped(), "Height should be updated")
	assert.Equal(t, BreakpointXL, ws.Breakpoint.GetTyped(), "Breakpoint should be XL")
	assert.True(t, ws.SidebarVisible.GetTyped(), "Sidebar should be visible at XL")
	assert.Equal(t, 4, ws.GridColumns.GetTyped(), "Grid columns should be 4 at XL")
}

// TestUseWindowSize_MinDimensionEnforcement tests that minimum dimensions are enforced
func TestUseWindowSize_MinDimensionEnforcement(t *testing.T) {
	tests := []struct {
		name           string
		minWidth       int
		minHeight      int
		setWidth       int
		setHeight      int
		expectedWidth  int
		expectedHeight int
	}{
		{"below min width", 40, 10, 20, 24, 40, 24},
		{"below min height", 40, 10, 80, 5, 80, 10},
		{"both below min", 40, 10, 20, 5, 40, 10},
		{"above min - no change", 40, 10, 100, 30, 100, 30},
		{"exact min", 40, 10, 40, 10, 40, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			ws := UseWindowSize(ctx, WithMinDimensions(tt.minWidth, tt.minHeight))

			ws.SetSize(tt.setWidth, tt.setHeight)

			assert.Equal(t, tt.expectedWidth, ws.Width.GetTyped(),
				"Width should be clamped to min")
			assert.Equal(t, tt.expectedHeight, ws.Height.GetTyped(),
				"Height should be clamped to min")
		})
	}
}

// TestUseWindowSize_CustomBreakpointConfiguration tests custom breakpoint thresholds
func TestUseWindowSize_CustomBreakpointConfiguration(t *testing.T) {
	ctx := createTestContext()

	// Custom breakpoints: tighter thresholds
	customConfig := BreakpointConfig{
		XS: 0,
		SM: 40,
		MD: 60,
		LG: 80,
		XL: 100,
	}
	ws := UseWindowSize(ctx, WithBreakpoints(customConfig))

	tests := []struct {
		width    int
		expected Breakpoint
	}{
		{30, BreakpointXS},
		{45, BreakpointSM},
		{70, BreakpointMD},
		{90, BreakpointLG},
		{120, BreakpointXL},
	}

	for _, tt := range tests {
		ws.SetSize(tt.width, 24)
		assert.Equal(t, tt.expected, ws.Breakpoint.GetTyped(),
			"Custom breakpoint for width %d", tt.width)
	}
}

// TestUseWindowSize_GetContentWidth tests content width calculation
func TestUseWindowSize_GetContentWidth(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		sidebarWidth  int
		sidebarOn     bool // whether breakpoint shows sidebar
		expectedWidth int
	}{
		// Large screen with sidebar visible
		{"xl with sidebar", 160, 30, true, 130},
		// Small screen without sidebar
		{"xs without sidebar", 50, 30, false, 50},
		// Medium screen with sidebar
		{"md with sidebar", 100, 25, true, 75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			ws := UseWindowSize(ctx, WithSidebarWidth(tt.sidebarWidth))

			ws.SetSize(tt.width, 24)

			contentWidth := ws.GetContentWidth()
			assert.Equal(t, tt.expectedWidth, contentWidth,
				"Content width for %s", tt.name)
		})
	}
}

// TestUseWindowSize_GetCardWidth tests card width calculation
func TestUseWindowSize_GetCardWidth(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		expectedCards int // cards = gridColumns, cardWidth = width / gridColumns
	}{
		{"xs - 1 column", 50, 1},
		{"sm - 2 columns", 70, 2},
		{"md - 2 columns", 100, 2},
		{"lg - 3 columns", 140, 3},
		{"xl - 4 columns", 180, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			ws := UseWindowSize(ctx)

			ws.SetSize(tt.width, 24)

			cardWidth := ws.GetCardWidth()
			expectedCardWidth := tt.width / tt.expectedCards

			assert.Equal(t, expectedCardWidth, cardWidth,
				"Card width for %d columns at width %d", tt.expectedCards, tt.width)
		})
	}
}

// TestUseWindowSize_SidebarVisibility tests sidebar visibility per breakpoint
func TestUseWindowSize_SidebarVisibility(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		shouldShow bool
	}{
		{"xs - hidden", 50, false},
		{"sm - hidden", 70, false},
		{"md - visible", 100, true},
		{"lg - visible", 140, true},
		{"xl - visible", 180, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			ws := UseWindowSize(ctx)

			ws.SetSize(tt.width, 24)

			assert.Equal(t, tt.shouldShow, ws.SidebarVisible.GetTyped(),
				"Sidebar visibility for width %d", tt.width)
		})
	}
}

// TestUseWindowSize_GridColumns tests grid column calculation per breakpoint
func TestUseWindowSize_GridColumns(t *testing.T) {
	tests := []struct {
		name            string
		width           int
		expectedColumns int
	}{
		{"xs - 1 column", 50, 1},
		{"sm - 2 columns", 70, 2},
		{"md - 2 columns", 100, 2},
		{"lg - 3 columns", 140, 3},
		{"xl - 4 columns", 180, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			ws := UseWindowSize(ctx)

			ws.SetSize(tt.width, 24)

			assert.Equal(t, tt.expectedColumns, ws.GridColumns.GetTyped(),
				"Grid columns for width %d", tt.width)
		})
	}
}

// TestUseWindowSize_DefaultValues tests default values on initialization
func TestUseWindowSize_DefaultValues(t *testing.T) {
	ctx := createTestContext()
	ws := UseWindowSize(ctx)

	// Default dimensions are 80x24 (standard terminal)
	assert.Equal(t, 80, ws.Width.GetTyped(), "Default width")
	assert.Equal(t, 24, ws.Height.GetTyped(), "Default height")
	assert.Equal(t, BreakpointMD, ws.Breakpoint.GetTyped(), "Default breakpoint for 80 cols")
	assert.True(t, ws.SidebarVisible.GetTyped(), "Sidebar visible at MD")
	assert.Equal(t, 2, ws.GridColumns.GetTyped(), "Grid columns at MD")
}

// TestUseWindowSize_WithAllOptions tests combining all options
func TestUseWindowSize_WithAllOptions(t *testing.T) {
	ctx := createTestContext()

	customConfig := BreakpointConfig{
		XS: 0,
		SM: 30,
		MD: 50,
		LG: 70,
		XL: 90,
	}

	ws := UseWindowSize(ctx,
		WithBreakpoints(customConfig),
		WithMinDimensions(25, 10),
		WithSidebarWidth(20),
	)

	// Test that all options work together
	ws.SetSize(60, 20)

	assert.Equal(t, 60, ws.Width.GetTyped())
	assert.Equal(t, 20, ws.Height.GetTyped())
	assert.Equal(t, BreakpointMD, ws.Breakpoint.GetTyped(), "60 is MD with custom config")

	// Content width should account for sidebar
	contentWidth := ws.GetContentWidth()
	assert.Equal(t, 40, contentWidth, "160 - 20 sidebar = 40")
}

// TestUseWindowSize_ZeroDimensions tests handling of zero dimensions
func TestUseWindowSize_ZeroDimensions(t *testing.T) {
	ctx := createTestContext()
	ws := UseWindowSize(ctx, WithMinDimensions(40, 10))

	// Set to 0x0 - should enforce minimums
	ws.SetSize(0, 0)

	assert.Equal(t, 40, ws.Width.GetTyped(), "Width should be clamped to min")
	assert.Equal(t, 10, ws.Height.GetTyped(), "Height should be clamped to min")
}

// TestUseWindowSize_WorksWithCreateShared tests shared composable pattern
func TestUseWindowSize_WorksWithCreateShared(t *testing.T) {
	// Create shared instance
	sharedWindowSize := CreateShared(func(ctx *bubbly.Context) *WindowSizeReturn {
		return UseWindowSize(ctx)
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	ws1 := sharedWindowSize(ctx1)
	ws2 := sharedWindowSize(ctx2)

	// Both should be the same instance
	ws1.SetSize(120, 30)

	assert.Equal(t, 120, ws2.Width.GetTyped(), "Shared instance should have same width")
	assert.Equal(t, 30, ws2.Height.GetTyped(), "Shared instance should have same height")
}

// TestUseWindowSize_BreakpointConstants verifies breakpoint constant values
func TestUseWindowSize_BreakpointConstants(t *testing.T) {
	assert.Equal(t, Breakpoint("xs"), BreakpointXS)
	assert.Equal(t, Breakpoint("sm"), BreakpointSM)
	assert.Equal(t, Breakpoint("md"), BreakpointMD)
	assert.Equal(t, Breakpoint("lg"), BreakpointLG)
	assert.Equal(t, Breakpoint("xl"), BreakpointXL)
}
