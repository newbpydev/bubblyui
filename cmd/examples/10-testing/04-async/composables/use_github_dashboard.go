package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Repository represents a GitHub repository
type Repository struct {
	Name        string
	Stars       int
	Language    string
	Description string
}

// Activity represents a GitHub activity event
type Activity struct {
	Type      string
	Repo      string
	Message   string
	Timestamp string
}

// GitHubAPI defines the interface for GitHub operations
// This is a synchronous interface - composables use goroutines, not tea.Cmd
type GitHubAPI interface {
	FetchRepositories(username string) ([]Repository, error)
	FetchActivity(username string) ([]Activity, error)
}

// GitHubDashboardComposable provides reactive GitHub dashboard management
type GitHubDashboardComposable struct {
	// State - interface{} refs (from ctx.Ref) for auto-command tracking
	Username        *bubbly.Ref[string]
	Repositories    *bubbly.Ref[interface{}] // []Repository - tracked by WithAutoCommands
	Activity        *bubbly.Ref[interface{}] // []Activity - tracked by WithAutoCommands
	LoadingRepos    *bubbly.Ref[interface{}] // bool - tracked by WithAutoCommands
	LoadingActivity *bubbly.Ref[interface{}] // bool - tracked by WithAutoCommands
	Error           *bubbly.Ref[interface{}] // string - tracked by WithAutoCommands
	LastRefresh     *bubbly.Ref[interface{}] // time.Time - tracked by WithAutoCommands

	// API client
	api GitHubAPI

	// Methods (no tea.Cmd - composables use goroutines)
	FetchRepos        func()
	FetchActivity     func()
	Refresh           func()
	HandleReposMsg    func(repos []Repository, err error)
	HandleActivityMsg func(activity []Activity, err error)
}

// UseGitHubDashboard creates a GitHub dashboard composable
// Manages async data fetching, loading states, and error handling
func UseGitHubDashboard(ctx *bubbly.Context, username string, api GitHubAPI) *GitHubDashboardComposable {
	// State refs - use ctx.Ref for tracked reactive state (triggers re-renders)
	usernameRef := bubbly.NewRef(username) // Username doesn't change, standalone is fine
	repositories := ctx.Ref(make([]Repository, 0))
	activity := ctx.Ref(make([]Activity, 0))
	loadingRepos := ctx.Ref(false)
	loadingActivity := ctx.Ref(false)
	errorRef := ctx.Ref("")
	lastRefresh := ctx.Ref(interface{}(nil))

	// Fetch repositories (uses goroutine like UseAsync)
	fetchRepos := func() {
		loadingRepos.Set(true)
		errorRef.Set("")

		go func() {
			// Small delay to ensure template rendering completes before Ref.Set()
			// This avoids race condition with inTemplate flag during initialization
			time.Sleep(10 * time.Millisecond)
			repos, err := api.FetchRepositories(usernameRef.Get().(string))
			loadingRepos.Set(false)

			if err != nil {
				errorRef.Set(err.Error())
			} else {
				repositories.Set(repos)
			}
		}()
	}

	// Fetch activity (uses goroutine like UseAsync)
	fetchActivity := func() {
		loadingActivity.Set(true)
		errorRef.Set("")

		go func() {
			// Small delay to ensure template rendering completes before Ref.Set()
			// This avoids race condition with inTemplate flag during initialization
			time.Sleep(10 * time.Millisecond)
			acts, err := api.FetchActivity(usernameRef.Get().(string))
			loadingActivity.Set(false)

			if err != nil {
				errorRef.Set(err.Error())
			} else {
				activity.Set(acts)
			}
		}()
	}

	// Refresh all data
	refresh := func() {
		fetchRepos()
		fetchActivity()
	}

	// Handle repos message
	handleReposMsg := func(repos []Repository, err error) {
		loadingRepos.Set(false)
		if err != nil {
			errorRef.Set(err.Error())
		} else {
			repositories.Set(repos)
		}
	}

	// Handle activity message
	handleActivityMsg := func(act []Activity, err error) {
		loadingActivity.Set(false)
		if err != nil {
			errorRef.Set(err.Error())
		} else {
			activity.Set(act)
		}
	}

	return &GitHubDashboardComposable{
		Username:          usernameRef,
		Repositories:      repositories,
		Activity:          activity,
		LoadingRepos:      loadingRepos,
		LoadingActivity:   loadingActivity,
		Error:             errorRef,
		LastRefresh:       lastRefresh,
		api:               api,
		FetchRepos:        fetchRepos,
		FetchActivity:     fetchActivity,
		Refresh:           refresh,
		HandleReposMsg:    handleReposMsg,
		HandleActivityMsg: handleActivityMsg,
	}
}
