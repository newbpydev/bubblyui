package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// TestNewNestedRoutesTester tests the constructor
func TestNewNestedRoutesTester(t *testing.T) {
	t.Parallel()

	r := router.NewRouter()
	nrt := NewNestedRoutesTester(r)

	assert.NotNil(t, nrt)
	assert.NotNil(t, nrt.router)
	assert.Nil(t, nrt.parentRoute)
	assert.Empty(t, nrt.childRoutes)
	assert.Empty(t, nrt.activeRoutes)
}

// TestAssertActiveRoutes_NoRoute tests when no route is active
func TestAssertActiveRoutes_NoRoute(t *testing.T) {
	t.Parallel()

	r := router.NewRouter()
	nrt := NewNestedRoutesTester(r)

	// Should pass with empty expected
	mockT := &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{})
	assert.False(t, mockT.failed, "should not fail with empty expected when no route active")

	// Should fail with non-empty expected
	mockT = &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/home"})
	assert.True(t, mockT.failed, "should fail when expecting routes but none active")
	assert.NotEmpty(t, mockT.errors)
	assert.Contains(t, mockT.errors[0], "no route is active")
}

// TestAssertActiveRoutes_SingleRoute tests with a single active route (no nesting)
func TestAssertActiveRoutes_SingleRoute(t *testing.T) {
	t.Parallel()

	// Create router with a single route
	rb := router.NewRouterBuilder()
	rb.Route("/home", "home")
	r, err := rb.Build()
	assert.NoError(t, err)

	// Navigate to the route
	target := &router.NavigationTarget{Path: "/home"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	nrt := NewNestedRoutesTester(r)

	// Should pass with correct path
	mockT := &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/home"})
	assert.False(t, mockT.failed, "should pass with correct single route")

	// Should fail with wrong path
	mockT = &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/about"})
	assert.True(t, mockT.failed, "should fail with wrong route path")

	// Should fail with wrong count
	mockT = &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/home", "/home/stats"})
	assert.True(t, mockT.failed, "should fail with wrong route count")
}

// TestAssertActiveRoutes_NestedRoutes tests with nested routes
func TestAssertActiveRoutes_NestedRoutes(t *testing.T) {
	t.Parallel()

	// Create nested route structure using Child() API
	childStats := router.Child("/stats", router.WithName("dashboard-stats"))

	// Create router with nested routes using Builder
	rb := router.NewRouterBuilder()
	rb.RouteWithOptions("/dashboard",
		router.WithName("dashboard"),
		router.WithChildren(childStats),
	)
	r, err := rb.Build()
	assert.NoError(t, err)

	// Navigate to nested route
	target := &router.NavigationTarget{Path: "/dashboard/stats"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	nrt := NewNestedRoutesTester(r)

	// Should pass with correct hierarchy (using relative paths as stored in RouteRecord)
	mockT := &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/dashboard", "/stats"})
	assert.False(t, mockT.failed, "should pass with correct nested routes")

	// Should fail with wrong order
	mockT = &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/stats", "/dashboard"})
	assert.True(t, mockT.failed, "should fail with wrong route order")

	// Should fail with missing parent
	mockT = &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/stats"})
	assert.True(t, mockT.failed, "should fail with missing parent route")
}

// TestAssertActiveRoutes_DeepNesting tests with 3+ levels of nesting
func TestAssertActiveRoutes_DeepNesting(t *testing.T) {
	t.Parallel()

	// Create deeply nested route structure (3 levels)
	grandchildEdit := router.Child("/edit", router.WithName("admin-users-edit"))
	childUsers := router.Child("/users",
		router.WithName("admin-users"),
		router.WithChildren(grandchildEdit),
	)

	// Create router with deep nesting
	rb := router.NewRouterBuilder()
	rb.RouteWithOptions("/admin",
		router.WithName("admin"),
		router.WithChildren(childUsers),
	)
	r, err := rb.Build()
	assert.NoError(t, err)

	// Navigate to deeply nested route
	target := &router.NavigationTarget{Path: "/admin/users/edit"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	nrt := NewNestedRoutesTester(r)

	// Should pass with full hierarchy (using relative paths)
	mockT := &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/admin", "/users", "/edit"})
	assert.False(t, mockT.failed, "should pass with correct deep nesting")

	// Should fail with partial hierarchy
	mockT = &mockTestingT{}
	nrt.AssertActiveRoutes(mockT, []string{"/admin", "/edit"})
	assert.True(t, mockT.failed, "should fail with missing intermediate route")
}

// TestAssertParentActive_NoRoute tests when no route is active
func TestAssertParentActive_NoRoute(t *testing.T) {
	t.Parallel()

	r := router.NewRouter()
	nrt := NewNestedRoutesTester(r)

	mockT := &mockTestingT{}
	nrt.AssertParentActive(mockT)
	assert.True(t, mockT.failed, "should fail when no route is active")
	assert.NotEmpty(t, mockT.errors)
	assert.Contains(t, mockT.errors[0], "no route is active")
}

// TestAssertParentActive_SingleRoute tests with single route (no parent)
func TestAssertParentActive_SingleRoute(t *testing.T) {
	t.Parallel()

	// Create router with single route
	rb := router.NewRouterBuilder()
	rb.Route("/home", "home")
	r, err := rb.Build()
	assert.NoError(t, err)

	// Navigate to route
	target := &router.NavigationTarget{Path: "/home"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	nrt := NewNestedRoutesTester(r)

	mockT := &mockTestingT{}
	nrt.AssertParentActive(mockT)
	assert.True(t, mockT.failed, "should fail when route has no parent")
	assert.NotEmpty(t, mockT.errors)
	assert.Contains(t, mockT.errors[0], "has no parent")
}

// TestAssertParentActive_NestedRoute tests with nested route (has parent)
func TestAssertParentActive_NestedRoute(t *testing.T) {
	t.Parallel()

	// Create nested routes
	childStats := router.Child("/stats", router.WithName("dashboard-stats"))

	rb := router.NewRouterBuilder()
	rb.RouteWithOptions("/dashboard",
		router.WithName("dashboard"),
		router.WithChildren(childStats),
	)
	r, err := rb.Build()
	assert.NoError(t, err)

	// Navigate to nested route
	target := &router.NavigationTarget{Path: "/dashboard/stats"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	nrt := NewNestedRoutesTester(r)

	mockT := &mockTestingT{}
	nrt.AssertParentActive(mockT)
	assert.False(t, mockT.failed, "should pass when parent route is active")
}

// TestAssertChildActive_NoRoute tests when no route is active
func TestAssertChildActive_NoRoute(t *testing.T) {
	t.Parallel()

	r := router.NewRouter()
	nrt := NewNestedRoutesTester(r)

	mockT := &mockTestingT{}
	nrt.AssertChildActive(mockT, "/dashboard/stats")
	assert.True(t, mockT.failed, "should fail when no route is active")
	assert.NotEmpty(t, mockT.errors)
	assert.Contains(t, mockT.errors[0], "no route is active")
}

// TestAssertChildActive_SingleRoute tests with single route (no parent)
func TestAssertChildActive_SingleRoute(t *testing.T) {
	t.Parallel()

	// Create router with single route
	rb := router.NewRouterBuilder()
	rb.Route("/home", "home")
	r, err := rb.Build()
	assert.NoError(t, err)

	// Navigate to route
	target := &router.NavigationTarget{Path: "/home"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	nrt := NewNestedRoutesTester(r)

	mockT := &mockTestingT{}
	nrt.AssertChildActive(mockT, "/home")
	assert.True(t, mockT.failed, "should fail when route has no parent")
	assert.NotEmpty(t, mockT.errors)
	assert.Contains(t, mockT.errors[0], "has no parent")
}

// TestAssertChildActive_CorrectChild tests with correct child route
func TestAssertChildActive_CorrectChild(t *testing.T) {
	t.Parallel()

	// Create nested routes
	childStats := router.Child("/stats", router.WithName("dashboard-stats"))

	rb := router.NewRouterBuilder()
	rb.RouteWithOptions("/dashboard",
		router.WithName("dashboard"),
		router.WithChildren(childStats),
	)
	r, err := rb.Build()
	assert.NoError(t, err)

	// Navigate to nested route
	target := &router.NavigationTarget{Path: "/dashboard/stats"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	nrt := NewNestedRoutesTester(r)

	mockT := &mockTestingT{}
	nrt.AssertChildActive(mockT, "/stats")
	assert.False(t, mockT.failed, "should pass with correct child route")
}

// TestAssertChildActive_WrongChild tests with wrong child route
func TestAssertChildActive_WrongChild(t *testing.T) {
	t.Parallel()

	// Create nested routes with multiple children
	childStats := router.Child("/stats", router.WithName("dashboard-stats"))
	childSettings := router.Child("/settings", router.WithName("dashboard-settings"))

	rb := router.NewRouterBuilder()
	rb.RouteWithOptions("/dashboard",
		router.WithName("dashboard"),
		router.WithChildren(childStats, childSettings),
	)
	r, err := rb.Build()
	assert.NoError(t, err)

	// Navigate to stats route
	target := &router.NavigationTarget{Path: "/dashboard/stats"}
	cmd := r.Push(target)
	if cmd != nil {
		_ = cmd()
	}

	nrt := NewNestedRoutesTester(r)

	mockT := &mockTestingT{}
	nrt.AssertChildActive(mockT, "/settings")
	assert.True(t, mockT.failed, "should fail with wrong child route")
	assert.NotEmpty(t, mockT.errors)
	assert.Contains(t, mockT.errors[0], "expected child route")
}

// TestAssertActiveRoutes_TableDriven uses table-driven tests for multiple scenarios
func TestAssertActiveRoutes_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		routes        []string // Routes to register
		navigateTo    string   // Path to navigate to
		expected      []string // Expected active routes
		shouldSucceed bool     // Whether assertion should pass
	}{
		{
			name:          "single route",
			routes:        []string{"/home"},
			navigateTo:    "/home",
			expected:      []string{"/home"},
			shouldSucceed: true,
		},
		{
			name:          "wrong path",
			routes:        []string{"/home"},
			navigateTo:    "/home",
			expected:      []string{"/about"},
			shouldSucceed: false,
		},
		// NOTE: Nested route tests skipped until router supports Matched field population
		// {
		//     name:          "two-level nesting",
		//     routes:        []string{"/dashboard", "/dashboard/stats"},
		//     navigateTo:    "/dashboard/stats",
		//     expected:      []string{"/dashboard", "/dashboard/stats"},
		//     shouldSucceed: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create router and register routes
			rb := router.NewRouterBuilder()
			for _, route := range tt.routes {
				rb.Route(route, "route-"+route)
			}
			r, err := rb.Build()
			assert.NoError(t, err)

			// Navigate to target
			target := &router.NavigationTarget{Path: tt.navigateTo}
			cmd := r.Push(target)
			if cmd != nil {
				_ = cmd()
			}

			nrt := NewNestedRoutesTester(r)

			// Run assertion
			mockT := &mockTestingT{}
			nrt.AssertActiveRoutes(mockT, tt.expected)

			if tt.shouldSucceed {
				assert.False(t, mockT.failed, "assertion should succeed")
			} else {
				assert.True(t, mockT.failed, "assertion should fail")
			}
		})
	}
}
