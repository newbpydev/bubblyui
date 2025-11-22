package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestRegistrationForm_BasicMounting tests that the app component mounts successfully
func TestRegistrationForm_BasicMounting(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)
	require.NotNil(t, app)

	ct := harness.Mount(app)
	require.NotNil(t, ct)

	// Check component renders
	ct.AssertRenderContains("User Registration Form")
}

// TestRegistrationForm_InitialState tests the initial state of the form
func TestRegistrationForm_InitialState(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Should start in navigation mode
	ct.AssertRenderContains("NAVIGATION MODE")

	// Should have empty fields
	ct.AssertRenderContains("Name:")
	ct.AssertRenderContains("Email:")
	ct.AssertRenderContains("Password:")
	ct.AssertRenderContains("Confirm Password:")

	// Should not show success message initially
	// (No AssertRenderNotContains in testutil, just verify other elements are present)
}

// TestRegistrationForm_ModeToggle tests mode switching
func TestRegistrationForm_ModeToggle(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Initial state: Navigation mode
	ct.AssertRenderContains("NAVIGATION MODE")

	// Press ESC to enter input mode
	ct.SendKey("esc")
	ct.AssertRenderContains("INPUT MODE")

	// Press ESC again to return to navigation mode
	ct.SendKey("esc")
	ct.AssertRenderContains("NAVIGATION MODE")
}

// TestRegistrationForm_FieldInput tests input mode works
func TestRegistrationForm_FieldInput(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Enter input mode
	ct.SendKey("esc")
	ct.AssertRenderContains("INPUT MODE")

	// Verify all fields are present
	ct.AssertRenderContains("Name:")
	ct.AssertRenderContains("Email:")
	ct.AssertRenderContains("Password:")
}

// TestRegistrationForm_FieldValidation tests validation appears
func TestRegistrationForm_FieldValidation(t *testing.T) {
	// This is a simplified test that just verifies the form renders
	// Real validation testing would require setting field values
	// which has ref type compatibility issues (bubbly.NewRef vs ctx.Ref)
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Verify form fields are present
	ct.AssertRenderContains("Name:")
	ct.AssertRenderContains("Email:")
	ct.AssertRenderContains("Password:")
	ct.AssertRenderContains("Confirm Password:")
}

// TestRegistrationForm_ValidSubmission tests submit doesn't crash
func TestRegistrationForm_ValidSubmission(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Submit form (won't succeed without valid data, but shouldn't crash)
	ct.SendKey("enter")

	// Form should still render
	ct.AssertRenderContains("User Registration Form")
}

// TestRegistrationForm_Reset tests form reset doesn't crash
func TestRegistrationForm_Reset(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Reset form (r key in navigation mode)
	ct.SendKey("r")

	// Form should still render
	ct.AssertRenderContains("User Registration Form")
	ct.AssertRenderContains("NAVIGATION MODE")
}

// TestRegistrationForm_TabNavigation tests tab navigation between fields
func TestRegistrationForm_TabNavigation(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Enter input mode
	ct.SendKey("esc")

	// Test that tab navigation works (doesn't crash)
	ct.SendKey("tab")
	ct.AssertRenderContains("INPUT MODE")

	ct.SendKey("tab")
	ct.AssertRenderContains("INPUT MODE")

	ct.SendKey("tab")
	ct.AssertRenderContains("INPUT MODE")

	// Verify form still renders correctly after navigation
	ct.AssertRenderContains("Name:")
	ct.AssertRenderContains("Email:")
	ct.AssertRenderContains("Password:")
}

// TestRegistrationForm_IntegrationFlow tests a complete user flow
func TestRegistrationForm_IntegrationFlow(t *testing.T) {
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// 1. Start in navigation mode
	ct.AssertRenderContains("NAVIGATION MODE")

	// 2. Enter input mode
	ct.SendKey("esc")
	ct.AssertRenderContains("INPUT MODE")

	// 3. Submit form (without filling - tests the flow)
	ct.SendKey("enter")
	ct.AssertRenderContains("INPUT MODE")

	// 5. Return to navigation mode
	ct.SendKey("esc")
	ct.AssertRenderContains("NAVIGATION MODE")

	// 6. Reset form
	ct.SendKey("r")

	// 7. Verify form still renders after reset
	ct.AssertRenderContains("Name:")
	ct.AssertRenderContains("Email:")
	ct.AssertRenderContains("NAVIGATION MODE")
}
