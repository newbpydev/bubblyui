package testutil

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// MockRouter is a mock implementation of the router for testing.
// It tracks all navigation calls (Push, Replace, Back) and allows
// setting the current route for testing purposes.
//
// MockRouter provides:
//   - Navigation tracking: Records all Push, Replace, and Back calls
//   - Current route control: Set current route for testing
//   - Assertion helpers: Verify navigation behavior
//   - Thread-safe operations: Safe for concurrent use
//
// Call tracking:
//   - pushCalls: Slice of NavigationTarget for each Push() call
//   - replaceCalls: Slice of NavigationTarget for each Replace() call
//   - backCalls: Counter for Back() calls
//   - currentRoute: The current active route (settable for testing)
//
// Example:
//
//	mockRouter := NewMockRouter()
//	mockRouter.SetCurrentRoute(homeRoute)
//
//	// Trigger navigation
//	cmd := mockRouter.Push(&router.NavigationTarget{Path: "/about"})
//
//	// Assert navigation
//	mockRouter.AssertPushed(t, "/about")
//	assert.Equal(t, 1, mockRouter.GetPushCallCount())
type MockRouter struct {
	mu sync.RWMutex

	// Current state
	currentRoute *router.Route

	// Call tracking
	pushCalls    []*router.NavigationTarget
	replaceCalls []*router.NavigationTarget
	backCalls    int
}

// NewMockRouter creates a new mock router with empty state.
// All call counters are initialized to zero.
//
// Example:
//
//	mockRouter := NewMockRouter()
//	fmt.Println(mockRouter.GetPushCallCount())  // 0
//	fmt.Println(mockRouter.CurrentRoute())      // nil
func NewMockRouter() *MockRouter {
	return &MockRouter{
		pushCalls:    make([]*router.NavigationTarget, 0),
		replaceCalls: make([]*router.NavigationTarget, 0),
		backCalls:    0,
	}
}

// SetCurrentRoute sets the current route for testing.
// This allows tests to simulate being at a specific route.
//
// Example:
//
//	route := router.NewRoute("/home", "home", nil, nil, "", nil, nil)
//	mockRouter.SetCurrentRoute(route)
//	assert.Equal(t, "/home", mockRouter.CurrentRoute().Path)
func (mr *MockRouter) SetCurrentRoute(route *router.Route) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.currentRoute = route
}

// CurrentRoute returns the current active route.
// This implements the Router interface.
//
// Returns:
//   - *router.Route: The current route, or nil if no route is set
//
// Example:
//
//	route := mockRouter.CurrentRoute()
//	if route != nil {
//	    fmt.Printf("Current path: %s\n", route.Path)
//	}
func (mr *MockRouter) CurrentRoute() *router.Route {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.currentRoute
}

// Push records a Push navigation call and returns a no-op command.
// The navigation target is stored for later assertion.
//
// This implements the Router interface Push method.
//
// Parameters:
//   - target: The navigation target to push
//
// Returns:
//   - tea.Cmd: A no-op command (returns nil message)
//
// Example:
//
//	cmd := mockRouter.Push(&router.NavigationTarget{Path: "/about"})
//	mockRouter.AssertPushed(t, "/about")
func (mr *MockRouter) Push(target *router.NavigationTarget) tea.Cmd {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.pushCalls = append(mr.pushCalls, target)
	return func() tea.Msg { return nil }
}

// Replace records a Replace navigation call and returns a no-op command.
// The navigation target is stored for later assertion.
//
// This implements the Router interface Replace method.
//
// Parameters:
//   - target: The navigation target to replace with
//
// Returns:
//   - tea.Cmd: A no-op command (returns nil message)
//
// Example:
//
//	cmd := mockRouter.Replace(&router.NavigationTarget{Path: "/login"})
//	mockRouter.AssertReplaced(t, "/login")
func (mr *MockRouter) Replace(target *router.NavigationTarget) tea.Cmd {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.replaceCalls = append(mr.replaceCalls, target)
	return func() tea.Msg { return nil }
}

// Back records a Back navigation call and returns a no-op command.
// The call counter is incremented for later assertion.
//
// This implements the Router interface Back method.
//
// Returns:
//   - tea.Cmd: A no-op command (returns nil message)
//
// Example:
//
//	cmd := mockRouter.Back()
//	mockRouter.AssertBackCalled(t)
func (mr *MockRouter) Back() tea.Cmd {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.backCalls++
	return func() tea.Msg { return nil }
}

// GetPushCallCount returns the number of times Push was called.
//
// Example:
//
//	mockRouter.Push(&router.NavigationTarget{Path: "/about"})
//	mockRouter.Push(&router.NavigationTarget{Path: "/contact"})
//	assert.Equal(t, 2, mockRouter.GetPushCallCount())
func (mr *MockRouter) GetPushCallCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return len(mr.pushCalls)
}

// GetReplaceCallCount returns the number of times Replace was called.
//
// Example:
//
//	mockRouter.Replace(&router.NavigationTarget{Path: "/login"})
//	assert.Equal(t, 1, mockRouter.GetReplaceCallCount())
func (mr *MockRouter) GetReplaceCallCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return len(mr.replaceCalls)
}

// GetBackCallCount returns the number of times Back was called.
//
// Example:
//
//	mockRouter.Back()
//	mockRouter.Back()
//	assert.Equal(t, 2, mockRouter.GetBackCallCount())
func (mr *MockRouter) GetBackCallCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.backCalls
}

// GetPushCalls returns all Push navigation targets.
// Returns a copy to prevent external modification.
//
// Example:
//
//	calls := mockRouter.GetPushCalls()
//	for _, target := range calls {
//	    fmt.Printf("Pushed to: %s\n", target.Path)
//	}
func (mr *MockRouter) GetPushCalls() []*router.NavigationTarget {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	// Return a copy to prevent external modification
	calls := make([]*router.NavigationTarget, len(mr.pushCalls))
	copy(calls, mr.pushCalls)
	return calls
}

// GetReplaceCalls returns all Replace navigation targets.
// Returns a copy to prevent external modification.
//
// Example:
//
//	calls := mockRouter.GetReplaceCalls()
//	for _, target := range calls {
//	    fmt.Printf("Replaced with: %s\n", target.Path)
//	}
func (mr *MockRouter) GetReplaceCalls() []*router.NavigationTarget {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	// Return a copy to prevent external modification
	calls := make([]*router.NavigationTarget, len(mr.replaceCalls))
	copy(calls, mr.replaceCalls)
	return calls
}

// Reset clears all call tracking and resets the current route to nil.
// Useful for resetting state between test cases.
//
// Example:
//
//	mockRouter.Push(&router.NavigationTarget{Path: "/about"})
//	mockRouter.Reset()
//	assert.Equal(t, 0, mockRouter.GetPushCallCount())
//	assert.Nil(t, mockRouter.CurrentRoute())
func (mr *MockRouter) Reset() {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.pushCalls = make([]*router.NavigationTarget, 0)
	mr.replaceCalls = make([]*router.NavigationTarget, 0)
	mr.backCalls = 0
	mr.currentRoute = nil
}

// AssertPushed asserts that Push was called with the given path.
// Uses t.Helper() for proper stack traces in test output.
//
// Parameters:
//   - t: The testing interface
//   - path: The expected path that was pushed
//
// Example:
//
//	mockRouter.Push(&router.NavigationTarget{Path: "/about"})
//	mockRouter.AssertPushed(t, "/about")  // Passes
//	mockRouter.AssertPushed(t, "/contact")  // Fails with error
func (mr *MockRouter) AssertPushed(t testingT, path string) {
	t.Helper()
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, target := range mr.pushCalls {
		if target.Path == path {
			return // Found it
		}
	}

	// Not found - report error
	t.Errorf("Expected Push to be called with path %q, but it was not called with that path", path)
}

// AssertReplaced asserts that Replace was called with the given path.
// Uses t.Helper() for proper stack traces in test output.
//
// Parameters:
//   - t: The testing interface
//   - path: The expected path that was replaced
//
// Example:
//
//	mockRouter.Replace(&router.NavigationTarget{Path: "/login"})
//	mockRouter.AssertReplaced(t, "/login")  // Passes
func (mr *MockRouter) AssertReplaced(t testingT, path string) {
	t.Helper()
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	for _, target := range mr.replaceCalls {
		if target.Path == path {
			return // Found it
		}
	}

	// Not found - report error
	t.Errorf("Expected Replace to be called with path %q, but it was not called with that path", path)
}

// AssertBackCalled asserts that Back was called at least once.
// Uses t.Helper() for proper stack traces in test output.
//
// Example:
//
//	mockRouter.Back()
//	mockRouter.AssertBackCalled(t)  // Passes
func (mr *MockRouter) AssertBackCalled(t testingT) {
	t.Helper()
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if mr.backCalls == 0 {
		t.Errorf("Expected Back to be called at least once, but it was not called")
	}
}

// AssertBackNotCalled asserts that Back was never called.
// Uses t.Helper() for proper stack traces in test output.
//
// Example:
//
//	mockRouter.AssertBackNotCalled(t)  // Passes if Back never called
func (mr *MockRouter) AssertBackNotCalled(t testingT) {
	t.Helper()
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if mr.backCalls > 0 {
		t.Errorf("Expected Back to not be called, but it was called %d times", mr.backCalls)
	}
}

// AssertPushCount asserts that Push was called exactly the specified number of times.
// Uses t.Helper() for proper stack traces in test output.
//
// Parameters:
//   - t: The testing interface
//   - count: The expected number of Push calls
//
// Example:
//
//	mockRouter.Push(&router.NavigationTarget{Path: "/about"})
//	mockRouter.Push(&router.NavigationTarget{Path: "/contact"})
//	mockRouter.AssertPushCount(t, 2)  // Passes
func (mr *MockRouter) AssertPushCount(t testingT, count int) {
	t.Helper()
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if len(mr.pushCalls) != count {
		t.Errorf("Expected Push to be called %d times, but it was called %d times", count, len(mr.pushCalls))
	}
}

// AssertReplaceCount asserts that Replace was called exactly the specified number of times.
// Uses t.Helper() for proper stack traces in test output.
//
// Parameters:
//   - t: The testing interface
//   - count: The expected number of Replace calls
//
// Example:
//
//	mockRouter.Replace(&router.NavigationTarget{Path: "/login"})
//	mockRouter.AssertReplaceCount(t, 1)  // Passes
func (mr *MockRouter) AssertReplaceCount(t testingT, count int) {
	t.Helper()
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if len(mr.replaceCalls) != count {
		t.Errorf("Expected Replace to be called %d times, but it was called %d times", count, len(mr.replaceCalls))
	}
}

// AssertBackCount asserts that Back was called exactly the specified number of times.
// Uses t.Helper() for proper stack traces in test output.
//
// Parameters:
//   - t: The testing interface
//   - count: The expected number of Back calls
//
// Example:
//
//	mockRouter.Back()
//	mockRouter.Back()
//	mockRouter.AssertBackCount(t, 2)  // Passes
func (mr *MockRouter) AssertBackCount(t testingT, count int) {
	t.Helper()
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if mr.backCalls != count {
		t.Errorf("Expected Back to be called %d times, but it was called %d times", count, mr.backCalls)
	}
}
