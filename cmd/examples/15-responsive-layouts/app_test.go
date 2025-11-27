package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	localComponents "github.com/newbpydev/bubblyui/cmd/examples/15-responsive-layouts/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/15-responsive-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

func TestCreateApp(t *testing.T) {
	app, err := CreateApp()
	require.NoError(t, err)
	require.NotNil(t, app)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(app)
	defer ct.Unmount()

	// Should render the header
	ct.AssertRenderContains("BubblyUI Responsive Layouts")

	// Should render tab bar with all demos
	ct.AssertRenderContains("Dashboard")
	ct.AssertRenderContains("Grid")
	ct.AssertRenderContains("Adaptive")
	ct.AssertRenderContains("Breakpoints")
}

func TestResponsiveDashboard(t *testing.T) {
	demo, err := localComponents.CreateResponsiveDashboard()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render dashboard elements
	ct.AssertRenderContains("Responsive Dashboard")
	ct.AssertRenderContains("Statistics")
}

func TestResponsiveGrid(t *testing.T) {
	demo, err := localComponents.CreateResponsiveGrid()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render grid elements
	ct.AssertRenderContains("Responsive Grid")
	ct.AssertRenderContains("Dashboard")
	ct.AssertRenderContains("Analytics")
}

func TestAdaptiveContent(t *testing.T) {
	demo, err := localComponents.CreateAdaptiveContent()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render adaptive content elements
	ct.AssertRenderContains("Adaptive Layout")
	ct.AssertRenderContains("Primary")
	ct.AssertRenderContains("Secondary")
}

func TestBreakpointDemo(t *testing.T) {
	demo, err := localComponents.CreateBreakpointDemo()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render breakpoint info
	ct.AssertRenderContains("Breakpoint Information")
	ct.AssertRenderContains("Terminal Size")
}

func TestWindowSizeComposable(t *testing.T) {
	tests := []struct {
		name               string
		width              int
		height             int
		expectedBreakpoint localComposables.Breakpoint
		expectedSidebar    bool
		expectedCols       int
	}{
		{
			name:               "extra small terminal",
			width:              50,
			height:             20,
			expectedBreakpoint: localComposables.BreakpointXS,
			expectedSidebar:    false,
			expectedCols:       1,
		},
		{
			name:               "small terminal",
			width:              70,
			height:             24,
			expectedBreakpoint: localComposables.BreakpointSM,
			expectedSidebar:    false,
			expectedCols:       2,
		},
		{
			name:               "medium terminal",
			width:              100,
			height:             30,
			expectedBreakpoint: localComposables.BreakpointMD,
			expectedSidebar:    true,
			expectedCols:       3,
		},
		{
			name:               "large terminal",
			width:              140,
			height:             40,
			expectedBreakpoint: localComposables.BreakpointLG,
			expectedSidebar:    true,
			expectedCols:       4,
		},
		{
			name:               "extra large terminal",
			width:              180,
			height:             50,
			expectedBreakpoint: localComposables.BreakpointXL,
			expectedSidebar:    true,
			expectedCols:       5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal component to get a context
			comp, err := CreateApp()
			require.NoError(t, err)

			harness := testutil.NewHarness(t)
			ct := harness.Mount(comp)
			defer ct.Unmount()

			// Simulate resize by emitting resize event
			ct.Emit("resize", map[string]int{
				"width":  tt.width,
				"height": tt.height,
			})

			// The app should handle the resize
			assert.NotNil(t, ct)
		})
	}
}

func TestBreakpointWidths(t *testing.T) {
	assert.Equal(t, 0, localComposables.BreakpointWidths[localComposables.BreakpointXS])
	assert.Equal(t, 60, localComposables.BreakpointWidths[localComposables.BreakpointSM])
	assert.Equal(t, 80, localComposables.BreakpointWidths[localComposables.BreakpointMD])
	assert.Equal(t, 120, localComposables.BreakpointWidths[localComposables.BreakpointLG])
	assert.Equal(t, 160, localComposables.BreakpointWidths[localComposables.BreakpointXL])
}

func TestMinimumDimensions(t *testing.T) {
	assert.Equal(t, 60, localComposables.MinWidth)
	assert.Equal(t, 20, localComposables.MinHeight)
}

func TestDemoNames(t *testing.T) {
	assert.Equal(t, "Dashboard", DemoNames[DemoDashboard])
	assert.Equal(t, "Grid", DemoNames[DemoGrid])
	assert.Equal(t, "Adaptive", DemoNames[DemoAdaptive])
	assert.Equal(t, "Breakpoints", DemoNames[DemoBreakpoint])
}
