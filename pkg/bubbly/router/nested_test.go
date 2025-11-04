package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChild_CreatesChildRoute tests that Child() creates a child route record.
func TestChild_CreatesChildRoute(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		routeName string
		opts      []RouteOption
		wantPath  string
		wantName  string
	}{
		{
			name:      "simple child route",
			path:      "/profile",
			routeName: "",
			opts:      []RouteOption{WithName("user-profile")},
			wantPath:  "/profile",
			wantName:  "user-profile",
		},
		{
			name:      "child route with params",
			path:      "/:tab",
			routeName: "",
			opts:      []RouteOption{WithName("user-tab")},
			wantPath:  "/:tab",
			wantName:  "user-tab",
		},
		{
			name:      "child route with metadata",
			path:      "/settings",
			routeName: "",
			opts: []RouteOption{
				WithName("user-settings"),
				WithMeta(map[string]interface{}{"requiresAuth": true}),
			},
			wantPath: "/settings",
			wantName: "user-settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := Child(tt.path, tt.opts...)

			assert.NotNil(t, child)
			assert.Equal(t, tt.wantPath, child.Path)
			assert.Equal(t, tt.wantName, child.Name)
		})
	}
}

// TestRouteRecord_ParentChildLinks tests that parent-child relationships are established.
func TestRouteRecord_ParentChildLinks(t *testing.T) {
	// Create parent and child routes
	childRoute := Child("/profile", WithName("user-profile"))
	parentRoute := &RouteRecord{
		Path:     "/user/:id",
		Name:     "user",
		Children: []*RouteRecord{childRoute},
	}

	// Establish parent link (this should be done by the system)
	childRoute.Parent = parentRoute

	// Verify bidirectional link
	assert.NotNil(t, parentRoute.Children)
	assert.Len(t, parentRoute.Children, 1)
	assert.Equal(t, childRoute, parentRoute.Children[0])
	assert.Equal(t, parentRoute, childRoute.Parent)
}

// TestNestedRoutes_PathResolution tests that nested route paths are resolved correctly.
func TestNestedRoutes_PathResolution(t *testing.T) {
	tests := []struct {
		name       string
		parentPath string
		childPath  string
		wantFull   string
	}{
		{
			name:       "simple nested path",
			parentPath: "/user/:id",
			childPath:  "/profile",
			wantFull:   "/user/:id/profile",
		},
		{
			name:       "nested path with child param",
			parentPath: "/user/:id",
			childPath:  "/:tab",
			wantFull:   "/user/:id/:tab",
		},
		{
			name:       "deeply nested path",
			parentPath: "/dashboard",
			childPath:  "/settings",
			wantFull:   "/dashboard/settings",
		},
		{
			name:       "child with empty path (default child)",
			parentPath: "/user/:id",
			childPath:  "",
			wantFull:   "/user/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := resolveNestedPath(tt.parentPath, tt.childPath)
			assert.Equal(t, tt.wantFull, fullPath)
		})
	}
}

// TestMatcher_MatchedArray tests that the Matched array contains parent and child routes.
func TestMatcher_MatchedArray(t *testing.T) {
	// Create nested route structure
	childProfile := Child("/profile", WithName("user-profile"))
	childSettings := Child("/settings", WithName("user-settings"))

	parentRoute := &RouteRecord{
		Path:     "/user/:id",
		Name:     "user",
		Children: []*RouteRecord{childProfile, childSettings},
	}

	// Establish parent links
	childProfile.Parent = parentRoute
	childSettings.Parent = parentRoute

	// Create matcher and add routes
	matcher := NewRouteMatcher()

	// Add parent route (should also register children)
	err := matcher.AddRouteRecord(parentRoute)
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		wantMatched int
		wantRoutes  []string
	}{
		{
			name:        "match parent only",
			path:        "/user/123",
			wantMatched: 1,
			wantRoutes:  []string{"user"},
		},
		{
			name:        "match parent and child",
			path:        "/user/123/profile",
			wantMatched: 2,
			wantRoutes:  []string{"user", "user-profile"},
		},
		{
			name:        "match different child",
			path:        "/user/456/settings",
			wantMatched: 2,
			wantRoutes:  []string{"user", "user-settings"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matcher.Match(tt.path)
			require.NoError(t, err)

			assert.Len(t, match.Matched, tt.wantMatched)

			for i, expectedName := range tt.wantRoutes {
				assert.Equal(t, expectedName, match.Matched[i].Name)
			}
		})
	}
}

// TestNestedRoutes_ParamsExtraction tests that params from parent and child are combined.
func TestNestedRoutes_ParamsExtraction(t *testing.T) {
	// Create nested route with params in both parent and child
	childRoute := Child("/:tab", WithName("user-tab"))
	parentRoute := &RouteRecord{
		Path:     "/user/:id",
		Name:     "user",
		Children: []*RouteRecord{childRoute},
	}
	childRoute.Parent = parentRoute

	matcher := NewRouteMatcher()
	err := matcher.AddRouteRecord(parentRoute)
	require.NoError(t, err)

	tests := []struct {
		name       string
		path       string
		wantParams map[string]string
	}{
		{
			name: "parent params only",
			path: "/user/123",
			wantParams: map[string]string{
				"id": "123",
			},
		},
		{
			name: "parent and child params",
			path: "/user/456/profile",
			wantParams: map[string]string{
				"id":  "456",
				"tab": "profile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := matcher.Match(tt.path)
			require.NoError(t, err)

			for key, expectedValue := range tt.wantParams {
				assert.Equal(t, expectedValue, match.Params[key])
			}
		})
	}
}

// TestWithChildren_Option tests the WithChildren option for adding child routes.
func TestWithChildren_Option(t *testing.T) {
	child1 := Child("/profile", WithName("profile"))
	child2 := Child("/settings", WithName("settings"))

	parent := &RouteRecord{
		Path: "/user/:id",
		Name: "user",
	}

	// Apply WithChildren option
	opt := WithChildren(child1, child2)
	opt(parent)

	assert.Len(t, parent.Children, 2)
	assert.Equal(t, child1, parent.Children[0])
	assert.Equal(t, child2, parent.Children[1])
}

// TestWithChildren_Appends tests that WithChildren appends to existing children.
func TestWithChildren_Appends(t *testing.T) {
	child1 := Child("/profile", WithName("profile"))
	child2 := Child("/settings", WithName("settings"))
	child3 := Child("/posts", WithName("posts"))

	parent := &RouteRecord{
		Path:     "/user/:id",
		Name:     "user",
		Children: []*RouteRecord{child1},
	}

	// Apply WithChildren to append more children
	opt := WithChildren(child2, child3)
	opt(parent)

	assert.Len(t, parent.Children, 3)
	assert.Equal(t, child1, parent.Children[0])
	assert.Equal(t, child2, parent.Children[1])
	assert.Equal(t, child3, parent.Children[2])
}

// TestNestedRoutes_EmptyChildPath tests default child route (empty path).
func TestNestedRoutes_EmptyChildPath(t *testing.T) {
	// Default child route with empty path
	defaultChild := Child("", WithName("user-home"))
	parentRoute := &RouteRecord{
		Path:     "/user/:id",
		Name:     "user",
		Children: []*RouteRecord{defaultChild},
	}
	defaultChild.Parent = parentRoute

	matcher := NewRouteMatcher()
	err := matcher.AddRouteRecord(parentRoute)
	require.NoError(t, err)

	// Match parent path - should match both parent and default child
	match, err := matcher.Match("/user/123")
	require.NoError(t, err)

	assert.Len(t, match.Matched, 2)
	assert.Equal(t, "user", match.Matched[0].Name)
	assert.Equal(t, "user-home", match.Matched[1].Name)
}

// TestNestedRoutes_MultiLevel tests deeply nested routes (3+ levels).
func TestNestedRoutes_MultiLevel(t *testing.T) {
	// Create 3-level nested structure
	grandchild := Child("/edit", WithName("post-edit"))
	child := Child("/:postId", WithName("post-detail"))
	child.Children = []*RouteRecord{grandchild}
	grandchild.Parent = child

	parent := &RouteRecord{
		Path:     "/user/:userId",
		Name:     "user",
		Children: []*RouteRecord{child},
	}
	child.Parent = parent

	matcher := NewRouteMatcher()
	err := matcher.AddRouteRecord(parent)
	require.NoError(t, err)

	// Match deeply nested path
	match, err := matcher.Match("/user/123/456/edit")
	require.NoError(t, err)

	assert.Len(t, match.Matched, 3)
	assert.Equal(t, "user", match.Matched[0].Name)
	assert.Equal(t, "post-detail", match.Matched[1].Name)
	assert.Equal(t, "post-edit", match.Matched[2].Name)

	// Verify all params extracted
	assert.Equal(t, "123", match.Params["userId"])
	assert.Equal(t, "456", match.Params["postId"])
}
