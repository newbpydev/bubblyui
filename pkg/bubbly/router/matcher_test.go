package router

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRouteMatcher tests matcher creation
func TestNewRouteMatcher(t *testing.T) {
	matcher := NewRouteMatcher()
	assert.NotNil(t, matcher)
	assert.Empty(t, matcher.routes)
}

// TestRouteMatcher_AddRoute tests route registration
func TestRouteMatcher_AddRoute(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid static route",
			path:    "/users",
			wantErr: false,
		},
		{
			name:    "valid param route",
			path:    "/user/:id",
			wantErr: false,
		},
		{
			name:    "valid optional route",
			path:    "/profile/:id?",
			wantErr: false,
		},
		{
			name:    "valid wildcard route",
			path:    "/docs/:path*",
			wantErr: false,
		},
		{
			name:    "invalid empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "invalid no leading slash",
			path:    "users",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewRouteMatcher()
			err := matcher.AddRoute(tt.path, "route-"+tt.name)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, matcher.routes, 1)
			}
		})
	}
}

// TestRouteMatcher_Match_StaticRoutes tests static route matching
func TestRouteMatcher_Match_StaticRoutes(t *testing.T) {
	matcher := NewRouteMatcher()
	require.NoError(t, matcher.AddRoute("/", "home"))
	require.NoError(t, matcher.AddRoute("/users", "users"))
	require.NoError(t, matcher.AddRoute("/users/list", "users-list"))
	require.NoError(t, matcher.AddRoute("/about", "about"))

	tests := []struct {
		name       string
		path       string
		wantName   string
		wantParams map[string]string
		wantErr    bool
	}{
		{
			name:       "root path",
			path:       "/",
			wantName:   "home",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "simple path",
			path:       "/users",
			wantName:   "users",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "nested path",
			path:       "/users/list",
			wantName:   "users-list",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "trailing slash",
			path:       "/about/",
			wantName:   "about",
			wantParams: map[string]string{},
			wantErr:    false,
		},
		{
			name:    "not found",
			path:    "/notfound",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matcher.Match(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrNoMatch))
				assert.Nil(t, match)
			} else {
				require.NoError(t, err)
				require.NotNil(t, match)
				assert.Equal(t, tt.wantName, match.Route.Name)
				assert.Equal(t, tt.wantParams, match.Params)
			}
		})
	}
}

// TestRouteMatcher_Match_DynamicParams tests parameter extraction
func TestRouteMatcher_Match_DynamicParams(t *testing.T) {
	matcher := NewRouteMatcher()
	require.NoError(t, matcher.AddRoute("/user/:id", "user-detail"))
	require.NoError(t, matcher.AddRoute("/post/:postId/comment/:commentId", "comment"))

	tests := []struct {
		name       string
		path       string
		wantName   string
		wantParams map[string]string
		wantErr    bool
	}{
		{
			name:     "single param",
			path:     "/user/123",
			wantName: "user-detail",
			wantParams: map[string]string{
				"id": "123",
			},
			wantErr: false,
		},
		{
			name:     "multiple params",
			path:     "/post/456/comment/789",
			wantName: "comment",
			wantParams: map[string]string{
				"postId":    "456",
				"commentId": "789",
			},
			wantErr: false,
		},
		{
			name:    "missing param segment",
			path:    "/user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matcher.Match(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, match)
				assert.Equal(t, tt.wantName, match.Route.Name)
				assert.Equal(t, tt.wantParams, match.Params)
			}
		})
	}
}

// TestRouteMatcher_Match_OptionalParams tests optional parameters
func TestRouteMatcher_Match_OptionalParams(t *testing.T) {
	matcher := NewRouteMatcher()
	require.NoError(t, matcher.AddRoute("/profile/:id?", "profile"))

	tests := []struct {
		name       string
		path       string
		wantParams map[string]string
	}{
		{
			name: "with optional param",
			path: "/profile/123",
			wantParams: map[string]string{
				"id": "123",
			},
		},
		{
			name:       "without optional param",
			path:       "/profile",
			wantParams: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matcher.Match(tt.path)

			require.NoError(t, err)
			require.NotNil(t, match)
			assert.Equal(t, "profile", match.Route.Name)
			assert.Equal(t, tt.wantParams, match.Params)
		})
	}
}

// TestRouteMatcher_Match_Wildcards tests wildcard matching
func TestRouteMatcher_Match_Wildcards(t *testing.T) {
	matcher := NewRouteMatcher()
	require.NoError(t, matcher.AddRoute("/docs/:path*", "docs"))

	tests := []struct {
		name       string
		path       string
		wantParams map[string]string
	}{
		{
			name: "single level wildcard",
			path: "/docs/guide",
			wantParams: map[string]string{
				"path": "guide",
			},
		},
		{
			name: "multi level wildcard",
			path: "/docs/guide/getting-started",
			wantParams: map[string]string{
				"path": "guide/getting-started",
			},
		},
		{
			name:       "empty wildcard",
			path:       "/docs",
			wantParams: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matcher.Match(tt.path)

			require.NoError(t, err)
			require.NotNil(t, match)
			assert.Equal(t, "docs", match.Route.Name)

			// Handle optional wildcard (empty string)
			if len(tt.wantParams) == 0 {
				// Either no param or empty param is acceptable
				if len(match.Params) > 0 {
					assert.Equal(t, "", match.Params["path"])
				}
			} else {
				assert.Equal(t, tt.wantParams, match.Params)
			}
		})
	}
}

// TestRouteMatcher_Match_Precedence tests route specificity
func TestRouteMatcher_Match_Precedence(t *testing.T) {
	matcher := NewRouteMatcher()
	require.NoError(t, matcher.AddRoute("/users/:id", "user-param"))
	require.NoError(t, matcher.AddRoute("/users/new", "user-new"))
	require.NoError(t, matcher.AddRoute("/users/list", "user-list"))
	require.NoError(t, matcher.AddRoute("/:resource/:id", "generic"))

	tests := []struct {
		name     string
		path     string
		wantName string
		reason   string
	}{
		{
			name:     "static wins over param",
			path:     "/users/new",
			wantName: "user-new",
			reason:   "static segment more specific than param",
		},
		{
			name:     "static wins over param (list)",
			path:     "/users/list",
			wantName: "user-list",
			reason:   "static segment more specific than param",
		},
		{
			name:     "more specific param wins",
			path:     "/users/123",
			wantName: "user-param",
			reason:   "more static segments wins",
		},
		{
			name:     "least specific matches",
			path:     "/posts/456",
			wantName: "generic",
			reason:   "only route that matches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matcher.Match(tt.path)

			require.NoError(t, err, "should match: %s", tt.reason)
			require.NotNil(t, match)
			assert.Equal(t, tt.wantName, match.Route.Name, tt.reason)
		})
	}
}

// TestRouteMatcher_Match_Scoring tests score calculation
func TestRouteMatcher_Match_Scoring(t *testing.T) {
	tests := []struct {
		name                 string
		path                 string
		wantStaticSegments   int
		wantParamSegments    int
		wantOptionalSegments int
		wantWildcardSegments int
	}{
		{
			name:               "all static",
			path:               "/users/list",
			wantStaticSegments: 2,
		},
		{
			name:               "static and param",
			path:               "/user/:id",
			wantStaticSegments: 1,
			wantParamSegments:  1,
		},
		{
			name:                 "with optional",
			path:                 "/profile/:id?",
			wantStaticSegments:   1,
			wantOptionalSegments: 1,
		},
		{
			name:                 "with wildcard",
			path:                 "/docs/:path*",
			wantStaticSegments:   1,
			wantWildcardSegments: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewRouteMatcher()
			require.NoError(t, matcher.AddRoute(tt.path, "test"))

			match, err := matcher.Match(tt.path)
			require.NoError(t, err)
			require.NotNil(t, match)

			assert.Equal(t, tt.wantStaticSegments, match.Score.staticSegments)
			assert.Equal(t, tt.wantParamSegments, match.Score.paramSegments)
			assert.Equal(t, tt.wantOptionalSegments, match.Score.optionalSegments)
			assert.Equal(t, tt.wantWildcardSegments, match.Score.wildcardSegments)
		})
	}
}

// BenchmarkRouteMatcher_Match benchmarks matching performance
func BenchmarkRouteMatcher_Match(b *testing.B) {
	matcher := NewRouteMatcher()
	_ = matcher.AddRoute("/", "home")
	_ = matcher.AddRoute("/users", "users")
	_ = matcher.AddRoute("/user/:id", "user")
	_ = matcher.AddRoute("/posts/:id/comments/:commentId", "comment")
	_ = matcher.AddRoute("/docs/:path*", "docs")

	paths := []string{
		"/",
		"/users",
		"/user/123",
		"/posts/456/comments/789",
		"/docs/guide/getting-started",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		_, _ = matcher.Match(path)
	}
}
