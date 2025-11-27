package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// TestNewRouteGuardTester tests RouteGuardTester creation
func TestNewRouteGuardTester(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"creates tester with router"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := router.NewRouter()
			tester := NewRouteGuardTester(r)

			assert.NotNil(t, tester)
			assert.NotNil(t, tester.router)
			assert.Equal(t, 0, tester.guardCalls)
			assert.False(t, tester.blocked)
		})
	}
}

// TestRouteGuardTester_AttemptNavigation tests navigation attempts
func TestRouteGuardTester_AttemptNavigation(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		guardBehavior string // "allow", "block", "redirect"
		expectedCalls int
		expectedBlock bool
	}{
		{
			name:          "guard allows navigation",
			path:          "/home",
			guardBehavior: "allow",
			expectedCalls: 1,
			expectedBlock: false,
		},
		{
			name:          "guard blocks navigation",
			path:          "/protected",
			guardBehavior: "block",
			expectedCalls: 1,
			expectedBlock: true,
		},
		{
			name:          "guard redirects navigation",
			path:          "/admin",
			guardBehavior: "redirect",
			expectedCalls: 1,
			expectedBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create builder
			builder := router.NewRouterBuilder().
				Route(tt.path, "test")

			// Create tester (will be updated with built router)
			tester := &RouteGuardTester{}

			// Register guard based on behavior
			switch tt.guardBehavior {
			case "allow":
				builder.BeforeEach(func(to, from *router.Route, next router.NextFunc) {
					tester.guardCalls++
					next(nil) // Allow
				})
			case "block":
				builder.BeforeEach(func(to, from *router.Route, next router.NextFunc) {
					tester.guardCalls++
					tester.blocked = true
					next(&router.NavigationTarget{}) // Block
				})
			case "redirect":
				builder.BeforeEach(func(to, from *router.Route, next router.NextFunc) {
					tester.guardCalls++
					next(&router.NavigationTarget{Path: "/login"}) // Redirect
				})
			}

			// Build router
			r, err := builder.Build()
			assert.NoError(t, err)
			tester.router = r

			// Attempt navigation
			tester.AttemptNavigation(tt.path)

			// Verify guard was called
			assert.Equal(t, tt.expectedCalls, tester.guardCalls)
			assert.Equal(t, tt.expectedBlock, tester.blocked)
		})
	}
}

// TestRouteGuardTester_AssertGuardCalled tests guard call assertions
func TestRouteGuardTester_AssertGuardCalled(t *testing.T) {
	tests := []struct {
		name          string
		navigations   int
		expectedCalls int
		shouldPass    bool
	}{
		{
			name:          "guard called once",
			navigations:   1,
			expectedCalls: 1,
			shouldPass:    true,
		},
		{
			name:          "guard called multiple times",
			navigations:   3,
			expectedCalls: 3,
			shouldPass:    true,
		},
		{
			name:          "guard not called",
			navigations:   0,
			expectedCalls: 0,
			shouldPass:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create tester
			tester := &RouteGuardTester{}

			// Build router with guard
			r, err := router.NewRouterBuilder().
				Route("/test", "test").
				BeforeEach(func(to, from *router.Route, next router.NextFunc) {
					tester.guardCalls++
					next(nil)
				}).
				Build()
			assert.NoError(t, err)
			tester.router = r

			// Perform navigations
			for i := 0; i < tt.navigations; i++ {
				tester.AttemptNavigation("/test")
			}

			// Create mock testing.T
			mockT := &mockTestingT{}
			tester.AssertGuardCalled(mockT, tt.expectedCalls)

			if tt.shouldPass {
				assert.False(t, mockT.failed, "AssertGuardCalled should pass")
			} else {
				assert.True(t, mockT.failed, "AssertGuardCalled should fail")
			}
		})
	}
}

// TestRouteGuardTester_MultipleGuards tests multiple guards
func TestRouteGuardTester_MultipleGuards(t *testing.T) {
	guard1Calls := 0
	guard2Calls := 0
	tester := &RouteGuardTester{}

	// Build router with multiple guards
	r, err := router.NewRouterBuilder().
		Route("/test", "test").
		BeforeEach(func(to, from *router.Route, next router.NextFunc) {
			guard1Calls++
			next(nil)
		}).
		BeforeEach(func(to, from *router.Route, next router.NextFunc) {
			guard2Calls++
			next(nil)
		}).
		Build()
	assert.NoError(t, err)
	tester.router = r

	// Navigate
	tester.AttemptNavigation("/test")

	// Both guards should be called
	assert.Equal(t, 1, guard1Calls)
	assert.Equal(t, 1, guard2Calls)
}

// TestRouteGuardTester_GuardStopsChain tests that blocking guard stops chain
func TestRouteGuardTester_GuardStopsChain(t *testing.T) {
	guard1Calls := 0
	guard2Calls := 0
	tester := &RouteGuardTester{}

	// Build router with guards (first blocks)
	r, err := router.NewRouterBuilder().
		Route("/test", "test").
		BeforeEach(func(to, from *router.Route, next router.NextFunc) {
			guard1Calls++
			tester.blocked = true
			next(&router.NavigationTarget{}) // Block
		}).
		BeforeEach(func(to, from *router.Route, next router.NextFunc) {
			guard2Calls++
			next(nil)
		}).
		Build()
	assert.NoError(t, err)
	tester.router = r

	// Navigate
	tester.AttemptNavigation("/test")

	// Only first guard should be called
	assert.Equal(t, 1, guard1Calls)
	assert.Equal(t, 0, guard2Calls)
	assert.True(t, tester.blocked)
}

// TestRouteGuardTester_InvalidRoute tests navigation to non-existent route
func TestRouteGuardTester_InvalidRoute(t *testing.T) {
	guardCalls := 0
	tester := &RouteGuardTester{}

	// Build router with guard but no matching route
	r, err := router.NewRouterBuilder().
		Route("/test", "test").
		BeforeEach(func(to, from *router.Route, next router.NextFunc) {
			guardCalls++
			next(nil)
		}).
		Build()
	assert.NoError(t, err)
	tester.router = r

	// Attempt navigation to non-existent route
	tester.AttemptNavigation("/nonexistent")

	// Guard should not be called for invalid route
	assert.Equal(t, 0, guardCalls)
}

// TestRouteGuardTester_AssertGuardCalledFailure tests assertion failure
func TestRouteGuardTester_AssertGuardCalledFailure(t *testing.T) {
	tester := &RouteGuardTester{}

	// Build router with guard
	r, err := router.NewRouterBuilder().
		Route("/test", "test").
		BeforeEach(func(to, from *router.Route, next router.NextFunc) {
			tester.guardCalls++
			next(nil)
		}).
		Build()
	assert.NoError(t, err)
	tester.router = r

	// Navigate once
	tester.AttemptNavigation("/test")

	// Create mock testing.T
	mockT := &mockTestingT{}

	// Assert wrong count (should fail)
	tester.AssertGuardCalled(mockT, 2)

	// Verify assertion failed
	assert.True(t, mockT.failed, "AssertGuardCalled should fail when count doesn't match")
	assert.Contains(t, mockT.errors[0], "expected guard to be called 2 times")
}

// Note: mockTestingT is already defined in assertions_state_test.go
