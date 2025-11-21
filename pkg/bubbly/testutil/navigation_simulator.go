package testutil

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// NavigationSimulator simulates router navigation and history for testing.
//
// It provides a simplified interface for testing navigation flows, tracking
// history as a slice of path strings, and managing back/forward navigation.
// This is useful for testing router behavior without needing a full Bubbletea
// application.
//
// Type Safety:
//   - Thread-safe navigation operations
//   - Simplified history tracking (paths only)
//   - Clear assertion methods for test verification
//
// Example:
//
//	func TestNavigation(t *testing.T) {
//		r := router.NewRouter()
//		r.Register("/home", "home", nil)
//		r.Register("/about", "about", nil)
//
//		ns := testutil.NewNavigationSimulator(r)
//
//		// Navigate forward
//		ns.Navigate("/home")
//		ns.Navigate("/about")
//		ns.AssertCurrentPath(t, "/about")
//		ns.AssertHistoryLength(t, 2)
//
//		// Navigate back
//		ns.Back()
//		ns.AssertCurrentPath(t, "/home")
//		ns.AssertCanGoForward(t, true)
//	}
type NavigationSimulator struct {
	// router is the router instance being tested
	router *router.Router

	// history tracks navigation paths in order
	history []string

	// currentIdx is the current position in history (-1 if empty)
	currentIdx int
}

// NewNavigationSimulator creates a new NavigationSimulator for testing navigation.
//
// Parameters:
//   - router: The router instance to test
//
// Returns:
//   - *NavigationSimulator: A new simulator instance
//
// Example:
//
//	r := router.NewRouter()
//	ns := testutil.NewNavigationSimulator(r)
func NewNavigationSimulator(r *router.Router) *NavigationSimulator {
	return &NavigationSimulator{
		router:     r,
		history:    make([]string, 0),
		currentIdx: -1,
	}
}

// Navigate simulates navigation to the specified path.
//
// This method:
//  1. Calls router.Push() to perform actual navigation
//  2. Executes the returned command to complete navigation
//  3. Tracks the path in history
//  4. Truncates forward history if navigating after going back
//
// Parameters:
//   - path: The path to navigate to
//
// Example:
//
//	ns.Navigate("/home")
//	ns.Navigate("/about")
//	// History: ["/home", "/about"], currentIdx: 1
func (ns *NavigationSimulator) Navigate(path string) {
	// Create navigation target
	target := &router.NavigationTarget{
		Path: path,
	}

	// Perform navigation using router.Push
	cmd := ns.router.Push(target)

	// Execute the command to complete navigation
	if cmd != nil {
		_ = cmd()
	}

	// Truncate forward history if we're not at the end
	// This mimics browser behavior: navigating after going back
	// removes all forward history entries
	if ns.currentIdx < len(ns.history)-1 {
		ns.history = ns.history[:ns.currentIdx+1]
	}

	// Add to history
	ns.history = append(ns.history, path)
	ns.currentIdx = len(ns.history) - 1
}

// Back simulates back navigation.
//
// This method:
//  1. Checks if back navigation is possible
//  2. Decrements currentIdx
//  3. Calls router.Back() to perform actual navigation
//  4. Executes the returned command
//
// If already at the start of history, this is a no-op.
//
// Example:
//
//	ns.Navigate("/home")
//	ns.Navigate("/about")
//	ns.Back()
//	// currentIdx: 0, current path: "/home"
func (ns *NavigationSimulator) Back() {
	// Can't go back if at start or empty
	if ns.currentIdx <= 0 {
		return
	}

	// Decrement index
	ns.currentIdx--

	// Perform back navigation using router.Back
	cmd := ns.router.Back()

	// Execute the command to complete navigation
	if cmd != nil {
		_ = cmd()
	}
}

// Forward simulates forward navigation.
//
// This method:
//  1. Checks if forward navigation is possible
//  2. Increments currentIdx
//  3. Calls router.Forward() to perform actual navigation
//  4. Executes the returned command
//
// If already at the end of history, this is a no-op.
//
// Example:
//
//	ns.Navigate("/home")
//	ns.Navigate("/about")
//	ns.Back()
//	ns.Forward()
//	// currentIdx: 1, current path: "/about"
func (ns *NavigationSimulator) Forward() {
	// Can't go forward if at end or empty
	if ns.currentIdx >= len(ns.history)-1 {
		return
	}

	// Increment index
	ns.currentIdx++

	// Perform forward navigation using router.Forward
	cmd := ns.router.Forward()

	// Execute the command to complete navigation
	if cmd != nil {
		_ = cmd()
	}
}

// AssertCurrentPath asserts that the current route path matches expected.
//
// This method uses the testingT interface for compatibility with both real
// and mock testing.T instances.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: The expected current path
//
// Example:
//
//	ns.Navigate("/home")
//	ns.AssertCurrentPath(t, "/home")
func (ns *NavigationSimulator) AssertCurrentPath(t testingT, expected string) {
	t.Helper()
	current := ns.router.CurrentRoute()
	if current == nil {
		t.Errorf("expected current path %q, but current route is nil", expected)
		return
	}
	if current.Path != expected {
		t.Errorf("expected current path %q, got %q", expected, current.Path)
	}
}

// AssertHistoryLength asserts that the history has the expected length.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: The expected history length
//
// Example:
//
//	ns.Navigate("/home")
//	ns.Navigate("/about")
//	ns.AssertHistoryLength(t, 2)
func (ns *NavigationSimulator) AssertHistoryLength(t testingT, expected int) {
	t.Helper()
	if len(ns.history) != expected {
		t.Errorf("expected history length %d, got %d", expected, len(ns.history))
	}
}

// AssertCanGoBack asserts whether back navigation is possible.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: Whether back navigation should be possible
//
// Example:
//
//	ns.Navigate("/home")
//	ns.AssertCanGoBack(t, false) // At start
//	ns.Navigate("/about")
//	ns.AssertCanGoBack(t, true)  // Can go back
func (ns *NavigationSimulator) AssertCanGoBack(t testingT, expected bool) {
	t.Helper()
	canGoBack := ns.currentIdx > 0
	if canGoBack != expected {
		t.Errorf("expected canGoBack=%v, got %v (currentIdx=%d, historyLen=%d)",
			expected, canGoBack, ns.currentIdx, len(ns.history))
	}
}

// AssertCanGoForward asserts whether forward navigation is possible.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: Whether forward navigation should be possible
//
// Example:
//
//	ns.Navigate("/home")
//	ns.Navigate("/about")
//	ns.AssertCanGoForward(t, false) // At end
//	ns.Back()
//	ns.AssertCanGoForward(t, true)  // Can go forward
func (ns *NavigationSimulator) AssertCanGoForward(t testingT, expected bool) {
	t.Helper()
	canGoForward := ns.currentIdx < len(ns.history)-1
	if canGoForward != expected {
		t.Errorf("expected canGoForward=%v, got %v (currentIdx=%d, historyLen=%d)",
			expected, canGoForward, ns.currentIdx, len(ns.history))
	}
}
