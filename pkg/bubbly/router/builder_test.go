package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRouterBuilder verifies builder creation
func TestNewRouterBuilder(t *testing.T) {
	builder := NewRouterBuilder()

	require.NotNil(t, builder, "Builder should not be nil")
	assert.Empty(t, builder.routes, "Routes should be empty initially")
	assert.Empty(t, builder.beforeHooks, "Before hooks should be empty initially")
	assert.Empty(t, builder.afterHooks, "After hooks should be empty initially")
}

// TestRouterBuilder_Route verifies route registration
func TestRouterBuilder_Route(t *testing.T) {
	builder := NewRouterBuilder()

	// Add route
	builder.Route("/home", "home")

	assert.Len(t, builder.routes, 1, "Should have one route")
	assert.Equal(t, "/home", builder.routes[0].Path)
	assert.Equal(t, "home", builder.routes[0].Name)
}

// TestRouterBuilder_FluentAPI verifies method chaining
func TestRouterBuilder_FluentAPI(t *testing.T) {
	builder := NewRouterBuilder().
		Route("/home", "home").
		Route("/about", "about").
		Route("/contact", "contact")

	assert.Len(t, builder.routes, 3, "Should have three routes")
}

// TestRouterBuilder_BeforeEach verifies guard registration
func TestRouterBuilder_BeforeEach(t *testing.T) {
	builder := NewRouterBuilder()

	guard := func(to, from *Route, next NextFunc) {
		next(nil)
	}

	builder.BeforeEach(guard)

	assert.Len(t, builder.beforeHooks, 1, "Should have one before hook")
}

// TestRouterBuilder_AfterEach verifies after hook registration
func TestRouterBuilder_AfterEach(t *testing.T) {
	builder := NewRouterBuilder()

	hook := func(to, from *Route) {
		// Hook logic
	}

	builder.AfterEach(hook)

	assert.Len(t, builder.afterHooks, 1, "Should have one after hook")
}

// TestRouterBuilder_Build verifies router creation
func TestRouterBuilder_Build(t *testing.T) {
	builder := NewRouterBuilder().
		Route("/home", "home").
		Route("/about", "about")

	router, err := builder.Build()

	require.NoError(t, err, "Build should succeed")
	require.NotNil(t, router, "Router should not be nil")

	// Verify routes were registered
	route, found := router.registry.GetByName("home")
	require.True(t, found, "Route 'home' should be found")
	assert.Equal(t, "/home", route.Path)

	route, found = router.registry.GetByName("about")
	require.True(t, found, "Route 'about' should be found")
	assert.Equal(t, "/about", route.Path)
}

// TestRouterBuilder_BuildWithGuards verifies guards are registered
func TestRouterBuilder_BuildWithGuards(t *testing.T) {
	guardCalled := false
	hookCalled := false

	builder := NewRouterBuilder().
		Route("/test", "test").
		BeforeEach(func(to, from *Route, next NextFunc) {
			guardCalled = true
			next(nil)
		}).
		AfterEach(func(to, from *Route) {
			hookCalled = true
		})

	router, err := builder.Build()
	require.NoError(t, err)

	// Navigate to trigger guards
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	cmd()

	assert.True(t, guardCalled, "Guard should be called")
	assert.True(t, hookCalled, "Hook should be called")
}

// TestRouterBuilder_Validation verifies build validation
func TestRouterBuilder_Validation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*RouterBuilder)
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty path",
			setup: func(rb *RouterBuilder) {
				rb.Route("", "empty")
			},
			expectError: true,
			errorMsg:    "path cannot be empty",
		},
		{
			name: "duplicate paths",
			setup: func(rb *RouterBuilder) {
				rb.Route("/home", "home1")
				rb.Route("/home", "home2")
			},
			expectError: true,
			errorMsg:    "duplicate path",
		},
		{
			name: "duplicate names",
			setup: func(rb *RouterBuilder) {
				rb.Route("/home", "home")
				rb.Route("/about", "home")
			},
			expectError: true,
			errorMsg:    "duplicate name",
		},
		{
			name: "valid routes",
			setup: func(rb *RouterBuilder) {
				rb.Route("/home", "home")
				rb.Route("/about", "about")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewRouterBuilder()
			tt.setup(builder)

			router, err := builder.Build()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, router)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, router)
			}
		})
	}
}

// TestRouterBuilder_RouteWithMeta verifies meta data registration
func TestRouterBuilder_RouteWithMeta(t *testing.T) {
	meta := map[string]interface{}{
		"requiresAuth": true,
		"title":        "Home Page",
	}

	builder := NewRouterBuilder().
		RouteWithMeta("/home", "home", meta)

	router, err := builder.Build()
	require.NoError(t, err)

	route, found := router.registry.GetByName("home")
	require.True(t, found, "Route 'home' should be found")

	assert.Equal(t, true, route.Meta["requiresAuth"])
	assert.Equal(t, "Home Page", route.Meta["title"])
}

// TestRouterBuilder_ComplexScenario verifies complex builder usage
func TestRouterBuilder_ComplexScenario(t *testing.T) {
	authGuard := func(to, from *Route, next NextFunc) {
		if requiresAuth, ok := to.Meta["requiresAuth"].(bool); ok && requiresAuth {
			// In real app, check auth state
			next(nil) // Allow for test
		} else {
			next(nil)
		}
	}

	analyticsHook := func(to, from *Route) {
		// Track page view
		_ = to.Path
	}

	builder := NewRouterBuilder().
		Route("/", "home").
		RouteWithMeta("/dashboard", "dashboard", map[string]interface{}{
			"requiresAuth": true,
		}).
		Route("/login", "login").
		BeforeEach(authGuard).
		AfterEach(analyticsHook)

	router, err := builder.Build()
	require.NoError(t, err)

	// Verify all routes registered
	assert.Len(t, builder.routes, 3)

	// Verify guards registered
	assert.Len(t, router.beforeHooks, 1)
	assert.Len(t, router.afterHooks, 1)

	// Test navigation
	cmd := router.Push(&NavigationTarget{Path: "/dashboard"})
	msg := cmd()
	assert.IsType(t, RouteChangedMsg{}, msg)
}

// TestRouterBuilder_EmptyBuild verifies empty builder
func TestRouterBuilder_EmptyBuild(t *testing.T) {
	builder := NewRouterBuilder()

	router, err := builder.Build()

	// Empty router is valid (routes can be added later)
	assert.NoError(t, err)
	assert.NotNil(t, router)
}

// TestRouterBuilder_MultipleBuilds verifies builder can be reused
func TestRouterBuilder_MultipleBuilds(t *testing.T) {
	builder := NewRouterBuilder().
		Route("/home", "home")

	// First build
	router1, err := builder.Build()
	require.NoError(t, err)

	// Second build
	router2, err := builder.Build()
	require.NoError(t, err)

	// Should be different router instances
	assert.NotSame(t, router1, router2)

	// But both should have the same routes
	route1, _ := router1.registry.GetByName("home")
	route2, _ := router2.registry.GetByName("home")
	assert.Equal(t, route1.Path, route2.Path)
}
