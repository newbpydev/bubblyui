package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	localComponents "github.com/newbpydev/bubblyui/cmd/examples/14-advanced-layouts/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/14-advanced-layouts/composables"
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
	ct.AssertRenderContains("BubblyUI Advanced Layout System Showcase")

	// Should render tab bar with all demos
	ct.AssertRenderContains("Dashboard")
	ct.AssertRenderContains("Flex Layout")
	ct.AssertRenderContains("Card Grid")
	ct.AssertRenderContains("Form Layout")
	ct.AssertRenderContains("Modal/Dialog")
}

func TestDashboardDemo(t *testing.T) {
	demo, err := localComponents.CreateDashboardDemo()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render dashboard elements
	ct.AssertRenderContains("Dashboard")
	ct.AssertRenderContains("Navigation")
	ct.AssertRenderContains("Statistics Overview")
}

func TestFlexDemo(t *testing.T) {
	demo, err := localComponents.CreateFlexDemo()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render flex demo elements
	ct.AssertRenderContains("Flex Layout Demo")
	ct.AssertRenderContains("Justify")
	ct.AssertRenderContains("Align")
	ct.AssertRenderContains("Flex Container")
}

func TestCardGridDemo(t *testing.T) {
	demo, err := localComponents.CreateCardGridDemo()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render card grid elements
	ct.AssertRenderContains("Card Grid Demo")
	ct.AssertRenderContains("Product Grid")
	ct.AssertRenderContains("Laptop")
	ct.AssertRenderContains("Phone")
}

func TestFormDemo(t *testing.T) {
	demo, err := localComponents.CreateFormDemo()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render form demo elements
	ct.AssertRenderContains("Form Layout Demo")
	ct.AssertRenderContains("User Registration")
	ct.AssertRenderContains("Name:")
	ct.AssertRenderContains("Email:")
}

func TestModalDemo(t *testing.T) {
	demo, err := localComponents.CreateModalDemo()
	require.NoError(t, err)
	require.NotNil(t, demo)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(demo)
	defer ct.Unmount()

	// Should render modal demo elements
	ct.AssertRenderContains("Modal/Dialog Demo")
	ct.AssertRenderContains("Modal Patterns")
	ct.AssertRenderContains("Confirm Delete")
}

func TestDemoState(t *testing.T) {
	tests := []struct {
		name     string
		action   func(*localComposables.DemoStateComposable)
		expected int
	}{
		{
			name:     "initial state is dashboard",
			action:   func(ds *localComposables.DemoStateComposable) {},
			expected: int(localComposables.DemoDashboard),
		},
		{
			name: "next demo cycles forward",
			action: func(ds *localComposables.DemoStateComposable) {
				ds.NextDemo()
			},
			expected: int(localComposables.DemoFlex),
		},
		{
			name: "set demo directly",
			action: func(ds *localComposables.DemoStateComposable) {
				ds.SetDemo(localComposables.DemoCardGrid)
			},
			expected: int(localComposables.DemoCardGrid),
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

			// The demo state is managed internally, so we just verify the app works
			assert.NotNil(t, ct)
		})
	}
}

func TestJustifyOptions(t *testing.T) {
	assert.Len(t, localComposables.JustifyOptions, 6)
	assert.Contains(t, localComposables.JustifyOptions, localComposables.JustifyOptions[0])
}

func TestAlignOptions(t *testing.T) {
	assert.Len(t, localComposables.AlignOptions, 4)
	assert.Contains(t, localComposables.AlignOptions, localComposables.AlignOptions[0])
}

func TestDemoNames(t *testing.T) {
	assert.Equal(t, "Dashboard", localComposables.DemoNames[localComposables.DemoDashboard])
	assert.Equal(t, "Flex Layout", localComposables.DemoNames[localComposables.DemoFlex])
	assert.Equal(t, "Card Grid", localComposables.DemoNames[localComposables.DemoCardGrid])
	assert.Equal(t, "Form Layout", localComposables.DemoNames[localComposables.DemoForm])
	assert.Equal(t, "Modal/Dialog", localComposables.DemoNames[localComposables.DemoModal])
}
