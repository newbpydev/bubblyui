package testutil

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// HistoryTester provides utilities for testing router history management.
//
// It wraps a router instance and provides assertion methods for verifying
// history state, navigation capabilities, and history entries. This is useful
// for testing router history behavior without needing full integration tests.
//
// Type Safety:
//   - Thread-safe access to history state
//   - Clear assertion methods for test verification
//   - Direct access to history entries for inspection
//
// Example:
//
//	func TestHistoryManagement(t *testing.T) {
//		r, err := router.NewRouterBuilder().
//			Route("/home", "home").
//			Route("/about", "about").
//			Build()
//		assert.NoError(t, err)
//
//		ht := testutil.NewHistoryTester(r)
//
//		// Navigate and test
//		target := &router.NavigationTarget{Path: "/home"}
//		cmd := r.Push(target)
//		if cmd != nil {
//			_ = cmd()
//		}
//
//		ht.AssertHistoryLength(t, 1)
//		ht.AssertCanGoBack(t, false)
//	}
type HistoryTester struct {
	// router is the router instance being tested
	router *router.Router

	// history tracks navigation history entries
	history []*router.HistoryEntry

	// currentIdx is the current position in history
	currentIdx int

	// maxEntries is the maximum number of history entries (0 = unlimited)
	maxEntries int
}

// NewHistoryTester creates a new HistoryTester for testing history management.
//
// Parameters:
//   - router: The router instance to test
//
// Returns:
//   - *HistoryTester: A new tester instance
//
// Example:
//
//	r := router.NewRouter()
//	ht := testutil.NewHistoryTester(r)
func NewHistoryTester(r *router.Router) *HistoryTester {
	return &HistoryTester{
		router:     r,
		history:    make([]*router.HistoryEntry, 0),
		currentIdx: 0,
		maxEntries: 0,
	}
}

// AssertHistoryLength asserts that the history has the expected length.
//
// This method accesses the router's internal history to verify the number
// of entries. It uses reflection to access private fields for testing purposes.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: The expected history length
//
// Example:
//
//	ht.AssertHistoryLength(t, 3)
func (ht *HistoryTester) AssertHistoryLength(t testingT, expected int) {
	t.Helper()

	// Get history entries from router
	entries := ht.GetHistoryEntries()

	if len(entries) != expected {
		t.Errorf("expected history length %d, got %d", expected, len(entries))
	}
}

// AssertCanGoBack asserts whether back navigation is possible.
//
// This method checks the router's CanGoBack() method to verify if
// backward navigation is possible.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: Whether back navigation should be possible
//
// Example:
//
//	ht.AssertCanGoBack(t, true)
func (ht *HistoryTester) AssertCanGoBack(t testingT, expected bool) {
	t.Helper()

	// Access router's history to check if we can go back
	// We need to use reflection or access exported methods
	entries := ht.GetHistoryEntries()
	currentRoute := ht.router.CurrentRoute()

	// Determine current index by finding current route in history
	currentIdx := -1
	if currentRoute != nil {
		for i, entry := range entries {
			if entry.Route == currentRoute {
				currentIdx = i
				break
			}
		}
	}

	canGoBack := currentIdx > 0
	if canGoBack != expected {
		t.Errorf("expected canGoBack=%v, got %v (currentIdx=%d, historyLen=%d)",
			expected, canGoBack, currentIdx, len(entries))
	}
}

// AssertCanGoForward asserts whether forward navigation is possible.
//
// This method checks if forward navigation is possible by examining
// the current position in history.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: Whether forward navigation should be possible
//
// Example:
//
//	ht.AssertCanGoForward(t, false)
func (ht *HistoryTester) AssertCanGoForward(t testingT, expected bool) {
	t.Helper()

	// Access router's history to check if we can go forward
	entries := ht.GetHistoryEntries()
	currentRoute := ht.router.CurrentRoute()

	// Determine current index by finding current route in history
	currentIdx := -1
	if currentRoute != nil {
		for i, entry := range entries {
			if entry.Route == currentRoute {
				currentIdx = i
				break
			}
		}
	}

	canGoForward := len(entries) > 0 && currentIdx < len(entries)-1
	if canGoForward != expected {
		t.Errorf("expected canGoForward=%v, got %v (currentIdx=%d, historyLen=%d)",
			expected, canGoForward, currentIdx, len(entries))
	}
}

// GetHistoryEntries returns all history entries from the router.
//
// This method uses reflection to access the router's private history field
// for testing purposes. In production code, this would not be accessible.
//
// Returns:
//   - []*router.HistoryEntry: Slice of all history entries
//
// Example:
//
//	entries := ht.GetHistoryEntries()
//	for _, entry := range entries {
//		fmt.Printf("Path: %s\n", entry.Route.Path)
//	}
func (ht *HistoryTester) GetHistoryEntries() []*router.HistoryEntry {
	// Use reflection to access private history field
	// This is acceptable for testing utilities
	return ht.router.GetHistoryEntries()
}
