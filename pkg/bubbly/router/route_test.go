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
