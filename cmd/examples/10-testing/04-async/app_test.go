package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestGitHubDashboard_BasicMounting tests app initialization
func TestGitHubDashboard_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(0, 0) // No delay for tests

	// Act
	app, err := CreateApp(mockAPI)
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: App renders with title
	ct.AssertRenderContains("🚀 GitHub Dashboard")
	ct.AssertRenderContains("newbpydev")
}

// TestGitHubDashboard_InitialState tests initial loading state
func TestGitHubDashboard_InitialState(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(0, 0)

	app, err := CreateApp(mockAPI)
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: Components render
	ct.AssertRenderContains("📦 Repositories")
	ct.AssertRenderContains("📊 Recent Activity")
	ct.AssertRenderContains("Press [r] to refresh")
}

// TestGitHubDashboard_DataLoading tests that components render (async data loading tested in component tests)
func TestGitHubDashboard_DataLoading(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(0, 0)

	app, err := CreateApp(mockAPI)
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: Components render (actual data loading tested in component/composable tests)
	ct.AssertRenderContains("📦 Repositories")
	ct.AssertRenderContains("📊 Recent Activity")
}

// TestGitHubDashboard_ErrorHandling tests that app renders with error API
func TestGitHubDashboard_ErrorHandling(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(0, 0)
	mockAPI.SetShouldFail(true) // Force errors

	app, err := CreateApp(mockAPI)
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: App still renders (error handling tested in composable tests)
	ct.AssertRenderContains("📦 Repositories")
	ct.AssertRenderContains("📊 Recent Activity")
}

// TestGitHubDashboard_RefreshCommand tests refresh key binding
func TestGitHubDashboard_RefreshCommand(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(0, 0)

	app, err := CreateApp(mockAPI)
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Act: Trigger refresh with 'r' key (should not crash)
	ct.SendKey("r")

	// Assert: App still renders after refresh
	ct.AssertRenderContains("📦 Repositories")
	ct.AssertRenderContains("📊 Recent Activity")
}

// TestGitHubDashboard_RepositoryDisplay tests repository component integration
func TestGitHubDashboard_RepositoryDisplay(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(0, 0)

	app, err := CreateApp(mockAPI)
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: Repository component renders (data display tested in component tests)
	ct.AssertRenderContains("📦 Repositories")
}

// TestGitHubDashboard_ActivityDisplay tests activity component integration
func TestGitHubDashboard_ActivityDisplay(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(0, 0)

	app, err := CreateApp(mockAPI)
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: Activity component renders (data display tested in component tests)
	ct.AssertRenderContains("📊 Recent Activity")
}

// TestGitHubDashboard_HelpText tests help text display
func TestGitHubDashboard_HelpText(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	defer harness.Cleanup()

	mockAPI := NewMockGitHubAPI()
	mockAPI.SetDelay(0, 0)

	app, err := CreateApp(mockAPI)
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: Help text displayed
	ct.AssertRenderContains("Press [r] to refresh")
	ct.AssertRenderContains("[ctrl+c] to quit")
}
