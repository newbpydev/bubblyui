package components

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestRepoList_BasicMounting tests component initialization and mounting
func TestRepoList_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	reposRef := bubbly.NewRef[interface{}]([]composables.Repository{})
	loadingRef := bubbly.NewRef[interface{}](false)

	// Act: Create and mount component
	repoList, err := CreateRepoList(RepoListProps{
		Repositories: reposRef,
		Loading:      loadingRef,
		Width:        60,
	})
	require.NoError(t, err)

	ct := harness.Mount(repoList)

	// Assert: Component renders
	ct.AssertRenderContains("📦 Repositories")
}

// TestRepoList_LoadingState tests loading state display
func TestRepoList_LoadingState(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	reposRef := bubbly.NewRef[interface{}]([]composables.Repository{})
	loadingRef := bubbly.NewRef[interface{}](true)

	repoList, err := CreateRepoList(RepoListProps{
		Repositories: reposRef,
		Loading:      loadingRef,
		Width:        60,
	})
	require.NoError(t, err)

	ct := harness.Mount(repoList)

	// Assert: Loading message displayed
	ct.AssertRenderContains("Loading repositories...")
}

// TestRepoList_EmptyState tests empty state display
func TestRepoList_EmptyState(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	reposRef := bubbly.NewRef[interface{}]([]composables.Repository{})
	loadingRef := bubbly.NewRef[interface{}](false)

	repoList, err := CreateRepoList(RepoListProps{
		Repositories: reposRef,
		Loading:      loadingRef,
		Width:        60,
	})
	require.NoError(t, err)

	ct := harness.Mount(repoList)

	// Assert: Empty message displayed
	ct.AssertRenderContains("No repositories found")
}

// TestRepoList_WithRepositories tests repository display
func TestRepoList_WithRepositories(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	repos := []composables.Repository{
		{Name: "bubblyui", Stars: 142, Language: "Go", Description: "TUI framework"},
		{Name: "awesome-tui", Stars: 89, Language: "Go", Description: "TUI collection"},
	}
	reposRef := bubbly.NewRef[interface{}](repos)
	loadingRef := bubbly.NewRef[interface{}](false)

	repoList, err := CreateRepoList(RepoListProps{
		Repositories: reposRef,
		Loading:      loadingRef,
		Width:        60,
	})
	require.NoError(t, err)

	ct := harness.Mount(repoList)

	// Assert: Repositories displayed
	ct.AssertRenderContains("bubblyui")
	ct.AssertRenderContains("awesome-tui")
	ct.AssertRenderContains("⭐ 142")
	ct.AssertRenderContains("⭐ 89")
	ct.AssertRenderContains("Go")
}

// TestRepoList_StateChange tests reactive state updates
func TestRepoList_StateChange(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	reposRef := bubbly.NewRef[interface{}]([]composables.Repository{})
	loadingRef := bubbly.NewRef[interface{}](true)

	repoList, err := CreateRepoList(RepoListProps{
		Repositories: reposRef,
		Loading:      loadingRef,
		Width:        60,
	})
	require.NoError(t, err)

	ct := harness.Mount(repoList)

	// Assert: Initially loading
	ct.AssertRenderContains("Loading repositories...")

	// Act: Update to loaded state with data
	loadingRef.Set(false)
	repos := []composables.Repository{
		{Name: "test-repo", Stars: 50, Language: "Go", Description: "Test"},
	}
	reposRef.Set(repos)

	// Assert: Repositories displayed
	ct.AssertRenderContains("test-repo")
	ct.AssertRenderContains("⭐ 50")
}
