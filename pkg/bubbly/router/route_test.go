package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRoute tests route creation
func TestNewRoute(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		routeName string
		params    map[string]string
		query     map[string]string
		hash      string
		meta      map[string]interface{}
		matched   []*RouteRecord
	}{
		{
			name:      "simple route",
			path:      "/users",
			routeName: "users-list",
			params:    map[string]string{},
			query:     map[string]string{},
			hash:      "",
			meta:      map[string]interface{}{},
			matched:   []*RouteRecord{},
		},
		{
			name:      "route with params",
			path:      "/user/:id",
			routeName: "user-detail",
			params: map[string]string{
				"id": "123",
			},
			query:   map[string]string{},
			hash:    "",
			meta:    map[string]interface{}{},
			matched: []*RouteRecord{},
		},
		{
			name:      "route with query",
			path:      "/search",
			routeName: "search",
			params:    map[string]string{},
			query: map[string]string{
				"q":    "golang",
				"page": "1",
			},
			hash:    "",
			meta:    map[string]interface{}{},
			matched: []*RouteRecord{},
		},
		{
			name:      "route with hash",
			path:      "/docs",
			routeName: "docs",
			params:    map[string]string{},
			query:     map[string]string{},
			hash:      "section-1",
			meta:      map[string]interface{}{},
			matched:   []*RouteRecord{},
		},
		{
			name:      "route with meta",
			path:      "/admin",
			routeName: "admin",
			params:    map[string]string{},
			query:     map[string]string{},
			hash:      "",
			meta: map[string]interface{}{
				"requiresAuth": true,
				"title":        "Admin Panel",
			},
			matched: []*RouteRecord{},
		},
		{
			name:      "route with matched chain",
			path:      "/dashboard/stats",
			routeName: "dashboard-stats",
			params:    map[string]string{},
			query:     map[string]string{},
			hash:      "",
			meta:      map[string]interface{}{},
			matched: []*RouteRecord{
				{Path: "/dashboard", Name: "dashboard"},
				{Path: "/dashboard/stats", Name: "dashboard-stats"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := NewRoute(tt.path, tt.routeName, tt.params, tt.query, tt.hash, tt.meta, tt.matched)

			assert.Equal(t, tt.path, route.Path)
			assert.Equal(t, tt.routeName, route.Name)
			assert.Equal(t, tt.params, route.Params)
			assert.Equal(t, tt.query, route.Query)
			assert.Equal(t, tt.hash, route.Hash)
			assert.Equal(t, tt.meta, route.Meta)
			assert.Equal(t, len(tt.matched), len(route.Matched))
		})
	}
}

// TestRoute_FullPath tests FullPath generation
func TestRoute_FullPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		query        map[string]string
		hash         string
		expectedFull string
	}{
		{
			name:         "path only",
			path:         "/users",
			query:        map[string]string{},
			hash:         "",
			expectedFull: "/users",
		},
		{
			name: "path with query",
			path: "/search",
			query: map[string]string{
				"q": "golang",
			},
			hash:         "",
			expectedFull: "/search?q=golang",
		},
		{
			name: "path with multiple query params",
			path: "/search",
			query: map[string]string{
				"q":    "golang",
				"page": "1",
			},
			hash:         "",
			expectedFull: "/search?page=1&q=golang", // Sorted alphabetically
		},
		{
			name:         "path with hash",
			path:         "/docs",
			query:        map[string]string{},
			hash:         "section-1",
			expectedFull: "/docs#section-1",
		},
		{
			name: "path with query and hash",
			path: "/docs",
			query: map[string]string{
				"version": "1.0",
			},
			hash:         "api",
			expectedFull: "/docs?version=1.0#api",
		},
		{
			name:         "root path",
			path:         "/",
			query:        map[string]string{},
			hash:         "",
			expectedFull: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := NewRoute(tt.path, "test", map[string]string{}, tt.query, tt.hash, nil, nil)
			assert.Equal(t, tt.expectedFull, route.FullPath)
		})
	}
}

// TestRoute_Immutability tests defensive copying
func TestRoute_Immutability(t *testing.T) {
	t.Run("params immutability", func(t *testing.T) {
		originalParams := map[string]string{
			"id": "123",
		}

		route := NewRoute("/user/:id", "user", originalParams, nil, "", nil, nil)

		// Modify original params
		originalParams["id"] = "456"
		originalParams["new"] = "value"

		// Route params should be unchanged
		assert.Equal(t, "123", route.Params["id"])
		assert.NotContains(t, route.Params, "new")
	})

	t.Run("query immutability", func(t *testing.T) {
		originalQuery := map[string]string{
			"page": "1",
		}

		route := NewRoute("/search", "search", nil, originalQuery, "", nil, nil)

		// Modify original query
		originalQuery["page"] = "2"
		originalQuery["new"] = "value"

		// Route query should be unchanged
		assert.Equal(t, "1", route.Query["page"])
		assert.NotContains(t, route.Query, "new")
	})

	t.Run("meta immutability", func(t *testing.T) {
		originalMeta := map[string]interface{}{
			"requiresAuth": true,
		}

		route := NewRoute("/admin", "admin", nil, nil, "", originalMeta, nil)

		// Modify original meta
		originalMeta["requiresAuth"] = false
		originalMeta["new"] = "value"

		// Route meta should be unchanged
		assert.Equal(t, true, route.Meta["requiresAuth"])
		assert.NotContains(t, route.Meta, "new")
	})

	t.Run("matched immutability", func(t *testing.T) {
		originalMatched := []*RouteRecord{
			{Path: "/dashboard", Name: "dashboard"},
		}

		route := NewRoute("/dashboard/stats", "stats", nil, nil, "", nil, originalMatched)

		// Append to original slice - this should not affect the route
		originalMatched = append(originalMatched, &RouteRecord{Path: "/new", Name: "new"})

		// Route matched slice should be unchanged (slice is copied)
		assert.Len(t, route.Matched, 1)
		assert.Equal(t, "/dashboard", route.Matched[0].Path)

		// Note: The RouteRecord pointers themselves are shared (shallow copy)
		// This is correct behavior - RouteRecords should not be modified after creation
		// They are managed by the router registry
	})
}

// TestRoute_GetMeta tests meta field access
func TestRoute_GetMeta(t *testing.T) {
	meta := map[string]interface{}{
		"requiresAuth": true,
		"title":        "Admin Panel",
		"roles":        []string{"admin", "superuser"},
	}

	route := NewRoute("/admin", "admin", nil, nil, "", meta, nil)

	tests := []struct {
		name      string
		key       string
		wantValue interface{}
		wantFound bool
	}{
		{
			name:      "existing boolean meta",
			key:       "requiresAuth",
			wantValue: true,
			wantFound: true,
		},
		{
			name:      "existing string meta",
			key:       "title",
			wantValue: "Admin Panel",
			wantFound: true,
		},
		{
			name:      "existing slice meta",
			key:       "roles",
			wantValue: []string{"admin", "superuser"},
			wantFound: true,
		},
		{
			name:      "non-existent meta",
			key:       "notFound",
			wantValue: nil,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found := route.GetMeta(tt.key)
			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}

// TestRoute_NilMaps tests handling of nil maps
func TestRoute_NilMaps(t *testing.T) {
	route := NewRoute("/test", "test", nil, nil, "", nil, nil)

	assert.NotNil(t, route.Params)
	assert.NotNil(t, route.Query)
	assert.NotNil(t, route.Meta)
	assert.NotNil(t, route.Matched)

	assert.Empty(t, route.Params)
	assert.Empty(t, route.Query)
	assert.Empty(t, route.Meta)
	assert.Empty(t, route.Matched)
}

// TestRoute_EmptyValues tests handling of empty values
func TestRoute_EmptyValues(t *testing.T) {
	route := NewRoute("", "", map[string]string{}, map[string]string{}, "", map[string]interface{}{}, []*RouteRecord{})

	assert.Equal(t, "", route.Path)
	assert.Equal(t, "", route.Name)
	assert.Equal(t, "", route.Hash)
	assert.Equal(t, "", route.FullPath)
	assert.Empty(t, route.Params)
	assert.Empty(t, route.Query)
	assert.Empty(t, route.Meta)
	assert.Empty(t, route.Matched)
}

// TestRoute_MatchedChain tests nested route matching
func TestRoute_MatchedChain(t *testing.T) {
	parent := &RouteRecord{
		Path: "/dashboard",
		Name: "dashboard",
		Meta: map[string]interface{}{
			"layout": "admin",
		},
	}

	child := &RouteRecord{
		Path: "/dashboard/stats",
		Name: "dashboard-stats",
		Meta: map[string]interface{}{
			"title": "Statistics",
		},
	}

	route := NewRoute("/dashboard/stats", "dashboard-stats", nil, nil, "", nil, []*RouteRecord{parent, child})

	require.Len(t, route.Matched, 2)
	assert.Equal(t, "/dashboard", route.Matched[0].Path)
	assert.Equal(t, "dashboard", route.Matched[0].Name)
	assert.Equal(t, "/dashboard/stats", route.Matched[1].Path)
	assert.Equal(t, "dashboard-stats", route.Matched[1].Name)
}

// TestRoute_MetaInheritancePattern tests accessing parent meta via matched array
// This follows Vue Router's pattern where meta fields are NOT automatically inherited,
// but can be accessed by iterating through the matched array
func TestRoute_MetaInheritancePattern(t *testing.T) {
	t.Run("access parent meta via matched array", func(t *testing.T) {
		parent := &RouteRecord{
			Path: "/dashboard",
			Name: "dashboard",
			Meta: map[string]interface{}{
				"requiresAuth": true,
				"layout":       "admin",
			},
		}

		child := &RouteRecord{
			Path: "/dashboard/stats",
			Name: "dashboard-stats",
			Meta: map[string]interface{}{
				"title": "Statistics",
			},
		}

		route := NewRoute("/dashboard/stats", "dashboard-stats", nil, nil, "", nil, []*RouteRecord{parent, child})

		// Child route's direct meta should NOT contain parent's meta
		_, hasAuth := route.GetMeta("requiresAuth")
		assert.False(t, hasAuth, "child route should not inherit parent meta directly")

		// But we can access parent meta via matched array (Vue Router pattern)
		var requiresAuth bool
		for _, record := range route.Matched {
			if auth, ok := record.Meta["requiresAuth"]; ok {
				requiresAuth = auth.(bool)
				break
			}
		}
		assert.True(t, requiresAuth, "should find requiresAuth in parent via matched array")

		// Child's own meta should be accessible
		title, hasTitle := route.Matched[1].Meta["title"]
		assert.True(t, hasTitle)
		assert.Equal(t, "Statistics", title)
	})

	t.Run("check meta in navigation guard pattern", func(t *testing.T) {
		// Simulate Vue Router's beforeEach guard pattern
		parent := &RouteRecord{
			Path: "/admin",
			Name: "admin",
			Meta: map[string]interface{}{
				"requiresAuth": true,
				"roles":        []string{"admin", "superuser"},
			},
		}

		child := &RouteRecord{
			Path: "/admin/users",
			Name: "admin-users",
			Meta: map[string]interface{}{
				"title": "User Management",
			},
		}

		route := NewRoute("/admin/users", "admin-users", nil, nil, "", nil, []*RouteRecord{parent, child})

		// Check if any matched record requires auth (like Vue Router's to.matched.some())
		requiresAuth := false
		for _, record := range route.Matched {
			if auth, ok := record.Meta["requiresAuth"]; ok && auth.(bool) {
				requiresAuth = true
				break
			}
		}
		assert.True(t, requiresAuth, "should detect requiresAuth in matched chain")

		// Check roles from parent
		var roles []string
		for _, record := range route.Matched {
			if r, ok := record.Meta["roles"]; ok {
				roles = r.([]string)
				break
			}
		}
		assert.Equal(t, []string{"admin", "superuser"}, roles)
	})

	t.Run("deeply nested meta access", func(t *testing.T) {
		grandparent := &RouteRecord{
			Path: "/app",
			Name: "app",
			Meta: map[string]interface{}{
				"theme": "dark",
			},
		}

		parent := &RouteRecord{
			Path: "/app/dashboard",
			Name: "dashboard",
			Meta: map[string]interface{}{
				"requiresAuth": true,
			},
		}

		child := &RouteRecord{
			Path: "/app/dashboard/stats",
			Name: "stats",
			Meta: map[string]interface{}{
				"refreshInterval": 5000,
			},
		}

		route := NewRoute("/app/dashboard/stats", "stats", nil, nil, "", nil,
			[]*RouteRecord{grandparent, parent, child})

		// Collect all meta from matched chain
		allMeta := make(map[string]interface{})
		for _, record := range route.Matched {
			for k, v := range record.Meta {
				allMeta[k] = v
			}
		}

		assert.Equal(t, "dark", allMeta["theme"])
		assert.Equal(t, true, allMeta["requiresAuth"])
		assert.Equal(t, 5000, allMeta["refreshInterval"])
	})
}

// TestRoute_MetaTypeAssertions tests type assertions for various meta value types
func TestRoute_MetaTypeAssertions(t *testing.T) {
	tests := []struct {
		name          string
		meta          map[string]interface{}
		key           string
		expectedType  string
		assertionFunc func(t *testing.T, value interface{})
	}{
		{
			name: "boolean type",
			meta: map[string]interface{}{
				"requiresAuth": true,
			},
			key:          "requiresAuth",
			expectedType: "bool",
			assertionFunc: func(t *testing.T, value interface{}) {
				boolVal, ok := value.(bool)
				assert.True(t, ok, "should be bool type")
				assert.True(t, boolVal)
			},
		},
		{
			name: "string type",
			meta: map[string]interface{}{
				"title": "Dashboard",
			},
			key:          "title",
			expectedType: "string",
			assertionFunc: func(t *testing.T, value interface{}) {
				strVal, ok := value.(string)
				assert.True(t, ok, "should be string type")
				assert.Equal(t, "Dashboard", strVal)
			},
		},
		{
			name: "int type",
			meta: map[string]interface{}{
				"maxRetries": 3,
			},
			key:          "maxRetries",
			expectedType: "int",
			assertionFunc: func(t *testing.T, value interface{}) {
				intVal, ok := value.(int)
				assert.True(t, ok, "should be int type")
				assert.Equal(t, 3, intVal)
			},
		},
		{
			name: "string slice type",
			meta: map[string]interface{}{
				"roles": []string{"admin", "user"},
			},
			key:          "roles",
			expectedType: "[]string",
			assertionFunc: func(t *testing.T, value interface{}) {
				sliceVal, ok := value.([]string)
				assert.True(t, ok, "should be []string type")
				assert.Equal(t, []string{"admin", "user"}, sliceVal)
			},
		},
		{
			name: "map type",
			meta: map[string]interface{}{
				"permissions": map[string]bool{
					"read":  true,
					"write": false,
				},
			},
			key:          "permissions",
			expectedType: "map[string]bool",
			assertionFunc: func(t *testing.T, value interface{}) {
				mapVal, ok := value.(map[string]bool)
				assert.True(t, ok, "should be map[string]bool type")
				assert.True(t, mapVal["read"])
				assert.False(t, mapVal["write"])
			},
		},
		{
			name: "struct type",
			meta: map[string]interface{}{
				"config": struct {
					Timeout int
					Retry   bool
				}{
					Timeout: 5000,
					Retry:   true,
				},
			},
			key:          "config",
			expectedType: "struct",
			assertionFunc: func(t *testing.T, value interface{}) {
				type Config struct {
					Timeout int
					Retry   bool
				}
				configVal, ok := value.(struct {
					Timeout int
					Retry   bool
				})
				assert.True(t, ok, "should be struct type")
				assert.Equal(t, 5000, configVal.Timeout)
				assert.True(t, configVal.Retry)
			},
		},
		{
			name: "float64 type",
			meta: map[string]interface{}{
				"version": 1.5,
			},
			key:          "version",
			expectedType: "float64",
			assertionFunc: func(t *testing.T, value interface{}) {
				floatVal, ok := value.(float64)
				assert.True(t, ok, "should be float64 type")
				assert.Equal(t, 1.5, floatVal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := NewRoute("/test", "test", nil, nil, "", tt.meta, nil)

			value, found := route.GetMeta(tt.key)
			assert.True(t, found, "meta key should exist")
			assert.NotNil(t, value, "meta value should not be nil")

			tt.assertionFunc(t, value)
		})
	}
}

// TestRoute_MetaFieldsSet tests that meta fields are properly set
func TestRoute_MetaFieldsSet(t *testing.T) {
	t.Run("meta fields set on route creation", func(t *testing.T) {
		meta := map[string]interface{}{
			"requiresAuth": true,
			"title":        "Admin Panel",
			"roles":        []string{"admin"},
		}

		route := NewRoute("/admin", "admin", nil, nil, "", meta, nil)

		assert.NotNil(t, route.Meta)
		assert.Len(t, route.Meta, 3)

		requiresAuth, ok := route.GetMeta("requiresAuth")
		assert.True(t, ok)
		assert.Equal(t, true, requiresAuth)

		title, ok := route.GetMeta("title")
		assert.True(t, ok)
		assert.Equal(t, "Admin Panel", title)

		roles, ok := route.GetMeta("roles")
		assert.True(t, ok)
		assert.Equal(t, []string{"admin"}, roles)
	})

	t.Run("meta fields accessible directly from map", func(t *testing.T) {
		meta := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		route := NewRoute("/test", "test", nil, nil, "", meta, nil)

		// Direct map access
		assert.Equal(t, "value1", route.Meta["key1"])
		assert.Equal(t, 42, route.Meta["key2"])

		// GetMeta method
		val1, ok1 := route.GetMeta("key1")
		assert.True(t, ok1)
		assert.Equal(t, "value1", val1)

		val2, ok2 := route.GetMeta("key2")
		assert.True(t, ok2)
		assert.Equal(t, 42, val2)
	})
}
