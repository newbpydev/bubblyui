package router

import (
	"sync"
)

// NavigationGuard is a function that executes before navigation.
//
// Guards can inspect the target route (to), the current route (from),
// and control navigation flow via the next function. Guards are used
// for authentication, authorization, data fetching, and other pre-navigation logic.
//
// Parameters:
//   - to: The target route being navigated to
//   - from: The current route being navigated from (nil if no current route)
//   - next: Function to control navigation flow (allow, cancel, or redirect)
//
// Example:
//
//	func authGuard(to, from *Route, next NextFunc) {
//		if to.Meta["requiresAuth"] == true && !isAuthenticated() {
//			// Redirect to login
//			next(&NavigationTarget{Path: "/login"})
//		} else {
//			// Allow navigation
//			next(nil)
//		}
//	}
type NavigationGuard func(to, from *Route, next NextFunc)

// NextFunc controls navigation flow in guards.
//
// Guards call next() to allow, cancel, or redirect navigation:
//   - next(nil): Allow navigation to proceed
//   - next(&NavigationTarget{}): Cancel navigation (empty target)
//   - next(&NavigationTarget{Path: "/other"}): Redirect to different route
//
// Example:
//
//	// Allow navigation
//	next(nil)
//
//	// Cancel navigation
//	next(&NavigationTarget{})
//
//	// Redirect to login
//	next(&NavigationTarget{Path: "/login", Query: map[string]string{"redirect": to.FullPath}})
type NextFunc func(target *NavigationTarget)

// AfterNavigationHook is a function that executes after navigation completes.
//
// After hooks are called after the route has changed and cannot affect
// navigation. They are useful for analytics, logging, focus management,
// and other post-navigation side effects.
//
// Parameters:
//   - to: The new current route
//   - from: The previous route (nil if no previous route)
//
// Example:
//
//	func analyticsHook(to, from *Route) {
//		analytics.TrackPageView(to.Path)
//		if to.Meta["title"] != nil {
//			setWindowTitle(to.Meta["title"].(string))
//		}
//	}
type AfterNavigationHook func(to, from *Route)

// NavigationTarget specifies where to navigate.
//
// A navigation target can be specified by path, name, or both.
// Parameters and query strings can be provided to build the final URL.
//
// Fields:
//   - Path: Direct path to navigate to (e.g., "/user/123")
//   - Name: Route name to navigate to (e.g., "user-detail")
//   - Params: Path parameters to inject (e.g., {"id": "123"})
//   - Query: Query string parameters (e.g., {"tab": "settings"})
//   - Hash: Hash fragment (e.g., "#section")
//
// Navigation Modes:
//   - By Path: Set Path directly, params extracted from path
//   - By Name: Set Name + Params, path built from route definition
//   - Mixed: Set both Path and Name for validation
//
// Example:
//
//	// Navigate by path
//	target := &NavigationTarget{Path: "/user/123"}
//
//	// Navigate by name with params
//	target := &NavigationTarget{
//		Name:   "user-detail",
//		Params: map[string]string{"id": "123"},
//		Query:  map[string]string{"tab": "profile"},
//	}
//
//	// Navigate with hash
//	target := &NavigationTarget{
//		Path: "/docs/guide",
//		Hash: "#installation",
//	}
//
// Note: This type will be used in Task 2.2 (Navigation Implementation).
type NavigationTarget struct {
	Path   string            // Direct path (e.g., "/user/123")
	Name   string            // Route name (e.g., "user-detail")
	Params map[string]string // Path parameters (e.g., {"id": "123"})
	Query  map[string]string // Query parameters (e.g., {"tab": "settings"})
	Hash   string            // Hash fragment (e.g., "#section")
}

// Router is the main router singleton that manages navigation state.
//
// The router maintains the current route, history stack, registered routes,
// and navigation guards. It provides thread-safe access to routing state
// and generates Bubbletea commands for navigation.
//
// Architecture:
//   - registry: Manages route registration and lookup
//   - matcher: Performs path matching and parameter extraction
//   - history: Maintains navigation history stack
//   - currentRoute: The active route (nil if no route)
//   - beforeHooks: Global guards executed before navigation
//   - afterHooks: Global hooks executed after navigation
//   - mu: RWMutex for thread-safe concurrent access
//
// Thread Safety:
// All public methods are thread-safe. Multiple goroutines can safely
// call CurrentRoute() concurrently, and navigation operations are
// serialized with appropriate locking.
//
// Usage:
//
//	router := NewRouter()
//
//	// Access current route (thread-safe)
//	route := router.CurrentRoute()
//	if route != nil {
//		fmt.Printf("Current path: %s\n", route.Path)
//	}
//
// Note: This is Task 2.1 - basic structure only.
// Navigation methods (Push, Replace) will be added in Task 2.2.
// Guard execution will be added in Task 2.3.
// History operations will be added in Task 3.1-3.3.
type Router struct {
	registry     *RouteRegistry        // Route registration and lookup
	matcher      *RouteMatcher         // Path matching engine
	history      *History              // Navigation history stack
	currentRoute *Route                // Current active route (nil if none)
	beforeHooks  []NavigationGuard     // Global before guards
	afterHooks   []AfterNavigationHook // Global after hooks
	mu           sync.RWMutex          // Protects currentRoute and hooks
}

// NewRouter creates a new router instance with initialized components.
//
// The router is created with:
//   - Empty route registry
//   - Empty route matcher
//   - Empty history stack
//   - No current route (nil)
//   - Empty hook arrays
//
// Returns:
//   - *Router: A new router instance ready for route registration
//
// Thread Safety:
// The returned router is safe for concurrent use across multiple goroutines.
//
// Example:
//
//	router := NewRouter()
//	fmt.Printf("Current route: %v\n", router.CurrentRoute()) // nil
//
// Note: This is a simple constructor for Task 2.1.
// Task 2.5 will add a RouterBuilder for fluent route configuration.
func NewRouter() *Router {
	return &Router{
		registry:     NewRouteRegistry(),
		matcher:      NewRouteMatcher(),
		history:      &History{},
		currentRoute: nil,
		beforeHooks:  make([]NavigationGuard, 0),
		afterHooks:   make([]AfterNavigationHook, 0),
	}
}

// CurrentRoute returns the current active route with thread-safe access.
//
// Returns:
//   - *Route: The current route, or nil if no route is active
//
// Thread Safety:
// This method acquires a read lock and is safe for concurrent use.
// Multiple goroutines can call CurrentRoute() simultaneously without
// blocking each other.
//
// Immutability:
// The returned Route struct is immutable by design (from Task 1.5).
// All maps and slices in Route are defensively copied during creation,
// so external modification is prevented.
//
// Example:
//
//	route := router.CurrentRoute()
//	if route == nil {
//		fmt.Println("No active route")
//	} else {
//		fmt.Printf("Current path: %s\n", route.Path)
//		fmt.Printf("Params: %v\n", route.Params)
//	}
//
// Use Cases:
//   - Display current path in UI
//   - Access route parameters in components
//   - Check current route in guards
//   - Conditional rendering based on route
func (r *Router) CurrentRoute() *Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentRoute
}

// GetHistoryEntries returns a copy of all history entries for testing.
//
// This method is primarily intended for testing utilities to inspect
// the router's navigation history. It returns a defensive copy to prevent
// external modification of the internal history state.
//
// Returns:
//   - []*HistoryEntry: Copy of all history entries
//
// Thread Safety:
// This method acquires the history mutex and is safe for concurrent use.
//
// Example:
//
//	entries := router.GetHistoryEntries()
//	for _, entry := range entries {
//		fmt.Printf("Path: %s\n", entry.Route.Path)
//	}
//
// Use Cases:
//   - Testing history management
//   - Debugging navigation flows
//   - History inspection in test utilities
func (r *Router) GetHistoryEntries() []*HistoryEntry {
	r.history.mu.Lock()
	defer r.history.mu.Unlock()

	// Return defensive copy
	entries := make([]*HistoryEntry, len(r.history.entries))
	copy(entries, r.history.entries)
	return entries
}
