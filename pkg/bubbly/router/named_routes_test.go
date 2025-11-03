package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildPath_StaticRoute tests building paths for static routes
func TestBuildPath_StaticRoute(t *testing.T) {
	tests := []struct {
		name      string
		routes    map[string]string // name -> path
		routeName string
		params    map[string]string
		query     map[string]string
		expected  string
		wantErr   bool
	}{
		{
			name:      "simple static route",
			routes:    map[string]string{"home": "/"},
			routeName: "home",
			params:    nil,
			query:     nil,
			expected:  "/",
			wantErr:   false,
		},
		{
			name:      "static route with path",
			routes:    map[string]string{"about": "/about"},
			routeName: "about",
			params:    nil,
			query:     nil,
			expected:  "/about",
			wantErr:   false,
		},
		{
			name:      "static route with query",
			routes:    map[string]string{"search": "/search"},
			routeName: "search",
			params:    nil,
			query:     map[string]string{"q": "golang", "page": "1"},
			expected:  "/search?page=1&q=golang",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			// Register routes
			for name, path := range tt.routes {
				err := router.registry.Register(path, name, nil)
				require.NoError(t, err)
			}

			// Build path
			result, err := router.BuildPath(tt.routeName, tt.params, tt.query)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestBuildPath_DynamicRoute tests building paths with parameters
func TestBuildPath_DynamicRoute(t *testing.T) {
	tests := []struct {
		name      string
		routes    map[string]string // name -> path
		routeName string
		params    map[string]string
		query     map[string]string
		expected  string
		wantErr   bool
	}{
		{
			name:      "single param",
			routes:    map[string]string{"user-detail": "/user/:id"},
			routeName: "user-detail",
			params:    map[string]string{"id": "123"},
			query:     nil,
			expected:  "/user/123",
			wantErr:   false,
		},
		{
			name:      "multiple params",
			routes:    map[string]string{"post-comment": "/post/:postId/comment/:commentId"},
			routeName: "post-comment",
			params:    map[string]string{"postId": "42", "commentId": "7"},
			query:     nil,
			expected:  "/post/42/comment/7",
			wantErr:   false,
		},
		{
			name:      "params with query",
			routes:    map[string]string{"user-detail": "/user/:id"},
			routeName: "user-detail",
			params:    map[string]string{"id": "123"},
			query:     map[string]string{"tab": "profile"},
			expected:  "/user/123?tab=profile",
			wantErr:   false,
		},
		{
			name:      "missing required param",
			routes:    map[string]string{"user-detail": "/user/:id"},
			routeName: "user-detail",
			params:    nil,
			query:     nil,
			expected:  "",
			wantErr:   true,
		},
		{
			name:      "missing one of multiple params",
			routes:    map[string]string{"post-comment": "/post/:postId/comment/:commentId"},
			routeName: "post-comment",
			params:    map[string]string{"postId": "42"},
			query:     nil,
			expected:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			// Register routes
			for name, path := range tt.routes {
				err := router.registry.Register(path, name, nil)
				require.NoError(t, err)
			}

			// Build path
			result, err := router.BuildPath(tt.routeName, tt.params, tt.query)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestBuildPath_OptionalParams tests building paths with optional parameters
func TestBuildPath_OptionalParams(t *testing.T) {
	tests := []struct {
		name      string
		routes    map[string]string // name -> path
		routeName string
		params    map[string]string
		query     map[string]string
		expected  string
		wantErr   bool
	}{
		{
			name:      "optional param provided",
			routes:    map[string]string{"profile": "/profile/:id?"},
			routeName: "profile",
			params:    map[string]string{"id": "123"},
			query:     nil,
			expected:  "/profile/123",
			wantErr:   false,
		},
		{
			name:      "optional param omitted",
			routes:    map[string]string{"profile": "/profile/:id?"},
			routeName: "profile",
			params:    nil,
			query:     nil,
			expected:  "/profile",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			// Register routes
			for name, path := range tt.routes {
				err := router.registry.Register(path, name, nil)
				require.NoError(t, err)
			}

			// Build path
			result, err := router.BuildPath(tt.routeName, tt.params, tt.query)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestBuildPath_Wildcards tests building paths with wildcard parameters
func TestBuildPath_Wildcards(t *testing.T) {
	tests := []struct {
		name      string
		routes    map[string]string // name -> path
		routeName string
		params    map[string]string
		query     map[string]string
		expected  string
		wantErr   bool
	}{
		{
			name:      "wildcard with single segment",
			routes:    map[string]string{"docs": "/docs/:path*"},
			routeName: "docs",
			params:    map[string]string{"path": "guide"},
			query:     nil,
			expected:  "/docs/guide",
			wantErr:   false,
		},
		{
			name:      "wildcard with multiple segments",
			routes:    map[string]string{"docs": "/docs/:path*"},
			routeName: "docs",
			params:    map[string]string{"path": "guide/getting-started"},
			query:     nil,
			expected:  "/docs/guide/getting-started",
			wantErr:   false,
		},
		{
			name:      "wildcard omitted",
			routes:    map[string]string{"docs": "/docs/:path*"},
			routeName: "docs",
			params:    nil,
			query:     nil,
			expected:  "/docs",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			// Register routes
			for name, path := range tt.routes {
				err := router.registry.Register(path, name, nil)
				require.NoError(t, err)
			}

			// Build path
			result, err := router.BuildPath(tt.routeName, tt.params, tt.query)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestBuildPath_InvalidName tests error handling for invalid route names
func TestBuildPath_InvalidName(t *testing.T) {
	router := NewRouter()
	err := router.registry.Register("/users", "users-list", nil)
	require.NoError(t, err)

	// Try to build path for non-existent route
	result, err := router.BuildPath("non-existent", nil, nil)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "route not found")
}

// TestPushNamed_BasicNavigation tests basic named route navigation
func TestPushNamed_BasicNavigation(t *testing.T) {
	router := NewRouter()

	// Register routes
	err := router.registry.Register("/", "home", nil)
	require.NoError(t, err)
	err = router.registry.Register("/about", "about", nil)
	require.NoError(t, err)
	err = router.registry.Register("/user/:id", "user-detail", nil)
	require.NoError(t, err)

	tests := []struct {
		name      string
		routeName string
		params    map[string]string
		query     map[string]string
		wantErr   bool
	}{
		{
			name:      "navigate to static route",
			routeName: "home",
			params:    nil,
			query:     nil,
			wantErr:   false,
		},
		{
			name:      "navigate to route with param",
			routeName: "user-detail",
			params:    map[string]string{"id": "123"},
			query:     nil,
			wantErr:   false,
		},
		{
			name:      "navigate with query string",
			routeName: "about",
			params:    nil,
			query:     map[string]string{"ref": "footer"},
			wantErr:   false,
		},
		{
			name:      "invalid route name",
			routeName: "non-existent",
			params:    nil,
			query:     nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := router.PushNamed(tt.routeName, tt.params, tt.query)

			if tt.wantErr {
				// Command should be nil or error command
				assert.Nil(t, cmd)
			} else {
				// Command should be non-nil
				assert.NotNil(t, cmd)
			}
		})
	}
}

// TestPushNamed_ParamInjection tests parameter injection in named routes
func TestPushNamed_ParamInjection(t *testing.T) {
	router := NewRouter()

	// Register route with multiple params
	err := router.registry.Register("/post/:postId/comment/:commentId", "post-comment", nil)
	require.NoError(t, err)

	// Navigate with params
	cmd := router.PushNamed("post-comment", map[string]string{
		"postId":    "42",
		"commentId": "7",
	}, nil)

	assert.NotNil(t, cmd)

	// Verify the route was built correctly by checking current route after navigation
	// (This would require executing the command in a real Bubbletea program)
}

// TestPushNamed_MissingParams tests error handling for missing required params
func TestPushNamed_MissingParams(t *testing.T) {
	router := NewRouter()

	// Register route with required param
	err := router.registry.Register("/user/:id", "user-detail", nil)
	require.NoError(t, err)

	// Try to navigate without required param
	cmd := router.PushNamed("user-detail", nil, nil)

	// Should return nil (error case)
	assert.Nil(t, cmd)
}
