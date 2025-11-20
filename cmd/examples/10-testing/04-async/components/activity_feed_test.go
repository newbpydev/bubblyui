package components

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestActivityFeed_BasicMounting tests component initialization and mounting
func TestActivityFeed_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	activityRef := bubbly.NewRef[interface{}]([]composables.Activity{})
	loadingRef := bubbly.NewRef[interface{}](false)

	// Act: Create and mount component
	feed, err := CreateActivityFeed(ActivityFeedProps{
		Activity: activityRef,
		Loading:  loadingRef,
		Width:    60,
	})
	require.NoError(t, err)

	ct := harness.Mount(feed)

	// Assert: Component renders
	ct.AssertRenderContains("📊 Recent Activity")
}

// TestActivityFeed_LoadingState tests loading state display
func TestActivityFeed_LoadingState(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	activityRef := bubbly.NewRef[interface{}]([]composables.Activity{})
	loadingRef := bubbly.NewRef[interface{}](true)

	feed, err := CreateActivityFeed(ActivityFeedProps{
		Activity: activityRef,
		Loading:  loadingRef,
		Width:    60,
	})
	require.NoError(t, err)

	ct := harness.Mount(feed)

	// Assert: Loading message displayed
	ct.AssertRenderContains("Loading activity...")
}

// TestActivityFeed_EmptyState tests empty state display
func TestActivityFeed_EmptyState(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	activityRef := bubbly.NewRef[interface{}]([]composables.Activity{})
	loadingRef := bubbly.NewRef[interface{}](false)

	feed, err := CreateActivityFeed(ActivityFeedProps{
		Activity: activityRef,
		Loading:  loadingRef,
		Width:    60,
	})
	require.NoError(t, err)

	ct := harness.Mount(feed)

	// Assert: Empty message displayed
	ct.AssertRenderContains("No recent activity")
}

// TestActivityFeed_WithActivity tests activity display
func TestActivityFeed_WithActivity(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	activities := []composables.Activity{
		{Type: "push", Repo: "bubblyui", Message: "Add async testing", Timestamp: "2024-01-01 12:00"},
		{Type: "pr", Repo: "awesome-tui", Message: "Merged PR #42", Timestamp: "2024-01-01 13:00"},
		{Type: "issue", Repo: "bubblyui", Message: "Opened issue #15", Timestamp: "2024-01-01 14:00"},
		{Type: "star", Repo: "cli-tools", Message: "Starred repository", Timestamp: "2024-01-01 15:00"},
	}
	activityRef := bubbly.NewRef[interface{}](activities)
	loadingRef := bubbly.NewRef[interface{}](false)

	feed, err := CreateActivityFeed(ActivityFeedProps{
		Activity: activityRef,
		Loading:  loadingRef,
		Width:    60,
	})
	require.NoError(t, err)

	ct := harness.Mount(feed)

	// Assert: Activities displayed
	ct.AssertRenderContains("PUSH")
	ct.AssertRenderContains("PR")
	ct.AssertRenderContains("ISSUE")
	ct.AssertRenderContains("STAR")
	ct.AssertRenderContains("bubblyui")
	ct.AssertRenderContains("awesome-tui")
	ct.AssertRenderContains("Add async testing")
	ct.AssertRenderContains("Merged PR #42")
}

// TestActivityFeed_ActivityIcons tests different activity type icons
func TestActivityFeed_ActivityIcons(t *testing.T) {
	tests := []struct {
		name         string
		activityType string
		expectedIcon string
	}{
		{"push", "push", "📝"},
		{"pr", "pr", "🔀"},
		{"issue", "issue", "🐛"},
		{"star", "star", "⭐"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)

			activities := []composables.Activity{
				{Type: tt.activityType, Repo: "test-repo", Message: "Test message", Timestamp: "2024-01-01"},
			}
			activityRef := bubbly.NewRef[interface{}](activities)
			loadingRef := bubbly.NewRef[interface{}](false)

			feed, err := CreateActivityFeed(ActivityFeedProps{
				Activity: activityRef,
				Loading:  loadingRef,
				Width:    60,
			})
			require.NoError(t, err)

			ct := harness.Mount(feed)

			// Assert: Icon displayed
			ct.AssertRenderContains(tt.expectedIcon)
		})
	}
}

// TestActivityFeed_StateChange tests reactive state updates
func TestActivityFeed_StateChange(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	activityRef := bubbly.NewRef[interface{}]([]composables.Activity{})
	loadingRef := bubbly.NewRef[interface{}](true)

	feed, err := CreateActivityFeed(ActivityFeedProps{
		Activity: activityRef,
		Loading:  loadingRef,
		Width:    60,
	})
	require.NoError(t, err)

	ct := harness.Mount(feed)

	// Assert: Initially loading
	ct.AssertRenderContains("Loading activity...")

	// Act: Update to loaded state with data
	loadingRef.Set(false)
	activities := []composables.Activity{
		{Type: "push", Repo: "test-repo", Message: "Test commit", Timestamp: "2024-01-01"},
	}
	activityRef.Set(activities)

	// Assert: Activity displayed
	ct.AssertRenderContains("test-repo")
	ct.AssertRenderContains("Test commit")
}
