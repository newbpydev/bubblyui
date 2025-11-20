package composables

import (
	tea "github.com/charmbracelet/bubbletea"
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
type GitHubAPI interface {
	FetchRepositories(username string) tea.Cmd
	FetchActivity(username string) tea.Cmd
}

// GitHubDashboardComposable provides reactive GitHub dashboard management
type GitHubDashboardComposable struct {
	// State
	Username        *bubbly.Ref[string]
	Repositories    *bubbly.Ref[interface{}] // []Repository
	Activity        *bubbly.Ref[interface{}] // []Activity
	LoadingRepos    *bubbly.Ref[interface{}] // bool
	LoadingActivity *bubbly.Ref[interface{}] // bool
	Error           *bubbly.Ref[interface{}] // string
	LastRefresh     *bubbly.Ref[interface{}] // time.Time

	// API client
	api GitHubAPI

	// Methods
	FetchRepos        func() tea.Cmd
	FetchActivity     func() tea.Cmd
	Refresh           func() tea.Cmd
	HandleReposMsg    func(repos []Repository, err error)
	HandleActivityMsg func(activity []Activity, err error)
}

// UseGitHubDashboard creates a GitHub dashboard composable
// Manages async data fetching, loading states, and error handling
func UseGitHubDashboard(ctx *bubbly.Context, username string, api GitHubAPI) *GitHubDashboardComposable {
	// State refs - use bubbly.NewRef for typed refs
	usernameRef := bubbly.NewRef(username)
	repositories := ctx.Ref(make([]Repository, 0))
	activity := ctx.Ref(make([]Activity, 0))
	loadingRepos := ctx.Ref(false)
	loadingActivity := ctx.Ref(false)
	errorRef := ctx.Ref("")
	lastRefresh := ctx.Ref(interface{}(nil))

	// Fetch repositories
	fetchRepos := func() tea.Cmd {
		loadingRepos.Set(true)
		errorRef.Set("")
		return api.FetchRepositories(usernameRef.Get().(string))
	}

	// Fetch activity
	fetchActivity := func() tea.Cmd {
		loadingActivity.Set(true)
		errorRef.Set("")
		return api.FetchActivity(usernameRef.Get().(string))
	}

	// Refresh all data
	refresh := func() tea.Cmd {
		return tea.Batch(fetchRepos(), fetchActivity())
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
