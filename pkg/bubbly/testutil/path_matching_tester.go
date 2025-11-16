package testutil

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// PathMatchingTester provides testing utilities for route path pattern matching and parameters.
//
// This tester helps verify:
//   - Static paths match exactly
//   - Dynamic segments are captured correctly
//   - Wildcard patterns work as expected
//   - Optional segments are supported
//   - Priority/specificity ordering is correct
//
// Usage:
//
//	r, _ := router.NewRouterBuilder().
//		Route("/user/:id", "user-detail").
//		Route("/users/new", "user-new").
//		Build()
//	tester := testutil.NewPathMatchingTester(r)
//
//	// Test pattern matching
//	assert.True(t, tester.TestMatch("/user/:id", "/user/123"))
//	assert.False(t, tester.TestMatch("/user/:id", "/post/123"))
//
//	// Use assertion helpers
//	tester.AssertMatches(t, "/user/:id", "/user/123")
//	tester.AssertNotMatches(t, "/user/:id", "/post/123")
//
//	// Extract parameters
//	params := tester.ExtractParams("/user/:id", "/user/123")
//	assert.Equal(t, "123", params["id"])
//
// Thread Safety:
//
// PathMatchingTester is not thread-safe. Each test should create its own tester instance.
type PathMatchingTester struct {
	// router is the router instance being tested
	router *router.Router
}

// NewPathMatchingTester creates a new path matching tester for the given router.
//
// Parameters:
//   - r: The router instance to test
//
// Returns:
//   - *PathMatchingTester: A new tester instance
//
// Example:
//
//	r, _ := router.NewRouterBuilder().
//		Route("/user/:id", "user-detail").
//		Build()
//	tester := testutil.NewPathMatchingTester(r)
func NewPathMatchingTester(r *router.Router) *PathMatchingTester {
	return &PathMatchingTester{
		router: r,
	}
}

// TestMatch tests if a pattern matches a given path.
//
// This method checks if the specified pattern successfully matches the provided path
// by attempting to navigate to the path and checking if the matched route's pattern
// matches the expected pattern.
//
// Parameters:
//   - pattern: The route pattern to test (e.g., "/user/:id")
//   - path: The path to match against (e.g., "/user/123")
//
// Returns:
//   - bool: true if the pattern matches the path, false otherwise
//
// Example:
//
//	// Static path matching
//	matches := tester.TestMatch("/home", "/home")  // true
//	matches = tester.TestMatch("/home", "/about")  // false
//
//	// Dynamic parameter matching
//	matches = tester.TestMatch("/user/:id", "/user/123")  // true
//	matches = tester.TestMatch("/user/:id", "/post/123")  // false
//
//	// Wildcard matching
//	matches = tester.TestMatch("/docs/:path*", "/docs/guide/intro")  // true
func (pmt *PathMatchingTester) TestMatch(pattern, path string) bool {
	// Navigate to the path to trigger route matching
	cmd := pmt.router.Push(&router.NavigationTarget{Path: path})
	if cmd != nil {
		cmd() // Execute the command
	}

	// Get the current route
	currentRoute := pmt.router.CurrentRoute()
	if currentRoute == nil {
		return false
	}

	// Check if the matched route's path matches the expected pattern
	return currentRoute.FullPath == pattern
}

// AssertMatches asserts that a pattern matches a given path.
//
// This method fails the test if the pattern does not match the path.
// It provides a clear error message indicating which pattern and path were tested.
//
// Parameters:
//   - t: The testing interface (usually *testing.T)
//   - pattern: The route pattern that should match (e.g., "/user/:id")
//   - path: The path to test (e.g., "/user/123")
//
// Example:
//
//	tester.AssertMatches(t, "/user/:id", "/user/123")  // Passes
//	tester.AssertMatches(t, "/user/:id", "/post/123")  // Fails with error
func (pmt *PathMatchingTester) AssertMatches(t testingT, pattern, path string) {
	t.Helper()

	if !pmt.TestMatch(pattern, path) {
		t.Errorf("Expected pattern %q to match path %q, but it did not", pattern, path)
	}
}

// AssertNotMatches asserts that a pattern does not match a given path.
//
// This method fails the test if the pattern matches the path when it shouldn't.
// It provides a clear error message indicating which pattern and path were tested.
//
// Parameters:
//   - t: The testing interface (usually *testing.T)
//   - pattern: The route pattern that should not match (e.g., "/user/:id")
//   - path: The path to test (e.g., "/post/123")
//
// Example:
//
//	tester.AssertNotMatches(t, "/user/:id", "/post/123")  // Passes
//	tester.AssertNotMatches(t, "/user/:id", "/user/123")  // Fails with error
func (pmt *PathMatchingTester) AssertNotMatches(t testingT, pattern, path string) {
	t.Helper()

	if pmt.TestMatch(pattern, path) {
		t.Errorf("Expected pattern %q to not match path %q, but it did", pattern, path)
	}
}

// ExtractParams extracts path parameters from a pattern and path match.
//
// This method navigates to the path and returns the extracted parameters from
// the matched route. If the pattern doesn't match the path, an empty map is returned.
//
// Parameters:
//   - pattern: The route pattern (e.g., "/user/:id")
//   - path: The path to extract parameters from (e.g., "/user/123")
//
// Returns:
//   - map[string]string: Extracted parameters (e.g., {"id": "123"})
//
// Example:
//
//	// Single parameter
//	params := tester.ExtractParams("/user/:id", "/user/123")
//	// params = {"id": "123"}
//
//	// Multiple parameters
//	params = tester.ExtractParams("/posts/:category/:id", "/posts/tech/456")
//	// params = {"category": "tech", "id": "456"}
//
//	// Wildcard parameter
//	params = tester.ExtractParams("/docs/:path*", "/docs/guide/intro")
//	// params = {"path": "guide/intro"}
//
//	// Optional parameter present
//	params = tester.ExtractParams("/profile/:id?", "/profile/123")
//	// params = {"id": "123"}
//
//	// Optional parameter absent
//	params = tester.ExtractParams("/profile/:id?", "/profile")
//	// params = {}
func (pmt *PathMatchingTester) ExtractParams(pattern, path string) map[string]string {
	// Navigate to the path to trigger route matching
	cmd := pmt.router.Push(&router.NavigationTarget{Path: path})
	if cmd != nil {
		cmd() // Execute the command
	}

	// Get the current route
	currentRoute := pmt.router.CurrentRoute()
	if currentRoute == nil {
		return map[string]string{}
	}

	// Check if the matched route's path matches the expected pattern
	if currentRoute.FullPath != pattern {
		return map[string]string{}
	}

	// Return the extracted parameters
	if currentRoute.Params == nil {
		return map[string]string{}
	}

	return currentRoute.Params
}
