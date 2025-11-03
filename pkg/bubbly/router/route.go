package router

import (
	"strings"
)

// Route represents the current active route in the application.
//
// A Route contains all information about the current navigation state,
// including the path, extracted parameters, query string, hash fragment,
// metadata, and the chain of matched routes for nested routing.
//
// Immutability:
// Route instances are immutable. All maps and slices are defensively copied
// during construction to prevent external modification. This ensures that
// route state remains consistent throughout the navigation lifecycle.
//
// Fields:
//   - Path: The matched route path pattern (e.g., "/user/:id")
//   - Name: The route name for programmatic navigation
//   - Params: Extracted path parameters (e.g., {"id": "123"})
//   - Query: Parsed query string parameters (e.g., {"page": "1"})
//   - Hash: URL hash fragment without the "#" (e.g., "section-1")
//   - Meta: Route metadata for custom data (e.g., {"requiresAuth": true})
//   - Matched: Chain of matched routes for nested routing
//   - FullPath: Complete path including query and hash (e.g., "/user/123?tab=profile#bio")
//
// Usage:
//
//	route := router.NewRoute(
//		"/user/:id",
//		"user-detail",
//		map[string]string{"id": "123"},
//		map[string]string{"tab": "profile"},
//		"bio",
//		map[string]interface{}{"requiresAuth": true},
//		[]*RouteRecord{parentRoute, childRoute},
//	)
//
//	fmt.Printf("Path: %s\n", route.Path)           // "/user/:id"
//	fmt.Printf("User ID: %s\n", route.Params["id"]) // "123"
//	fmt.Printf("Full Path: %s\n", route.FullPath)   // "/user/:id?tab=profile#bio"
//
// Thread Safety:
// Route instances are safe for concurrent reads since they are immutable.
// All internal maps and slices are copied during construction.
type Route struct {
	Path     string                 // Route path pattern
	Name     string                 // Route name
	Params   map[string]string      // Path parameters
	Query    map[string]string      // Query parameters
	Hash     string                 // Hash fragment (without #)
	Meta     map[string]interface{} // Route metadata
	Matched  []*RouteRecord         // Matched route chain (for nested routes)
	FullPath string                 // Complete path with query and hash
}

// NewRoute creates a new immutable Route instance with defensive copies.
//
// All map and slice parameters are defensively copied to ensure immutability.
// Nil maps are converted to empty maps, and nil slices to empty slices.
// The FullPath is automatically generated from path, query, and hash.
//
// Parameters:
//   - path: The route path pattern (e.g., "/user/:id")
//   - name: The route name (e.g., "user-detail")
//   - params: Path parameters extracted from the URL
//   - query: Query string parameters
//   - hash: Hash fragment (without the "#" prefix)
//   - meta: Route metadata
//   - matched: Chain of matched routes for nested routing
//
// Returns:
//   - *Route: A new immutable Route instance
//
// Examples:
//
//	// Simple route
//	route := NewRoute("/users", "users-list", nil, nil, "", nil, nil)
//
//	// Route with parameters
//	route := NewRoute(
//		"/user/:id",
//		"user-detail",
//		map[string]string{"id": "123"},
//		nil, "", nil, nil,
//	)
//
//	// Route with query and hash
//	route := NewRoute(
//		"/docs",
//		"docs",
//		nil,
//		map[string]string{"version": "1.0"},
//		"api",
//		nil, nil,
//	)
//
//	// Nested route with matched chain
//	route := NewRoute(
//		"/dashboard/stats",
//		"dashboard-stats",
//		nil, nil, "",
//		nil,
//		[]*RouteRecord{dashboardRoute, statsRoute},
//	)
func NewRoute(
	path string,
	name string,
	params map[string]string,
	query map[string]string,
	hash string,
	meta map[string]interface{},
	matched []*RouteRecord,
) *Route {
	route := &Route{
		Path:    path,
		Name:    name,
		Hash:    hash,
		Params:  copyStringMap(params),
		Query:   copyStringMap(query),
		Meta:    copyInterfaceMap(meta),
		Matched: copyRouteRecords(matched),
	}

	// Generate FullPath
	route.FullPath = generateFullPath(path, query, hash)

	return route
}

// GetMeta retrieves a metadata value by key.
//
// This is a convenience method for accessing route metadata with
// existence checking. It returns both the value and a boolean indicating
// whether the key exists.
//
// Parameters:
//   - key: The metadata key to look up
//
// Returns:
//   - interface{}: The metadata value (nil if not found)
//   - bool: true if the key exists, false otherwise
//
// Examples:
//
//	requiresAuth, ok := route.GetMeta("requiresAuth")
//	if ok && requiresAuth.(bool) {
//		// Route requires authentication
//	}
//
//	title, ok := route.GetMeta("title")
//	if ok {
//		fmt.Printf("Page title: %s\n", title.(string))
//	}
func (r *Route) GetMeta(key string) (interface{}, bool) {
	value, found := r.Meta[key]
	return value, found
}

// copyStringMap creates a defensive copy of a string map.
//
// Returns an empty map if the input is nil, ensuring that Route
// fields are never nil for easier usage.
func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return make(map[string]string)
	}

	copy := make(map[string]string, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// copyInterfaceMap creates a defensive copy of an interface{} map.
//
// Returns an empty map if the input is nil, ensuring that Route
// fields are never nil for easier usage.
func copyInterfaceMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return make(map[string]interface{})
	}

	copy := make(map[string]interface{}, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// copyRouteRecords creates a defensive copy of a RouteRecord slice.
//
// This performs a shallow copy of the slice itself. The RouteRecord
// pointers are copied, but the RouteRecord structs themselves are shared.
// This is intentional as RouteRecords are typically managed by the router
// registry and should not be modified after creation.
//
// Returns an empty slice if the input is nil, ensuring that Route
// fields are never nil for easier usage.
func copyRouteRecords(records []*RouteRecord) []*RouteRecord {
	if records == nil {
		return make([]*RouteRecord, 0)
	}

	// Shallow copy of the slice - pointers are copied but structs are shared
	// This prevents the slice itself from being modified, but the RouteRecords
	// are still shared references (which is correct behavior)
	copy := make([]*RouteRecord, len(records))
	copy = append([]*RouteRecord{}, records...)
	return copy
}

// generateFullPath constructs the complete path including query and hash.
//
// The format is: path?query#hash
// - Query parameters are sorted alphabetically for deterministic output
// - Empty components are omitted (no trailing ? or #)
//
// Examples:
//   - "/users", nil, "" → "/users"
//   - "/search", {"q": "golang"}, "" → "/search?q=golang"
//   - "/docs", {"v": "1.0"}, "api" → "/docs?v=1.0#api"
//   - "/page", nil, "section" → "/page#section"
func generateFullPath(path string, query map[string]string, hash string) string {
	var builder strings.Builder
	builder.WriteString(path)

	// Add query string if present
	if len(query) > 0 {
		parser := NewQueryParser()
		queryString := parser.Build(query)
		if queryString != "" {
			builder.WriteString("?")
			builder.WriteString(queryString)
		}
	}

	// Add hash if present
	if hash != "" {
		builder.WriteString("#")
		builder.WriteString(hash)
	}

	return builder.String()
}
