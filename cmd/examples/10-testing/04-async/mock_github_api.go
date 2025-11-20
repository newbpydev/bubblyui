package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Repository represents a GitHub repository
type Repository struct {
	Name        string
	Stars       int
	Language    string
	Description string
	UpdatedAt   time.Time
}

// Activity represents a GitHub activity event
type Activity struct {
	Type      string // "push", "pr", "issue", "star"
	Repo      string
	Message   string
	Timestamp time.Time
}

// GitHubAPI defines the interface for GitHub operations
type GitHubAPI interface {
	FetchRepositories(username string) tea.Cmd
	FetchActivity(username string) tea.Cmd
}

// MockGitHubAPI provides a mock implementation for testing
type MockGitHubAPI struct {
	repos      []Repository
	activity   []Activity
	repoDelay  time.Duration
	actDelay   time.Duration
	shouldFail bool
}

// NewMockGitHubAPI creates a new mock API with default data
func NewMockGitHubAPI() *MockGitHubAPI {
	return &MockGitHubAPI{
		repos: []Repository{
			{
				Name:        "bubblyui",
				Stars:       142,
				Language:    "Go",
				Description: "Vue-inspired TUI framework for Go",
				UpdatedAt:   time.Now().Add(-2 * time.Hour),
			},
			{
				Name:        "awesome-tui",
				Stars:       89,
				Language:    "Go",
				Description: "Collection of awesome TUI libraries",
				UpdatedAt:   time.Now().Add(-5 * time.Hour),
			},
			{
				Name:        "cli-tools",
				Stars:       56,
				Language:    "Go",
				Description: "Useful CLI utilities",
				UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
			},
		},
		activity: []Activity{
			{
				Type:      "push",
				Repo:      "bubblyui",
				Message:   "Add async testing utilities",
				Timestamp: time.Now().Add(-30 * time.Minute),
			},
			{
				Type:      "pr",
				Repo:      "awesome-tui",
				Message:   "Merged PR #42: Add new library",
				Timestamp: time.Now().Add(-2 * time.Hour),
			},
			{
				Type:      "issue",
				Repo:      "bubblyui",
				Message:   "Opened issue #15: Feature request",
				Timestamp: time.Now().Add(-4 * time.Hour),
			},
			{
				Type:      "star",
				Repo:      "cli-tools",
				Message:   "Starred repository",
				Timestamp: time.Now().Add(-6 * time.Hour),
			},
		},
		repoDelay:  100 * time.Millisecond,
		actDelay:   100 * time.Millisecond,
		shouldFail: false,
	}
}

// SetDelay sets the delay for API responses (for testing)
func (m *MockGitHubAPI) SetDelay(repoDelay, actDelay time.Duration) {
	m.repoDelay = repoDelay
	m.actDelay = actDelay
}

// SetShouldFail makes the API return errors (for testing)
func (m *MockGitHubAPI) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// SetRepositories sets custom repository data (for testing)
func (m *MockGitHubAPI) SetRepositories(repos []Repository) {
	m.repos = repos
}

// SetActivity sets custom activity data (for testing)
func (m *MockGitHubAPI) SetActivity(activity []Activity) {
	m.activity = activity
}

// ReposFetchedMsg is sent when repositories are fetched
type ReposFetchedMsg struct {
	Repos []Repository
	Error error
}

// ActivityFetchedMsg is sent when activity is fetched
type ActivityFetchedMsg struct {
	Activity []Activity
	Error    error
}

// FetchRepositories fetches repositories asynchronously
func (m *MockGitHubAPI) FetchRepositories(username string) tea.Cmd {
	return func() tea.Msg {
		// Simulate network delay
		time.Sleep(m.repoDelay)

		if m.shouldFail {
			return ReposFetchedMsg{
				Error: fmt.Errorf("failed to fetch repositories for %s", username),
			}
		}

		return ReposFetchedMsg{
			Repos: m.repos,
			Error: nil,
		}
	}
}

// FetchActivity fetches activity asynchronously
func (m *MockGitHubAPI) FetchActivity(username string) tea.Cmd {
	return func() tea.Msg {
		// Simulate network delay
		time.Sleep(m.actDelay)

		if m.shouldFail {
			return ActivityFetchedMsg{
				Error: fmt.Errorf("failed to fetch activity for %s", username),
			}
		}

		return ActivityFetchedMsg{
			Activity: m.activity,
			Error:    nil,
		}
	}
}

// RealGitHubAPI would implement actual GitHub API calls
// This is a placeholder for production use
type RealGitHubAPI struct {
	token string
}

// NewRealGitHubAPI creates a real GitHub API client
func NewRealGitHubAPI(token string) *RealGitHubAPI {
	return &RealGitHubAPI{token: token}
}

// FetchRepositories would fetch real repositories
func (r *RealGitHubAPI) FetchRepositories(username string) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement real GitHub API call
		return ReposFetchedMsg{
			Error: fmt.Errorf("real GitHub API not implemented yet"),
		}
	}
}

// FetchActivity would fetch real activity
func (r *RealGitHubAPI) FetchActivity(username string) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement real GitHub API call
		return ActivityFetchedMsg{
			Error: fmt.Errorf("real GitHub API not implemented yet"),
		}
	}
}
