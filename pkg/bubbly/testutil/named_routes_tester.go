package testutil

import (
	"fmt"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// NamedRoutesTester provides testing utilities for named route registration and navigation.
//
// This tester helps verify:
//   - Named route registration and lookup
//   - Navigation by route name with parameters
//   - URL generation from route names
//   - Route name uniqueness enforcement
//
// Usage:
//
//	r, _ := router.NewRouterBuilder().
//		Route("/user/:id", "user-detail").
//		Build()
//	tester := testutil.NewNamedRoutesTester(r)
//
//	// Test navigation by name
//	tester.NavigateByName("user-detail", map[string]string{"id": "123"})
//	tester.AssertRouteName(t, "user-detail")
//
//	// Test URL generation
//	url, err := tester.GetRouteURL("user-detail", map[string]string{"id": "456"})
//	assert.NoError(t, err)
//	assert.Equal(t, "/user/456", url)
type NamedRoutesTester struct {
	router *router.Router
}

// NewNamedRoutesTester creates a new named routes tester for the given router.
//
// Parameters:
//   - r: The router instance to test
//
// Returns:
//   - *NamedRoutesTester: A new tester instance
//
// Example:
//
//	r, _ := router.NewRouterBuilder().
//		Route("/home", "home").
//		Build()
//	tester := testutil.NewNamedRoutesTester(r)
func NewNamedRoutesTester(r *router.Router) *NamedRoutesTester {
	return &NamedRoutesTester{
		router: r,
	}
}

// NavigateByName navigates to a route by its name with optional parameters.
//
// This method uses the router's PushNamed method to navigate by route name
// instead of path. It's useful for testing programmatic navigation without
// hardcoded paths.
//
// Parameters:
//   - name: The route name to navigate to (e.g., "user-detail")
//   - params: Path parameters to inject (e.g., {"id": "123"})
//
// Example:
//
//	tester.NavigateByName("user-detail", map[string]string{"id": "123"})
//	assert.Equal(t, "/user/123", tester.router.CurrentRoute().FullPath)
func (nrt *NamedRoutesTester) NavigateByName(name string, params map[string]string) {
	// Use router's PushNamed to navigate by name
	cmd := nrt.router.PushNamed(name, params, nil)
	if cmd != nil {
		// Execute the command to perform navigation
		// The command returns a NavigationMsg which updates the router's state
		cmd()
	}
}

// AssertRouteName asserts that the current route has the expected name.
//
// This method checks the current route's name and fails the test if it
// doesn't match the expected value.
//
// Parameters:
//   - t: The testing interface (usually *testing.T)
//   - expected: The expected route name
//
// Example:
//
//	tester.NavigateByName("home", nil)
//	tester.AssertRouteName(t, "home") // Passes
//	tester.AssertRouteName(t, "about") // Fails
func (nrt *NamedRoutesTester) AssertRouteName(t testingT, expected string) {
	t.Helper()

	currentRoute := nrt.router.CurrentRoute()
	if currentRoute == nil {
		t.Errorf("Expected route name %q, but no route is active", expected)
		return
	}

	if currentRoute.Name != expected {
		t.Errorf("Expected route name %q, got %q", expected, currentRoute.Name)
	}
}

// AssertRouteExists asserts that a route with the given name is registered.
//
// This method checks if a route with the specified name exists in the router's
// registry, without navigating to it.
//
// Parameters:
//   - t: The testing interface (usually *testing.T)
//   - name: The route name to check
//
// Example:
//
//	tester.AssertRouteExists(t, "home") // Passes if route exists
//	tester.AssertRouteExists(t, "nonexistent") // Fails
func (nrt *NamedRoutesTester) AssertRouteExists(t testingT, name string) {
	t.Helper()

	// Try to build a path from the route name
	// If the route doesn't exist, BuildPath will return an error
	_, err := nrt.router.BuildPath(name, nil, nil)
	if err != nil {
		t.Errorf("Route %q does not exist: %v", name, err)
	}
}

// GetRouteURL generates a URL from a route name and parameters.
//
// This method uses the router's BuildPath to construct a full URL from
// a route name and parameters, without navigating to it.
//
// Parameters:
//   - name: The route name (e.g., "user-detail")
//   - params: Path parameters to inject (e.g., {"id": "123"})
//
// Returns:
//   - string: The generated URL (e.g., "/user/123")
//   - error: nil on success, error if route not found or params missing
//
// Example:
//
//	url, err := tester.GetRouteURL("user-detail", map[string]string{"id": "123"})
//	assert.NoError(t, err)
//	assert.Equal(t, "/user/123", url)
func (nrt *NamedRoutesTester) GetRouteURL(name string, params map[string]string) (string, error) {
	// Use router's BuildPath to generate URL
	url, err := nrt.router.BuildPath(name, params, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build URL for route %q: %w", name, err)
	}
	return url, nil
}
