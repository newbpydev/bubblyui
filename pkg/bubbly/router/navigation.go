package router

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// Message types are defined in messages.go

var (
	// ErrNilTarget is returned when navigation target is nil
	ErrNilTarget = errors.New("navigation target cannot be nil")

	// ErrEmptyTarget is returned when navigation target has no path or name
	ErrEmptyTarget = errors.New("navigation target must have path or name")
)

// Push navigates to a new route and adds it to history.
//
// Push generates a Bubbletea command that performs route matching,
// validation, and updates the current route. The command returns
// a RouteChangedMsg on success or NavigationErrorMsg on failure.
//
// Parameters:
//   - target: The navigation target (path, name, params, query, hash)
//
// Returns:
//   - tea.Cmd: Command that performs navigation and returns a message
//
// Navigation Flow:
//  1. Validate target (not nil, has path or name)
//  2. Match route from target path
//  3. Create Route object with params, query, hash
//  4. Update current route (with write lock)
//  5. Return RouteChangedMsg with from/to routes
//
// Error Handling:
//   - Nil target → NavigationErrorMsg with ErrNilTarget
//   - Empty target → NavigationErrorMsg with ErrEmptyTarget
//   - Route not found → NavigationErrorMsg with ErrNoMatch
//
// Thread Safety:
// The command executes asynchronously but updates router state with
// proper locking. Multiple Push() calls are serialized by Bubbletea.
//
// Example:
//
//	// Navigate by path
//	cmd := router.Push(&router.NavigationTarget{
//		Path: "/user/123",
//	})
//
//	// Navigate with query and hash
//	cmd := router.Push(&router.NavigationTarget{
//		Path:  "/search",
//		Query: map[string]string{"q": "golang"},
//		Hash:  "#results",
//	})
//
// Note: This is Task 2.2 - basic navigation only.
// Guard execution will be added in Task 2.3.
// History management will be added in Task 3.1.
func (r *Router) Push(target *NavigationTarget) tea.Cmd {
	return func() tea.Msg {
		// Use pushWithTracking for circular redirect detection
		return r.pushWithTracking(target, nil)
	}
}

// Replace navigates to a new route without adding to history.
//
// Replace is similar to Push() but doesn't create a history entry.
// This is useful for redirects, replacing the current route, or
// navigation that shouldn't be part of the back/forward stack.
//
// Parameters:
//   - target: The navigation target (path, name, params, query, hash)
//
// Returns:
//   - tea.Cmd: Command that performs navigation and returns a message
//
// Navigation Flow:
// Same as Push() but without history entry creation.
//
// Error Handling:
// Same as Push() - returns NavigationErrorMsg on failure.
//
// Thread Safety:
// Same as Push() - thread-safe with proper locking.
//
// Example:
//
//	// Replace current route (e.g., after login redirect)
//	cmd := router.Replace(&router.NavigationTarget{
//		Path: "/dashboard",
//	})
//
//	// Replace with query params
//	cmd := router.Replace(&router.NavigationTarget{
//		Path:  "/search",
//		Query: map[string]string{"q": "updated"},
//	})
//
// Use Cases:
//   - Login redirects (replace login page with dashboard)
//   - URL normalization (replace /users/ with /users)
//   - Error recovery (replace invalid route with valid one)
//   - Query parameter updates without history entry
//
// Note: This is Task 2.2 - basic navigation only.
// History management will be added in Task 3.1.
func (r *Router) Replace(target *NavigationTarget) tea.Cmd {
	return func() tea.Msg {
		// Use replaceWithTracking for circular redirect detection
		return r.replaceWithTracking(target, nil)
	}
}

// validateTarget checks if navigation target is valid.
//
// A valid target must:
//   - Not be nil
//   - Have at least a path or name
//
// Parameters:
//   - target: The navigation target to validate
//
// Returns:
//   - error: nil if valid, error describing the problem if invalid
//
// Errors:
//   - ErrNilTarget: target is nil
//   - ErrEmptyTarget: target has no path or name
func validateTarget(target *NavigationTarget) error {
	if target == nil {
		return ErrNilTarget
	}

	if target.Path == "" && target.Name == "" {
		return ErrEmptyTarget
	}

	return nil
}

// matchTarget matches a navigation target to a route.
//
// This method performs route matching based on the target's path,
// extracts parameters, and creates a complete Route object with
// params, query, and hash from the target.
//
// Parameters:
//   - target: The navigation target to match
//
// Returns:
//   - *Route: The matched route with params, query, hash
//   - error: ErrNoMatch if no route matches the path
//
// Matching Logic:
//  1. Use target.Path for matching (target.Name not yet supported)
//  2. Match against registered routes using matcher
//  3. Extract path parameters from match
//  4. Merge target.Query into route
//  5. Set target.Hash on route
//  6. Create Route object with all data
//
// Note: Named route navigation (target.Name) will be added in Task 4.5.
// For now, only path-based navigation is supported.
//
// Implementation Note:
// We need to sync registry routes to matcher. For now, we build a temporary
// matcher from registry routes. Task 2.5 (Router Builder) will handle this
// properly during router construction.
func (r *Router) matchTarget(target *NavigationTarget) (*Route, error) {
	// For now, only support path-based navigation
	// Named route navigation will be added in Task 4.5
	if target.Path == "" {
		return nil, fmt.Errorf("path-based navigation required (named routes not yet implemented)")
	}

	// Sync registry routes to matcher (temporary solution for Task 2.1/2.2)
	// Task 2.5 (Router Builder) will handle this properly
	// Skip sync if matcher already has routes (for testing with direct matcher manipulation)
	if len(r.matcher.routes) == 0 {
		r.syncRegistryToMatcher()
	}

	// Match route using matcher
	match, err := r.matcher.Match(target.Path)
	if err != nil {
		return nil, err
	}

	// Create Route object
	route := NewRoute(
		match.Route.Path, // Pattern path (e.g., "/user/:id")
		match.Route.Name, // Route name
		match.Params,     // Extracted params
		target.Query,     // Query params from target
		target.Hash,      // Hash from target
		match.Route.Meta, // Route metadata
		match.Matched,    // Matched chain (for nested routes, Task 4.1/4.2)
	)

	return route, nil
}

// syncRegistryToMatcher synchronizes routes from registry to matcher.
//
// This is a temporary helper for Task 2.1/2.2 to ensure routes registered
// in the registry are available in the matcher. Task 2.5 (Router Builder)
// will handle this properly during router construction.
//
// Thread Safety:
// This method should be called with appropriate locking if needed.
// For now, it's called within navigation commands which are serialized
// by Bubbletea.
func (r *Router) syncRegistryToMatcher() {
	// Get all routes from registry
	routes := r.registry.GetAll()

	// Clear matcher and re-add all routes
	// This is inefficient but works for Task 2.2
	// Task 2.5 will improve this
	r.matcher = NewRouteMatcher()
	for _, route := range routes {
		// Use AddRouteRecord to preserve Component field (Task 4.2/4.3)
		if err := r.matcher.AddRouteRecord(route); err != nil {
			// Report error to observability but continue - partial routing is better than no routing
			if reporter := observability.GetErrorReporter(); reporter != nil {
				ctx := &observability.ErrorContext{
					ComponentName: "Router",
					ComponentID:   "syncRegistryToMatcher",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"error_type": "add_route_failed",
						"route_path": route.Path,
					},
					Extra: map[string]interface{}{
						"route": route,
					},
				}
				reporter.ReportError(err, ctx)
			}
		}
	}
}
