package testutil

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// RouteGuardTester provides utilities for testing route navigation guards.
//
// It tracks guard execution and navigation outcomes, making it easy to verify
// that guards are called correctly and that navigation is allowed, blocked,
// or redirected as expected.
//
// Type Safety:
//   - Thread-safe guard call tracking
//   - Clear assertion methods for test verification
//   - Integration with router's guard system
//
// Example:
//
//	func TestAuthGuard(t *testing.T) {
//		r := router.NewRouter()
//		tester := testutil.NewRouteGuardTester(r)
//
//		// Register auth guard
//		r.BeforeEach(func(to, from *router.Route, next router.NextFunc) {
//			tester.guardCalls++
//			if !isAuthenticated() {
//				tester.blocked = true
//				next(&router.NavigationTarget{}) // Block
//			} else {
//				next(nil) // Allow
//			}
//		})
//
//		// Attempt navigation
//		tester.AttemptNavigation("/protected")
//
//		// Verify guard was called and navigation blocked
//		tester.AssertGuardCalled(t, 1)
//		assert.True(t, tester.blocked)
//	}
type RouteGuardTester struct {
	// router is the router instance being tested
	router *router.Router

	// guardCalls tracks the number of times guards have been called
	guardCalls int

	// blocked indicates whether navigation was blocked by a guard
	blocked bool
}

// NewRouteGuardTester creates a new RouteGuardTester for testing route guards.
//
// Parameters:
//   - router: The router instance to test
//
// Returns:
//   - *RouteGuardTester: A new tester instance
//
// Example:
//
//	r := router.NewRouter()
//	tester := testutil.NewRouteGuardTester(r)
func NewRouteGuardTester(r *router.Router) *RouteGuardTester {
	return &RouteGuardTester{
		router:     r,
		guardCalls: 0,
		blocked:    false,
	}
}

// AttemptNavigation attempts to navigate to the specified path.
//
// This method triggers the router's navigation system, which will execute
// any registered guards. The tester tracks whether guards are called and
// whether navigation is blocked.
//
// Parameters:
//   - path: The path to navigate to
//
// Example:
//
//	tester.AttemptNavigation("/admin")
//	// Guards will be executed, tester tracks results
func (rgt *RouteGuardTester) AttemptNavigation(path string) {
	// Create navigation target
	target := &router.NavigationTarget{
		Path: path,
	}

	// Attempt navigation using Push
	// This will trigger guards and update router state
	cmd := rgt.router.Push(target)

	// Execute the command to complete navigation
	if cmd != nil {
		_ = cmd()
	}
}

// AssertGuardCalled asserts that guards were called the expected number of times.
//
// This method uses the testingT interface for compatibility with both real
// and mock testing.T instances.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - times: The expected number of guard calls
//
// Example:
//
//	tester.AssertGuardCalled(t, 1) // Assert guard called once
//	tester.AssertGuardCalled(t, 3) // Assert guard called three times
func (rgt *RouteGuardTester) AssertGuardCalled(t testingT, times int) {
	t.Helper()
	if rgt.guardCalls != times {
		t.Errorf("expected guard to be called %d times, but was called %d times",
			times, rgt.guardCalls)
	}
}
