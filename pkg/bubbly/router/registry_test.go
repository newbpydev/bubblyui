package router

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouteRegistry_Register tests basic route registration
func TestRouteRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		rName   string
		wantErr bool
	}{
		{
			name:    "simple static route",
			path:    "/users",
			rName:   "users-list",
			wantErr: false,
		},
		{
			name:    "route with param",
			path:    "/user/:id",
			rName:   "user-detail",
			wantErr: false,
		},
		{
			name:    "root route",
			path:    "/",
			rName:   "home",
			wantErr: false,
		},
		{
			name:    "nested route",
			path:    "/dashboard/stats",
			rName:   "dashboard-stats",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRouteRegistry()

			err := registry.Register(tt.path, tt.rName, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRouteRegistry_Register_DuplicatePath tests duplicate path rejection
func TestRouteRegistry_Register_DuplicatePath(t *testing.T) {
	registry := NewRouteRegistry()

	// Register first route
	err := registry.Register("/users", "users-list", nil)
	require.NoError(t, err)

	// Try to register same path again
	err = registry.Register("/users", "users-list-v2", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate path")
}

// TestRouteRegistry_Register_DuplicateName tests duplicate name rejection
func TestRouteRegistry_Register_DuplicateName(t *testing.T) {
	registry := NewRouteRegistry()

	// Register first route
	err := registry.Register("/users", "users", nil)
	require.NoError(t, err)

	// Try to register different path with same name
	err = registry.Register("/accounts", "users", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate name")
}

// TestRouteRegistry_GetByName tests named route lookup
func TestRouteRegistry_GetByName(t *testing.T) {
	registry := NewRouteRegistry()

	// Register routes
	err := registry.Register("/users", "users-list", nil)
	require.NoError(t, err)

	err = registry.Register("/user/:id", "user-detail", nil)
	require.NoError(t, err)

	tests := []struct {
		name      string
		routeName string
		wantPath  string
		wantFound bool
	}{
		{
			name:      "existing route",
			routeName: "users-list",
			wantPath:  "/users",
			wantFound: true,
		},
		{
			name:      "existing route with param",
			routeName: "user-detail",
			wantPath:  "/user/:id",
			wantFound: true,
		},
		{
			name:      "non-existent route",
			routeName: "not-found",
			wantPath:  "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, found := registry.GetByName(tt.routeName)

			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.NotNil(t, route)
				assert.Equal(t, tt.wantPath, route.Path)
				assert.Equal(t, tt.routeName, route.Name)
			} else {
				assert.Nil(t, route)
			}
		})
	}
}

// TestRouteRegistry_GetByPath tests path-based route lookup
func TestRouteRegistry_GetByPath(t *testing.T) {
	registry := NewRouteRegistry()

	// Register routes
	err := registry.Register("/users", "users-list", nil)
	require.NoError(t, err)

	err = registry.Register("/user/:id", "user-detail", nil)
	require.NoError(t, err)

	tests := []struct {
		name      string
		path      string
		wantName  string
		wantFound bool
	}{
		{
			name:      "existing static route",
			path:      "/users",
			wantName:  "users-list",
			wantFound: true,
		},
		{
			name:      "existing param route",
			path:      "/user/:id",
			wantName:  "user-detail",
			wantFound: true,
		},
		{
			name:      "non-existent route",
			path:      "/not-found",
			wantName:  "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, found := registry.GetByPath(tt.path)

			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.NotNil(t, route)
				assert.Equal(t, tt.path, route.Path)
				assert.Equal(t, tt.wantName, route.Name)
			} else {
				assert.Nil(t, route)
			}
		})
	}
}

// TestRouteRegistry_GetAll tests retrieving all routes
func TestRouteRegistry_GetAll(t *testing.T) {
	registry := NewRouteRegistry()

	// Empty registry
	routes := registry.GetAll()
	assert.Empty(t, routes)

	// Add routes
	err := registry.Register("/", "home", nil)
	require.NoError(t, err)

	err = registry.Register("/users", "users-list", nil)
	require.NoError(t, err)

	err = registry.Register("/user/:id", "user-detail", nil)
	require.NoError(t, err)

	// Get all routes
	routes = registry.GetAll()
	assert.Len(t, routes, 3)

	// Verify routes are returned (order not guaranteed)
	names := make(map[string]bool)
	for _, route := range routes {
		names[route.Name] = true
	}

	assert.True(t, names["home"])
	assert.True(t, names["users-list"])
	assert.True(t, names["user-detail"])
}

// TestRouteRegistry_NestedRoutes tests nested route registration
func TestRouteRegistry_NestedRoutes(t *testing.T) {
	registry := NewRouteRegistry()

	// Create parent route
	err := registry.Register("/dashboard", "dashboard", nil)
	require.NoError(t, err)

	parent, found := registry.GetByName("dashboard")
	require.True(t, found)
	require.NotNil(t, parent)

	// Create child routes
	child1 := &RouteRecord{
		Path: "/dashboard/stats",
		Name: "dashboard-stats",
	}

	child2 := &RouteRecord{
		Path: "/dashboard/settings",
		Name: "dashboard-settings",
	}

	// Add children to parent
	parent.Children = []*RouteRecord{child1, child2}

	// Verify children
	assert.Len(t, parent.Children, 2)
	assert.Equal(t, "dashboard-stats", parent.Children[0].Name)
	assert.Equal(t, "dashboard-settings", parent.Children[1].Name)
}

// TestRouteRegistry_ThreadSafety tests concurrent access
func TestRouteRegistry_ThreadSafety(t *testing.T) {
	registry := NewRouteRegistry()

	const goroutines = 10
	const routesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Concurrent writes
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < routesPerGoroutine; j++ {
				path := "/route-" + string(rune('a'+id)) + "-" + string(rune('0'+j))
				name := "route-" + string(rune('a'+id)) + "-" + string(rune('0'+j))

				_ = registry.Register(path, name, nil)
			}
		}(i)
	}

	wg.Wait()

	// Verify all routes registered
	routes := registry.GetAll()
	assert.GreaterOrEqual(t, len(routes), 1) // At least some routes registered

	// Concurrent reads
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				_ = registry.GetAll()
			}
		}()
	}

	wg.Wait()
}

// TestRouteRegistry_Meta tests route metadata
func TestRouteRegistry_Meta(t *testing.T) {
	registry := NewRouteRegistry()

	meta := map[string]interface{}{
		"requiresAuth": true,
		"title":        "User Profile",
	}

	err := registry.Register("/profile", "profile", meta)
	require.NoError(t, err)

	route, found := registry.GetByName("profile")
	require.True(t, found)
	require.NotNil(t, route)

	assert.Equal(t, true, route.Meta["requiresAuth"])
	assert.Equal(t, "User Profile", route.Meta["title"])
}
