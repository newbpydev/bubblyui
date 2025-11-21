package testutil

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// NestedRoutesTester provides utilities for testing nested route configuration and rendering.
//
// It wraps a router instance and provides assertion methods for verifying nested route
// hierarchies, active route chains, and parent-child relationships. This is useful for
// testing complex route structures with multiple levels of nesting.
//
// Type Safety:
//   - Thread-safe access to router state
//   - Clear assertion methods for nested route verification
//   - Direct access to route hierarchy for inspection
//
// Example:
//
//	func TestNestedRoutes(t *testing.T) {
//		r, err := router.NewRouterBuilder().
//			Route("/dashboard", "dashboard").
//			Route("/dashboard/stats", "dashboard-stats").
//			Route("/dashboard/settings", "dashboard-settings").
//			Build()
//		assert.NoError(t, err)
//
//		nrt := testutil.NewNestedRoutesTester(r)
//
//		// Navigate to nested route
//		target := &router.NavigationTarget{Path: "/dashboard/stats"}
//		cmd := r.Push(target)
//		if cmd != nil {
//			_ = cmd()
//		}
//
//		// Verify nested route hierarchy
//		nrt.AssertActiveRoutes(t, []string{"/dashboard", "/dashboard/stats"})
//		nrt.AssertParentActive(t)
//		nrt.AssertChildActive(t, "/dashboard/stats")
//	}
type NestedRoutesTester struct {
	// router is the router instance being tested
	router *router.Router

	// parentRoute is the parent route in the hierarchy (nil if none)
	parentRoute *router.Route

	// childRoutes is the list of child routes under the parent
	childRoutes []*router.Route

	// activeRoutes tracks the currently active route chain (parent to child)
	activeRoutes []string
}

// NewNestedRoutesTester creates a new NestedRoutesTester for testing nested routes.
//
// Parameters:
//   - router: The router instance to test
//
// Returns:
//   - *NestedRoutesTester: A new tester instance
//
// Example:
//
//	r := router.NewRouter()
//	nrt := testutil.NewNestedRoutesTester(r)
func NewNestedRoutesTester(r *router.Router) *NestedRoutesTester {
	return &NestedRoutesTester{
		router:       r,
		parentRoute:  nil,
		childRoutes:  make([]*router.Route, 0),
		activeRoutes: make([]string, 0),
	}
}

// AssertActiveRoutes asserts that the active route chain matches the expected paths.
//
// This method verifies the full route hierarchy from parent to child by examining
// the router's current route and its Matched field. The expected paths should be
// ordered from parent to child (e.g., ["/dashboard", "/dashboard/stats"]).
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: The expected route paths in order from parent to child
//
// Example:
//
//	// Single route (no nesting)
//	nrt.AssertActiveRoutes(t, []string{"/home"})
//
//	// Nested routes (parent and child)
//	nrt.AssertActiveRoutes(t, []string{"/dashboard", "/dashboard/stats"})
//
//	// Deep nesting (3 levels)
//	nrt.AssertActiveRoutes(t, []string{"/admin", "/admin/users", "/admin/users/edit"})
func (nrt *NestedRoutesTester) AssertActiveRoutes(t testingT, expected []string) {
	t.Helper()

	// Get current route from router
	currentRoute := nrt.router.CurrentRoute()
	if currentRoute == nil {
		if len(expected) > 0 {
			t.Errorf("expected active routes %v, but no route is active", expected)
		}
		return
	}

	// Extract active route paths from Matched field
	actual := make([]string, 0)
	for _, record := range currentRoute.Matched {
		actual = append(actual, record.Path)
	}

	// Compare lengths
	if len(actual) != len(expected) {
		t.Errorf("expected %d active routes %v, got %d routes %v",
			len(expected), expected, len(actual), actual)
		return
	}

	// Compare each path
	for i, expectedPath := range expected {
		if actual[i] != expectedPath {
			t.Errorf("active route at index %d: expected %q, got %q",
				i, expectedPath, actual[i])
		}
	}
}

// AssertParentActive asserts that a parent route is currently active.
//
// This method checks if the current route has a parent by examining the Matched
// field. A parent is considered active if the Matched chain has at least 2 entries
// (parent + child).
//
// Parameters:
//   - t: The testing.T instance (or mock)
//
// Example:
//
//	// After navigating to /dashboard/stats
//	nrt.AssertParentActive(t) // Should pass (parent is /dashboard)
//
//	// After navigating to /home (no parent)
//	nrt.AssertParentActive(t) // Should fail
func (nrt *NestedRoutesTester) AssertParentActive(t testingT) {
	t.Helper()

	// Get current route from router
	currentRoute := nrt.router.CurrentRoute()
	if currentRoute == nil {
		t.Errorf("expected parent route to be active, but no route is active")
		return
	}

	// Check if there's a parent (Matched should have at least 2 entries)
	if len(currentRoute.Matched) < 2 {
		t.Errorf("expected parent route to be active, but current route has no parent (matched: %d routes)",
			len(currentRoute.Matched))
		return
	}

	// Parent is active - success
}

// AssertChildActive asserts that a specific child route is currently active.
//
// This method verifies that the current route matches the expected child path
// and that it has a parent (is part of a nested hierarchy).
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - childPath: The expected child route path
//
// Example:
//
//	// After navigating to /dashboard/stats
//	nrt.AssertChildActive(t, "/dashboard/stats") // Should pass
//
//	// Wrong child path
//	nrt.AssertChildActive(t, "/dashboard/settings") // Should fail
func (nrt *NestedRoutesTester) AssertChildActive(t testingT, childPath string) {
	t.Helper()

	// Get current route from router
	currentRoute := nrt.router.CurrentRoute()
	if currentRoute == nil {
		t.Errorf("expected child route %q to be active, but no route is active", childPath)
		return
	}

	// Check if there's a parent (Matched should have at least 2 entries)
	if len(currentRoute.Matched) < 2 {
		t.Errorf("expected child route %q to be active, but current route has no parent", childPath)
		return
	}

	// Get the last matched route (the child)
	lastMatched := currentRoute.Matched[len(currentRoute.Matched)-1]

	// Compare child path
	if lastMatched.Path != childPath {
		t.Errorf("expected child route %q to be active, got %q", childPath, lastMatched.Path)
	}
}
