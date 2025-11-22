package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

func TestApp_Creation(t *testing.T) {
	app, err := CreateApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, "SharedStateApp", app.Name())
}

func TestApp_SharedStateSync(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	defer ct.Unmount()

	// Initial render should show 0
	ct.AssertRenderContains("Current Value:")
	ct.AssertRenderContains("0")

	// Increment via event
	ct.Emit("increment", nil)

	// Both display and controls should show 1
	ct.AssertRenderContains("1")

	// Increment again
	ct.Emit("increment", nil)
	ct.AssertRenderContains("2")

	// Decrement
	ct.Emit("decrement", nil)
	ct.AssertRenderContains("1")

	// Reset
	ct.Emit("reset", nil)
	ct.AssertRenderContains("0")
}

func TestApp_HistoryTracking(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	defer ct.Unmount()

	// Initial history
	ct.AssertRenderContains("History:")
	ct.AssertRenderContains("0")

	// Increment multiple times
	ct.Emit("increment", nil)
	ct.Emit("increment", nil)
	ct.Emit("increment", nil)

	// Should show history
	ct.AssertRenderContains("→")
}

func TestApp_ComputedValues(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	defer ct.Unmount()

	// Reset to known state first
	ct.Emit("reset", nil)

	// Initial: 0 is even
	ct.AssertRenderContains("Is Even:")
	ct.AssertRenderContains("✓ Yes")

	// Increment to 1 (odd)
	ct.Emit("increment", nil)
	ct.AssertRenderContains("✗ No")

	// Doubled value should be 2
	ct.AssertRenderContains("Doubled:")
	ct.AssertRenderContains("Doubled: 2")
}

func TestApp_KeyBindings(t *testing.T) {
	app, err := CreateApp()
	require.NoError(t, err)

	bindings := app.KeyBindings()
	assert.NotEmpty(t, bindings)

	// Check multi-key bindings
	assert.Contains(t, bindings, "up")
	assert.Contains(t, bindings, "k")
	assert.Contains(t, bindings, "+")
	assert.Contains(t, bindings, "down")
	assert.Contains(t, bindings, "j")
	assert.Contains(t, bindings, "-")
	assert.Contains(t, bindings, "r")
	assert.Contains(t, bindings, "q")
}
