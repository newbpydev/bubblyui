package router

import (
	"strings"
)

// Child creates a child route record with the given path and options.
//
// Child routes are used to create nested route hierarchies where routes
// can have sub-routes. This is useful for layouts with nested views,
// such as a dashboard with multiple sections.
//
// The child path is relative to its parent. For example, if the parent
// path is "/user/:id" and the child path is "/profile", the full path
// will be "/user/:id/profile".
//
// An empty child path ("") creates a default child route that matches
// the parent path exactly. This is useful for rendering a default view
// when only the parent route is matched.
//
// Parameters:
//   - path: The child route path (relative to parent)
//   - opts: Optional RouteOption functions for configuration
//
// Returns:
//   - *RouteRecord: A new child route record
//
// Example:
//
//	// Simple child route
//	profileRoute := Child("/profile", WithName("user-profile"))
//
//	// Child with metadata
//	settingsRoute := Child("/settings",
//		WithName("user-settings"),
//		WithMeta(map[string]interface{}{"requiresAuth": true}),
//	)
//
//	// Default child (empty path)
//	homeRoute := Child("", WithName("user-home"))
//
//	// Parent with children
//	userRoute := &RouteRecord{
//		Path:     "/user/:id",
//		Name:     "user",
//		Children: []*RouteRecord{profileRoute, settingsRoute, homeRoute},
//	}
func Child(path string, opts ...RouteOption) *RouteRecord {
	record := &RouteRecord{
		Path: path,
	}

	// Apply all options
	for _, opt := range opts {
		opt(record)
	}

	return record
}

// resolveNestedPath combines parent and child paths into a full path.
//
// This function handles path resolution for nested routes by concatenating
// the parent and child paths with proper slash handling.
//
// Rules:
//   - If child path is empty, returns parent path (default child)
//   - If child path starts with "/", concatenates directly
//   - Ensures no double slashes in the result
//   - Preserves trailing slashes from child path
//
// Parameters:
//   - parentPath: The parent route path
//   - childPath: The child route path (relative to parent)
//
// Returns:
//   - string: The full resolved path
//
// Examples:
//
//	resolveNestedPath("/user/:id", "/profile")  // "/user/:id/profile"
//	resolveNestedPath("/user/:id", "/:tab")     // "/user/:id/:tab"
//	resolveNestedPath("/user/:id", "")          // "/user/:id"
//	resolveNestedPath("/dashboard", "/settings") // "/dashboard/settings"
func resolveNestedPath(parentPath, childPath string) string {
	// Empty child path means default child (matches parent exactly)
	if childPath == "" {
		return parentPath
	}

	// Ensure parent doesn't end with slash (unless it's root "/")
	parent := strings.TrimSuffix(parentPath, "/")
	if parent == "" {
		parent = "/"
	}

	// Ensure child starts with slash
	child := childPath
	if !strings.HasPrefix(child, "/") {
		child = "/" + child
	}

	// Concatenate
	if parent == "/" {
		return child
	}
	return parent + child
}

// buildMatchedArray constructs the Matched array for a route match.
//
// The Matched array contains all route records from the root to the
// matched route, in order. This is essential for nested route rendering
// where each level needs to know its parent routes.
//
// For example, matching "/user/123/profile" might produce:
//   - Matched[0]: /user/:id (parent)
//   - Matched[1]: /user/:id/profile (child)
//
// Parameters:
//   - route: The matched route record
//
// Returns:
//   - []*RouteRecord: Array of route records from root to matched route
//
// Example:
//
//	matched := buildMatchedArray(childRoute)
//	// matched = [parentRoute, childRoute]
func buildMatchedArray(route *RouteRecord) []*RouteRecord {
	if route == nil {
		return nil
	}

	// Build array by walking up the parent chain
	var matched []*RouteRecord
	current := route

	for current != nil {
		// Prepend to maintain root-to-leaf order
		matched = append([]*RouteRecord{current}, matched...)
		current = current.Parent
	}

	return matched
}

// establishParentLinks sets up bidirectional parent-child relationships.
//
// This function recursively walks the route tree and sets the Parent
// field on all child routes. This is necessary for building the Matched
// array and for route navigation.
//
// Parameters:
//   - parent: The parent route record
//
// Example:
//
//	parent := &RouteRecord{
//		Path:     "/user/:id",
//		Children: []*RouteRecord{child1, child2},
//	}
//	establishParentLinks(parent)
//	// child1.Parent == parent
//	// child2.Parent == parent
func establishParentLinks(parent *RouteRecord) {
	if parent == nil || parent.Children == nil {
		return
	}

	for _, child := range parent.Children {
		child.Parent = parent
		// Recursively establish links for grandchildren
		establishParentLinks(child)
	}
}

// buildFullPath constructs the full path for a route by walking up the parent chain.
//
// This function builds the complete path from root to the given route by
// concatenating all ancestor paths. This is necessary for nested routes
// where each level only stores its relative path.
//
// Parameters:
//   - route: The route to build the full path for
//
// Returns:
//   - string: The full resolved path from root to this route
//
// Example:
//
//	// parent.Path = "/user/:id"
//	// child.Path = "/:tab"
//	// buildFullPath(child) returns "/user/:id/:tab"
func buildFullPath(route *RouteRecord) string {
	if route == nil {
		return ""
	}

	// Build path segments by walking up parent chain
	var paths []string
	current := route

	for current != nil {
		paths = append([]string{current.Path}, paths...)
		current = current.Parent
	}

	// Resolve all paths from root to leaf
	if len(paths) == 0 {
		return ""
	}

	fullPath := paths[0]
	for i := 1; i < len(paths); i++ {
		fullPath = resolveNestedPath(fullPath, paths[i])
	}

	return fullPath
}
