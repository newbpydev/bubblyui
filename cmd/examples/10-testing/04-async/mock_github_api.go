package main

import (
	"fmt"
	"time"

	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/composables"
)

// MockGitHubAPI provides a synchronous mock implementation for testing
// This follows the UseAsync pattern - synchronous functions that block
type MockGitHubAPI struct {
	repos      []composables.Repository
	activity   []composables.Activity
	repoDelay  time.Duration
	actDelay   time.Duration
	shouldFail bool
}

// NewMockGitHubAPI creates a new mock API with default data
func NewMockGitHubAPI() *MockGitHubAPI {
	return &MockGitHubAPI{
		repos: []composables.Repository{
			{
				Name:        "bubblyui",
				Stars:       142,
				Language:    "Go",
				Description: "Vue-inspired TUI framework for Go",
			},
			{
				Name:        "go-patterns",
				Stars:       89,
				Language:    "Go",
				Description: "Collection of Go design patterns",
			},
			{
				Name:        "tui-toolkit",
				Stars:       56,
				Language:    "Go",
				Description: "Terminal UI component library",
			},
		},
		activity: []composables.Activity{
			{
				Type:      "push",
				Repo:      "bubblyui",
				Message:   "Added async composable support",
				Timestamp: time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04"),
			},
			{
				Type:      "pr",
				Repo:      "go-patterns",
				Message:   "Merged: Add observer pattern",
				Timestamp: time.Now().Add(-3 * time.Hour).Format("2006-01-02 15:04"),
			},
			{
				Type:      "issue",
				Repo:      "tui-toolkit",
				Message:   "Opened: Add table component",
				Timestamp: time.Now().Add(-5 * time.Hour).Format("2006-01-02 15:04"),
			},
		},
		repoDelay: 0,
		actDelay:  0,
	}
}

// FetchRepositories fetches repositories synchronously (blocks with delay)
func (m *MockGitHubAPI) FetchRepositories(username string) ([]composables.Repository, error) {
	// Simulate network delay
	if m.repoDelay > 0 {
		time.Sleep(m.repoDelay)
	}

	if m.shouldFail {
		return nil, fmt.Errorf("failed to fetch repositories")
	}

	return m.repos, nil
}

// FetchActivity fetches activity synchronously (blocks with delay)
func (m *MockGitHubAPI) FetchActivity(username string) ([]composables.Activity, error) {
	// Simulate network delay
	if m.actDelay > 0 {
		time.Sleep(m.actDelay)
	}

	if m.shouldFail {
		return nil, fmt.Errorf("failed to fetch activity")
	}

	return m.activity, nil
}

// SetDelay sets the simulated network delay
func (m *MockGitHubAPI) SetDelay(repoDelay, actDelay time.Duration) {
	m.repoDelay = repoDelay
	m.actDelay = actDelay
}

// SetShouldFail sets whether API calls should fail
func (m *MockGitHubAPI) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}
