package testutil

import (
	"reflect"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// QueryParamsTester provides utilities for testing query parameter handling in the router.
//
// This tester helps verify that query parameters are correctly parsed, updated,
// and preserved during navigation. It supports testing URL encoding/decoding,
// multiple parameters, and parameter removal.
//
// Features:
//   - Parse query params from URL
//   - Set individual query parameters
//   - Assert on query parameter values
//   - Clear all query parameters
//   - Test navigation with query params
//
// Usage:
//
//	// Create router and navigate to route with query params
//	router, _ := router.NewRouterBuilder().
//		Route("/search", "search").
//		Build()
//	router.Push(&router.NavigationTarget{
//		Path:  "/search",
//		Query: map[string]string{"q": "golang"},
//	})()
//
//	// Create tester
//	tester := testutil.NewQueryParamsTester(router)
//
//	// Assert query params
//	tester.AssertQueryParam(t, "q", "golang")
//	tester.AssertQueryParams(t, map[string]string{"q": "golang"})
//
//	// Update query params
//	tester.SetQueryParam("page", "2")
//	tester.AssertQueryParam(t, "page", "2")
//
//	// Clear all params
//	tester.ClearQueryParams()
//	tester.AssertQueryParams(t, map[string]string{})
//
// Thread Safety:
// QueryParamsTester is not thread-safe. Each test should create its own tester instance.
type QueryParamsTester struct {
	router *router.Router
}

// NewQueryParamsTester creates a new query params tester for the given router.
//
// The tester provides utilities for testing query parameter handling,
// including parsing, updating, and asserting on query parameters.
//
// Parameters:
//   - router: The router instance to test
//
// Returns:
//   - *QueryParamsTester: A new tester instance
//
// Example:
//
//	router, _ := router.NewRouterBuilder().
//		Route("/search", "search").
//		Build()
//	tester := testutil.NewQueryParamsTester(router)
func NewQueryParamsTester(r *router.Router) *QueryParamsTester {
	return &QueryParamsTester{
		router: r,
	}
}

// SetQueryParam sets a single query parameter by navigating with updated query params.
//
// This method updates the current route's query parameters by performing a new
// navigation with the updated query map. It preserves all existing parameters
// and adds or updates the specified parameter.
//
// Parameters:
//   - key: The query parameter key
//   - value: The query parameter value
//
// Behavior:
//   - Preserves existing query parameters
//   - Adds new parameter if key doesn't exist
//   - Updates existing parameter if key exists
//   - Uses router.Push() to trigger navigation
//   - Empty values are preserved (use ClearQueryParams to remove)
//
// Example:
//
//	// Current route: /search?q=golang
//	tester.SetQueryParam("page", "2")
//	// Result: /search?q=golang&page=2
//
//	tester.SetQueryParam("q", "rust")
//	// Result: /search?q=rust&page=2
func (qpt *QueryParamsTester) SetQueryParam(key, value string) {
	// Get current route
	currentRoute := qpt.router.CurrentRoute()
	if currentRoute == nil {
		return
	}

	// Copy existing query params
	newQuery := make(map[string]string)
	for k, v := range currentRoute.Query {
		newQuery[k] = v
	}

	// Set new param
	newQuery[key] = value

	// Navigate with updated query
	cmd := qpt.router.Push(&router.NavigationTarget{
		Path:  currentRoute.Path,
		Query: newQuery,
	})
	cmd()
}

// AssertQueryParam asserts that a query parameter has the expected value.
//
// This method checks if the specified query parameter exists in the current
// route and has the expected value. It fails the test if the parameter is
// missing or has a different value.
//
// Parameters:
//   - t: The testing.T instance for assertions
//   - key: The query parameter key to check
//   - expected: The expected value
//
// Behavior:
//   - Fails if parameter doesn't exist
//   - Fails if parameter value doesn't match expected
//   - Passes if parameter exists and matches expected
//
// Example:
//
//	// Current route: /search?q=golang&page=1
//	tester.AssertQueryParam(t, "q", "golang")      // Pass
//	tester.AssertQueryParam(t, "page", "1")        // Pass
//	tester.AssertQueryParam(t, "sort", "date")     // Fail - missing
//	tester.AssertQueryParam(t, "q", "rust")        // Fail - wrong value
func (qpt *QueryParamsTester) AssertQueryParam(t testingT, key, expected string) {
	currentRoute := qpt.router.CurrentRoute()
	if currentRoute == nil {
		t.Errorf("no current route")
		return
	}

	actual, ok := currentRoute.Query[key]
	if !ok {
		t.Errorf("query param %q not found in route", key)
		return
	}

	if actual != expected {
		t.Errorf("query param %q: expected %q, got %q", key, expected, actual)
	}
}

// AssertQueryParams asserts that all query parameters match the expected map.
//
// This method performs a deep equality check between the current route's
// query parameters and the expected map. It fails if there are any differences
// in keys or values.
//
// Parameters:
//   - t: The testing.T instance for assertions
//   - expected: Map of expected query parameters
//
// Behavior:
//   - Fails if any expected parameter is missing
//   - Fails if any parameter value doesn't match
//   - Fails if there are extra parameters not in expected
//   - Passes only if maps are exactly equal
//
// Example:
//
//	// Current route: /search?q=golang&page=1
//	tester.AssertQueryParams(t, map[string]string{
//		"q":    "golang",
//		"page": "1",
//	}) // Pass
//
//	tester.AssertQueryParams(t, map[string]string{
//		"q": "golang",
//	}) // Fail - missing "page" in expected but present in route
//
//	tester.AssertQueryParams(t, map[string]string{}) // Fail - route has params
func (qpt *QueryParamsTester) AssertQueryParams(t testingT, expected map[string]string) {
	currentRoute := qpt.router.CurrentRoute()
	if currentRoute == nil {
		t.Errorf("no current route")
		return
	}

	actual := currentRoute.Query

	// Use reflect.DeepEqual for map comparison
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("query params mismatch:\nexpected: %v\nactual:   %v", expected, actual)
	}
}

// ClearQueryParams clears all query parameters by navigating without query params.
//
// This method removes all query parameters from the current route by performing
// a new navigation to the same path without any query parameters.
//
// Behavior:
//   - Removes all query parameters
//   - Preserves current path
//   - Uses router.Push() to trigger navigation
//   - Results in route with empty Query map
//
// Example:
//
//	// Current route: /search?q=golang&page=1
//	tester.ClearQueryParams()
//	// Result: /search (no query params)
//
//	tester.AssertQueryParams(t, map[string]string{}) // Pass
func (qpt *QueryParamsTester) ClearQueryParams() {
	currentRoute := qpt.router.CurrentRoute()
	if currentRoute == nil {
		return
	}

	// Navigate with no query params
	cmd := qpt.router.Push(&router.NavigationTarget{
		Path:  currentRoute.Path,
		Query: map[string]string{},
	})
	cmd()
}
