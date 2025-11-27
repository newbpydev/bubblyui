package composables

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// createTestContext creates a minimal context for direct composable testing
func createTestContext() *bubbly.Context {
	var ctx *bubbly.Context
	component, _ := bubbly.NewComponent("Test").
		Setup(func(c *bubbly.Context) {
			ctx = c
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()
	// CRITICAL: Must call Init() to execute Setup and get the context
	component.Init()
	return ctx
}

// MockGitHubAPI for testing
type MockGitHubAPI struct {
	repos      []Repository
	activity   []Activity
	shouldFail bool
}

func NewMockAPI() *MockGitHubAPI {
	return &MockGitHubAPI{
		repos: []Repository{
			{Name: "repo1", Stars: 10, Language: "Go", Description: "Test repo 1"},
			{Name: "repo2", Stars: 20, Language: "Go", Description: "Test repo 2"},
		},
		activity: []Activity{
			{Type: "push", Repo: "repo1", Message: "Test push", Timestamp: "2024-01-01 12:00"},
			{Type: "pr", Repo: "repo2", Message: "Test PR", Timestamp: "2024-01-01 13:00"},
		},
		shouldFail: false,
	}
}

func (m *MockGitHubAPI) SetShouldFail(fail bool) {
	m.shouldFail = fail
}

func (m *MockGitHubAPI) FetchRepositories(username string) ([]Repository, error) {
	if m.shouldFail {
		return nil, errors.New("failed to fetch repos")
	}
	return m.repos, nil
}

func (m *MockGitHubAPI) FetchActivity(username string) ([]Activity, error) {
	if m.shouldFail {
		return nil, errors.New("failed to fetch activity")
	}
	return m.activity, nil
}

// TestUseGitHubDashboard_Initialization tests initial state
func TestUseGitHubDashboard_Initialization(t *testing.T) {
	ctx := createTestContext()
	mockAPI := NewMockAPI()
	dashboard := UseGitHubDashboard(ctx, "testuser", mockAPI)

	// Check initial state
	assert.Equal(t, "testuser", dashboard.Username.Get().(string))
	assert.Equal(t, 0, len(dashboard.Repositories.Get().([]Repository)))
	assert.Equal(t, 0, len(dashboard.Activity.Get().([]Activity)))
	assert.False(t, dashboard.LoadingRepos.Get().(bool))
	assert.False(t, dashboard.LoadingActivity.Get().(bool))
	assert.Equal(t, "", dashboard.Error.Get().(string))
}

// TestUseGitHubDashboard_FetchRepos tests repository fetching
func TestUseGitHubDashboard_FetchRepos(t *testing.T) {
	ctx := createTestContext()
	mockAPI := NewMockAPI()
	dashboard := UseGitHubDashboard(ctx, "testuser", mockAPI)

	// Fetch repos (triggers goroutine)
	dashboard.FetchRepos()

	// Should set loading state immediately
	assert.True(t, dashboard.LoadingRepos.Get().(bool))
	assert.Equal(t, "", dashboard.Error.Get().(string))
}

// TestUseGitHubDashboard_HandleReposMsg_Success tests successful repo fetch
func TestUseGitHubDashboard_HandleReposMsg_Success(t *testing.T) {
	ctx := createTestContext()
	mockAPI := NewMockAPI()
	dashboard := UseGitHubDashboard(ctx, "testuser", mockAPI)

	// Set loading state
	dashboard.LoadingRepos.Set(true)

	// Handle success message
	repos := []Repository{
		{Name: "repo1", Stars: 10, Language: "Go", Description: "Test"},
	}
	dashboard.HandleReposMsg(repos, nil)

	// Check state
	assert.False(t, dashboard.LoadingRepos.Get().(bool))
	assert.Equal(t, 1, len(dashboard.Repositories.Get().([]Repository)))
	assert.Equal(t, "repo1", dashboard.Repositories.Get().([]Repository)[0].Name)
	assert.Equal(t, "", dashboard.Error.Get().(string))
}

// TestUseGitHubDashboard_HandleReposMsg_Error tests error handling
func TestUseGitHubDashboard_HandleReposMsg_Error(t *testing.T) {
	ctx := createTestContext()
	mockAPI := NewMockAPI()
	dashboard := UseGitHubDashboard(ctx, "testuser", mockAPI)

	// Set loading state
	dashboard.LoadingRepos.Set(true)

	// Handle error message
	err := errors.New("fetch failed")
	dashboard.HandleReposMsg(nil, err)

	// Check state
	assert.False(t, dashboard.LoadingRepos.Get().(bool))
	assert.Equal(t, "fetch failed", dashboard.Error.Get().(string))
}

// TestUseGitHubDashboard_FetchActivity tests activity fetching
func TestUseGitHubDashboard_FetchActivity(t *testing.T) {
	ctx := createTestContext()
	mockAPI := NewMockAPI()
	dashboard := UseGitHubDashboard(ctx, "testuser", mockAPI)

	// Fetch activity (triggers goroutine)
	dashboard.FetchActivity()

	// Should set loading state immediately
	assert.True(t, dashboard.LoadingActivity.Get().(bool))
	assert.Equal(t, "", dashboard.Error.Get().(string))
}

// TestUseGitHubDashboard_HandleActivityMsg_Success tests successful activity fetch
func TestUseGitHubDashboard_HandleActivityMsg_Success(t *testing.T) {
	ctx := createTestContext()
	mockAPI := NewMockAPI()
	dashboard := UseGitHubDashboard(ctx, "testuser", mockAPI)

	// Set loading state
	dashboard.LoadingActivity.Set(true)

	// Handle success message
	activities := []Activity{
		{Type: "push", Repo: "repo1", Message: "Test", Timestamp: "2024-01-01"},
	}
	dashboard.HandleActivityMsg(activities, nil)

	// Check state
	assert.False(t, dashboard.LoadingActivity.Get().(bool))
	assert.Equal(t, 1, len(dashboard.Activity.Get().([]Activity)))
	assert.Equal(t, "push", dashboard.Activity.Get().([]Activity)[0].Type)
	assert.Equal(t, "", dashboard.Error.Get().(string))
}

// TestUseGitHubDashboard_HandleActivityMsg_Error tests error handling
func TestUseGitHubDashboard_HandleActivityMsg_Error(t *testing.T) {
	ctx := createTestContext()
	mockAPI := NewMockAPI()
	dashboard := UseGitHubDashboard(ctx, "testuser", mockAPI)

	// Set loading state
	dashboard.LoadingActivity.Set(true)

	// Handle error message
	err := errors.New("fetch failed")
	dashboard.HandleActivityMsg(nil, err)

	// Check state
	assert.False(t, dashboard.LoadingActivity.Get().(bool))
	assert.Equal(t, "fetch failed", dashboard.Error.Get().(string))
}

// TestUseGitHubDashboard_Refresh tests refresh functionality
func TestUseGitHubDashboard_Refresh(t *testing.T) {
	ctx := createTestContext()
	mockAPI := NewMockAPI()
	dashboard := UseGitHubDashboard(ctx, "testuser", mockAPI)

	// Refresh (triggers both goroutines)
	dashboard.Refresh()

	// Should set both loading states immediately
	assert.True(t, dashboard.LoadingRepos.Get().(bool))
	assert.True(t, dashboard.LoadingActivity.Get().(bool))
}
