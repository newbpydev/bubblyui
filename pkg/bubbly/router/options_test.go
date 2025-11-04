package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithName verifies WithName option
func TestWithName(t *testing.T) {
	record := &RouteRecord{Path: "/test"}

	option := WithName("test-route")
	option(record)

	assert.Equal(t, "test-route", record.Name)
}

// TestWithMeta verifies WithMeta option
func TestWithMeta(t *testing.T) {
	meta := map[string]interface{}{
		"requiresAuth": true,
		"title":        "Test Page",
	}

	record := &RouteRecord{Path: "/test"}

	option := WithMeta(meta)
	option(record)

	assert.Equal(t, true, record.Meta["requiresAuth"])
	assert.Equal(t, "Test Page", record.Meta["title"])
}

// TestWithGuard verifies WithGuard option
func TestWithGuard(t *testing.T) {
	guardCalled := false
	guard := func(to, from *Route, next NextFunc) {
		guardCalled = true
		next(nil)
	}

	record := &RouteRecord{Path: "/test"}

	option := WithGuard(guard)
	option(record)

	// Verify guard was stored
	require.NotNil(t, record.Meta)
	require.NotNil(t, record.Meta["beforeEnter"])

	// Call the guard to verify it works
	storedGuard := record.Meta["beforeEnter"].(NavigationGuard)
	storedGuard(nil, nil, func(target *NavigationTarget) {})

	assert.True(t, guardCalled)
}

// TestWithChildren verifies WithChildren option
func TestWithChildren(t *testing.T) {
	child1 := &RouteRecord{Path: "/child1", Name: "child1"}
	child2 := &RouteRecord{Path: "/child2", Name: "child2"}

	record := &RouteRecord{Path: "/parent"}

	option := WithChildren(child1, child2)
	option(record)

	require.Len(t, record.Children, 2)
	assert.Equal(t, "/child1", record.Children[0].Path)
	assert.Equal(t, "/child2", record.Children[1].Path)
}

// TestMultipleOptions verifies options can be combined
func TestMultipleOptions(t *testing.T) {
	meta := map[string]interface{}{
		"requiresAuth": true,
	}

	child := &RouteRecord{Path: "/child", Name: "child"}

	record := &RouteRecord{Path: "/test"}

	// Apply multiple options
	WithName("test-route")(record)
	WithMeta(meta)(record)
	WithChildren(child)(record)

	assert.Equal(t, "test-route", record.Name)
	assert.Equal(t, true, record.Meta["requiresAuth"])
	require.Len(t, record.Children, 1)
	assert.Equal(t, "/child", record.Children[0].Path)
}

// TestRouterBuilder_RouteWithOptions verifies builder accepts options
func TestRouterBuilder_RouteWithOptions(t *testing.T) {
	builder := NewRouterBuilder()

	builder.RouteWithOptions("/test",
		WithName("test-route"),
		WithMeta(map[string]interface{}{
			"requiresAuth": true,
		}),
	)

	require.Len(t, builder.routes, 1)
	assert.Equal(t, "/test", builder.routes[0].Path)
	assert.Equal(t, "test-route", builder.routes[0].Name)
	assert.Equal(t, true, builder.routes[0].Meta["requiresAuth"])
}

// TestRouterBuilder_RouteWithOptionsAndChildren verifies nested routes
func TestRouterBuilder_RouteWithOptionsAndChildren(t *testing.T) {
	child1 := &RouteRecord{Path: "/child1", Name: "child1"}
	child2 := &RouteRecord{Path: "/child2", Name: "child2"}

	builder := NewRouterBuilder()

	builder.RouteWithOptions("/parent",
		WithName("parent"),
		WithChildren(child1, child2),
	)

	require.Len(t, builder.routes, 1)
	assert.Equal(t, "/parent", builder.routes[0].Path)
	assert.Equal(t, "parent", builder.routes[0].Name)
	require.Len(t, builder.routes[0].Children, 2)
	assert.Equal(t, "/child1", builder.routes[0].Children[0].Path)
	assert.Equal(t, "/child2", builder.routes[0].Children[1].Path)
}

// TestRouteOptions_ComplexScenario verifies complex option combinations
func TestRouteOptions_ComplexScenario(t *testing.T) {
	authGuard := func(to, from *Route, next NextFunc) {
		next(nil)
	}

	dashboardChild := &RouteRecord{
		Path: "/overview",
		Name: "dashboard-overview",
	}

	settingsChild := &RouteRecord{
		Path: "/settings",
		Name: "dashboard-settings",
	}

	builder := NewRouterBuilder().
		RouteWithOptions("/",
			WithName("home"),
		).
		RouteWithOptions("/dashboard",
			WithName("dashboard"),
			WithMeta(map[string]interface{}{
				"requiresAuth": true,
				"title":        "Dashboard",
			}),
			WithGuard(authGuard),
			WithChildren(dashboardChild, settingsChild),
		).
		RouteWithOptions("/login",
			WithName("login"),
			WithMeta(map[string]interface{}{
				"title": "Login",
			}),
		)

	require.Len(t, builder.routes, 3)

	// Verify home route
	assert.Equal(t, "/", builder.routes[0].Path)
	assert.Equal(t, "home", builder.routes[0].Name)

	// Verify dashboard route
	dashboard := builder.routes[1]
	assert.Equal(t, "/dashboard", dashboard.Path)
	assert.Equal(t, "dashboard", dashboard.Name)
	assert.Equal(t, true, dashboard.Meta["requiresAuth"])
	assert.Equal(t, "Dashboard", dashboard.Meta["title"])
	assert.NotNil(t, dashboard.Meta["beforeEnter"])
	require.Len(t, dashboard.Children, 2)

	// Verify login route
	assert.Equal(t, "/login", builder.routes[2].Path)
	assert.Equal(t, "login", builder.routes[2].Name)
	assert.Equal(t, "Login", builder.routes[2].Meta["title"])
}

// TestWithMeta_MergesWithExisting verifies meta merging
func TestWithMeta_MergesWithExisting(t *testing.T) {
	record := &RouteRecord{
		Path: "/test",
		Meta: map[string]interface{}{
			"existing": "value",
		},
	}

	option := WithMeta(map[string]interface{}{
		"new": "data",
	})
	option(record)

	// Both old and new meta should be present
	assert.Equal(t, "value", record.Meta["existing"])
	assert.Equal(t, "data", record.Meta["new"])
}

// TestWithChildren_AppendsToExisting verifies children appending
func TestWithChildren_AppendsToExisting(t *testing.T) {
	existing := &RouteRecord{Path: "/existing", Name: "existing"}
	record := &RouteRecord{
		Path:     "/test",
		Children: []*RouteRecord{existing},
	}

	newChild := &RouteRecord{Path: "/new", Name: "new"}

	option := WithChildren(newChild)
	option(record)

	// Both old and new children should be present
	require.Len(t, record.Children, 2)
	assert.Equal(t, "/existing", record.Children[0].Path)
	assert.Equal(t, "/new", record.Children[1].Path)
}

// TestRouterBuilder_BuildWithOptions verifies options work with Build
func TestRouterBuilder_BuildWithOptions(t *testing.T) {
	builder := NewRouterBuilder().
		RouteWithOptions("/test",
			WithName("test-route"),
			WithMeta(map[string]interface{}{
				"requiresAuth": true,
			}),
		)

	router, err := builder.Build()
	require.NoError(t, err)

	// Verify route was registered correctly
	route, found := router.registry.GetByName("test-route")
	require.True(t, found)
	assert.Equal(t, "/test", route.Path)
	assert.Equal(t, true, route.Meta["requiresAuth"])
}
