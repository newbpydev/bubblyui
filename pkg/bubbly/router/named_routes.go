package router

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// BuildPath constructs a full path from a route name and parameters.
//
// BuildPath looks up the route by name, then builds the complete path by
// injecting parameters into the route pattern and appending query strings.
// This is useful for generating URLs programmatically without hardcoding paths.
//
// Parameters:
//   - name: The route name to look up (e.g., "user-detail")
//   - params: Path parameters to inject (e.g., {"id": "123"})
//   - query: Query string parameters (e.g., {"tab": "profile"})
//
// Returns:
//   - string: The complete path with params and query (e.g., "/user/123?tab=profile")
//   - error: nil on success, error if route not found or params missing
//
// Path Building Logic:
//  1. Look up route by name in registry
//  2. For each segment in route pattern:
//     - Static segments: use as-is
//     - Required params (:id): inject from params map (error if missing)
//     - Optional params (:id?): inject if present, omit if not
//     - Wildcards (:path*): inject if present, omit if not
//  3. Append query string if provided
//
// Error Handling:
//   - Route not found: Returns error with route name
//   - Missing required param: Returns error with param name
//   - Invalid route pattern: Returns error with details
//
// Thread Safety:
// This method is thread-safe. It acquires a read lock on the registry
// to look up the route.
//
// Example:
//
//	// Static route
//	path, err := router.BuildPath("home", nil, nil)
//	// Result: "/"
//
//	// Route with params
//	path, err := router.BuildPath("user-detail", map[string]string{"id": "123"}, nil)
//	// Result: "/user/123"
//
//	// Route with params and query
//	path, err := router.BuildPath("user-detail",
//		map[string]string{"id": "123"},
//		map[string]string{"tab": "profile"},
//	)
//	// Result: "/user/123?tab=profile"
//
//	// Optional param omitted
//	path, err := router.BuildPath("profile", nil, nil)
//	// Result: "/profile" (if pattern is "/profile/:id?")
//
//	// Wildcard
//	path, err := router.BuildPath("docs",
//		map[string]string{"path": "guide/getting-started"},
//		nil,
//	)
//	// Result: "/docs/guide/getting-started"
func (r *Router) BuildPath(name string, params, query map[string]string) (string, error) {
	// Look up route by name
	route, found := r.registry.GetByName(name)
	if !found {
		return "", fmt.Errorf("route not found: %s", name)
	}

	// Build path from pattern
	path, err := r.buildPathFromPattern(route.pattern, params)
	if err != nil {
		return "", fmt.Errorf("failed to build path for route %s: %w", name, err)
	}

	// Append query string if provided
	if len(query) > 0 {
		parser := NewQueryParser()
		queryString := parser.Build(query)
		path = path + "?" + queryString
	}

	return path, nil
}

// buildPathFromPattern constructs a path from a route pattern and parameters.
//
// This is an internal helper that handles the actual path construction logic
// by iterating through pattern segments and injecting parameters.
//
// Parameters:
//   - pattern: The compiled route pattern with segments
//   - params: Map of parameter values to inject
//
// Returns:
//   - string: The constructed path
//   - error: nil on success, error if required params are missing
func (r *Router) buildPathFromPattern(pattern *RoutePattern, params map[string]string) (string, error) {
	if pattern == nil {
		return "", fmt.Errorf("pattern cannot be nil")
	}

	// Initialize params map if nil
	if params == nil {
		params = make(map[string]string)
	}

	// Build path segments
	var segments []string

	for _, segment := range pattern.segments {
		switch segment.Kind {
		case SegmentStatic:
			// Static segment: use as-is
			segments = append(segments, segment.Value)

		case SegmentParam:
			// Required parameter: must be present
			value, ok := params[segment.Name]
			if !ok {
				return "", fmt.Errorf("missing required parameter: %s", segment.Name)
			}
			segments = append(segments, value)

		case SegmentOptional:
			// Optional parameter: include if present
			if value, ok := params[segment.Name]; ok {
				segments = append(segments, value)
			}
			// If not present, omit the segment

		case SegmentWildcard:
			// Wildcard parameter: include if present
			if value, ok := params[segment.Name]; ok {
				// Wildcard can contain slashes, so append as-is
				segments = append(segments, value)
			}
			// If not present, omit the segment
		}
	}

	// Join segments with "/"
	path := "/" + strings.Join(segments, "/")

	// Normalize path (remove double slashes, trailing slashes)
	path = normalizePath(path)

	return path, nil
}

// normalizePath cleans up a path by removing double slashes and trailing slashes.
//
// This ensures consistent path formatting regardless of how segments are joined.
//
// Parameters:
//   - path: The path to normalize
//
// Returns:
//   - string: The normalized path
func normalizePath(path string) string {
	// Remove double slashes
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	// Remove trailing slash (except for root "/")
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	// Ensure path starts with "/"
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}

// PushNamed navigates to a named route with parameters and query string.
//
// PushNamed is a convenience method that combines BuildPath() and Push()
// to navigate by route name instead of path. This is useful for avoiding
// hardcoded paths and ensuring type-safe navigation.
//
// Parameters:
//   - name: The route name to navigate to (e.g., "user-detail")
//   - params: Path parameters to inject (e.g., {"id": "123"})
//   - query: Query string parameters (e.g., {"tab": "profile"})
//
// Returns:
//   - tea.Cmd: Command that performs navigation, or nil on error
//
// Navigation Flow:
//  1. Build path from route name and params using BuildPath()
//  2. Create NavigationTarget with built path and query
//  3. Call Push() with the target
//  4. Return the navigation command
//
// Error Handling:
//   - Route not found: Returns nil (no navigation)
//   - Missing required params: Returns nil (no navigation)
//   - Other errors: Returns nil (no navigation)
//
// Note: Errors are not propagated as tea.Cmd. Instead, the method returns
// nil to indicate navigation failure. Applications should validate route
// names and parameters before calling PushNamed() if error handling is needed.
//
// Thread Safety:
// This method is thread-safe. It uses BuildPath() which acquires appropriate
// locks, and Push() which is also thread-safe.
//
// Example:
//
//	// Navigate to named route
//	cmd := router.PushNamed("home", nil, nil)
//
//	// Navigate with params
//	cmd := router.PushNamed("user-detail", map[string]string{"id": "123"}, nil)
//
//	// Navigate with params and query
//	cmd := router.PushNamed("user-detail",
//		map[string]string{"id": "123"},
//		map[string]string{"tab": "profile"},
//	)
//
//	// Handle in Update()
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//		switch msg := msg.(type) {
//		case tea.KeyMsg:
//			if msg.String() == "u" {
//				return m, m.router.PushNamed("user-detail",
//					map[string]string{"id": "123"},
//					nil,
//				)
//			}
//		}
//		return m, nil
//	}
func (r *Router) PushNamed(name string, params, query map[string]string) tea.Cmd {
	// Build path from route name and params
	path, err := r.BuildPath(name, params, query)
	if err != nil {
		// Return nil on error (no navigation)
		// Applications should validate route names and params beforehand
		return nil
	}

	// Create navigation target with built path
	target := &NavigationTarget{
		Path:  path,
		Name:  name, // Include name for validation/debugging
		Query: query,
	}

	// Use existing Push() method
	return r.Push(target)
}
