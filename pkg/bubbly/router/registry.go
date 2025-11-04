package router

import (
	"fmt"
	"sync"
)

// RouteRegistry manages route registration and lookup with thread-safe access.
//
// The registry maintains routes in multiple indexes for efficient lookup:
//   - routes: Ordered list of all registered routes
//   - byName: Map for O(1) lookup by route name
//   - byPath: Map for O(1) lookup by route path
//
// Thread Safety:
// All operations are protected by sync.RWMutex, allowing concurrent reads
// and exclusive writes. This is safe for use across multiple goroutines.
//
// Usage:
//
//	registry := router.NewRouteRegistry()
//
//	// Register routes
//	err := registry.Register("/users", "users-list", nil)
//	if err != nil {
//		return err
//	}
//
//	// Lookup by name
//	route, found := registry.GetByName("users-list")
//	if found {
//		fmt.Printf("Route path: %s\n", route.Path)
//	}
//
//	// Lookup by path
//	route, found = registry.GetByPath("/users")
//	if found {
//		fmt.Printf("Route name: %s\n", route.Name)
//	}
//
// Duplicate Detection:
// The registry prevents duplicate paths and names. Attempting to register
// a route with a duplicate path or name will return an error.
//
// Nested Routes:
// Routes can have children via the Children field. Child routes are not
// automatically registered; they must be registered explicitly if needed
// for independent access.
type RouteRegistry struct {
	routes []*RouteRecord          // Ordered list of all routes
	byName map[string]*RouteRecord // Name → RouteRecord for O(1) lookup
	byPath map[string]*RouteRecord // Path → RouteRecord for O(1) lookup
	mu     sync.RWMutex            // Protects all fields
}

// NewRouteRegistry creates a new route registry with empty indexes.
//
// The returned registry is ready to accept route registrations via Register().
// All maps are pre-allocated for optimal performance.
//
// Returns:
//   - *RouteRegistry: A new registry instance with empty route list and indexes
//
// Example:
//
//	registry := router.NewRouteRegistry()
//	fmt.Printf("Routes: %d\n", len(registry.GetAll())) // 0
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make([]*RouteRecord, 0),
		byName: make(map[string]*RouteRecord),
		byPath: make(map[string]*RouteRecord),
	}
}

// Register adds a new route to the registry with duplicate detection.
//
// The route is compiled into a RoutePattern and indexed by both name and path
// for efficient lookup. Duplicate paths or names are rejected with an error.
//
// Parameters:
//   - path: The route pattern (e.g., "/users/:id", "/docs/:path*")
//   - name: Human-readable identifier for the route (e.g., "user-detail")
//   - meta: Optional metadata map (can be nil)
//
// Returns:
//   - error: nil on success, error on duplicate path/name or invalid pattern
//
// Errors:
//   - Duplicate path: Another route already registered with same path
//   - Duplicate name: Another route already registered with same name
//   - Invalid pattern: Path compilation failed (see CompilePattern errors)
//
// Thread Safety:
// This method acquires a write lock and is safe for concurrent use.
//
// Example:
//
//	err := registry.Register("/users/:id", "user-detail", map[string]interface{}{
//		"requiresAuth": true,
//	})
//	if err != nil {
//		return fmt.Errorf("failed to register route: %w", err)
//	}
func (rr *RouteRegistry) Register(path, name string, meta map[string]interface{}) error {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	// Check for duplicate path
	if _, exists := rr.byPath[path]; exists {
		return fmt.Errorf("duplicate path: %s already registered", path)
	}

	// Check for duplicate name
	if _, exists := rr.byName[name]; exists {
		return fmt.Errorf("duplicate name: %s already registered", name)
	}

	// Compile pattern
	pattern, err := CompilePattern(path)
	if err != nil {
		return fmt.Errorf("failed to compile pattern: %w", err)
	}

	// Create route record
	route := &RouteRecord{
		Path:    path,
		Name:    name,
		Meta:    meta,
		pattern: pattern,
	}

	// Add to all indexes
	rr.routes = append(rr.routes, route)
	rr.byName[name] = route
	rr.byPath[path] = route

	return nil
}

// GetByName retrieves a route by its name with O(1) lookup.
//
// Parameters:
//   - name: The route name to look up (e.g., "user-detail")
//
// Returns:
//   - *RouteRecord: The route record if found, nil otherwise
//   - bool: true if route was found, false otherwise
//
// Thread Safety:
// This method acquires a read lock and is safe for concurrent use.
// Multiple goroutines can call GetByName simultaneously.
//
// Example:
//
//	route, found := registry.GetByName("user-detail")
//	if !found {
//		return fmt.Errorf("route not found")
//	}
//	fmt.Printf("Route path: %s\n", route.Path)
func (rr *RouteRegistry) GetByName(name string) (*RouteRecord, bool) {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	route, found := rr.byName[name]
	return route, found
}

// GetByPath retrieves a route by its path with O(1) lookup.
//
// Parameters:
//   - path: The route path to look up (e.g., "/users/:id")
//
// Returns:
//   - *RouteRecord: The route record if found, nil otherwise
//   - bool: true if route was found, false otherwise
//
// Thread Safety:
// This method acquires a read lock and is safe for concurrent use.
// Multiple goroutines can call GetByPath simultaneously.
//
// Example:
//
//	route, found := registry.GetByPath("/users/:id")
//	if !found {
//		return fmt.Errorf("route not found")
//	}
//	fmt.Printf("Route name: %s\n", route.Name)
func (rr *RouteRegistry) GetByPath(path string) (*RouteRecord, bool) {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	route, found := rr.byPath[path]
	return route, found
}

// GetAll returns a copy of all registered routes.
//
// Returns:
//   - []*RouteRecord: Slice containing all registered routes
//
// Thread Safety:
// This method acquires a read lock and returns a defensive copy of the
// routes slice. The returned slice is safe to modify without affecting
// the registry.
//
// Performance:
// O(n) where n is the number of registered routes. The slice is copied
// to prevent external modification of the internal routes list.
//
// Example:
//
//	routes := registry.GetAll()
//	fmt.Printf("Total routes: %d\n", len(routes))
//	for _, route := range routes {
//		fmt.Printf("  %s → %s\n", route.Name, route.Path)
//	}
func (rr *RouteRegistry) GetAll() []*RouteRecord {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	// Return defensive copy to prevent external modification
	routes := make([]*RouteRecord, len(rr.routes))
	copy(routes, rr.routes)
	return routes
}
