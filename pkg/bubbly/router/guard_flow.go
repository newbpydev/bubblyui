package router

import (
	"errors"
	"fmt"
)

const (
	// maxRedirectDepth is the maximum number of redirects allowed in a single navigation.
	// This prevents infinite redirect loops and stack overflow.
	// Default is 10, which should be sufficient for most use cases.
	maxRedirectDepth = 10
)

var (
	// ErrCircularRedirect is returned when a circular redirect is detected
	ErrCircularRedirect = errors.New("circular redirect detected")

	// ErrMaxRedirectDepth is returned when max redirect depth is exceeded
	ErrMaxRedirectDepth = fmt.Errorf("max redirect depth (%d) exceeded", maxRedirectDepth)
)

// redirectTracker tracks navigation paths to detect circular redirects.
//
// It maintains a set of visited paths during a navigation chain.
// If a path is visited twice, it indicates a circular redirect.
type redirectTracker struct {
	visited map[string]bool
	depth   int
}

// newRedirectTracker creates a new redirect tracker.
func newRedirectTracker() *redirectTracker {
	return &redirectTracker{
		visited: make(map[string]bool),
		depth:   0,
	}
}

// visit marks a path as visited and checks for circular redirects.
//
// Parameters:
//   - path: The path being visited
//
// Returns:
//   - error: ErrCircularRedirect if path was already visited, nil otherwise
func (rt *redirectTracker) visit(path string) error {
	if rt.visited[path] {
		return ErrCircularRedirect
	}
	rt.visited[path] = true
	return nil
}

// incrementDepth increments the redirect depth and checks the limit.
//
// Returns:
//   - error: ErrMaxRedirectDepth if depth exceeds limit, nil otherwise
func (rt *redirectTracker) incrementDepth() error {
	rt.depth++
	if rt.depth > maxRedirectDepth {
		return ErrMaxRedirectDepth
	}
	return nil
}

// pushWithTracking performs navigation with redirect tracking.
//
// This is an internal method used by Push() to track redirects and
// detect circular redirects or excessive redirect depth.
//
// Parameters:
//   - target: The navigation target
//   - tracker: The redirect tracker (nil for initial navigation)
//
// Returns:
//   - NavigationMsg: RouteChangedMsg on success, NavigationErrorMsg on failure
func (r *Router) pushWithTracking(target *NavigationTarget, tracker *redirectTracker) NavigationMsg {
	// Create tracker for initial navigation
	if tracker == nil {
		tracker = newRedirectTracker()
	}

	// Validate target
	if err := validateTarget(target); err != nil {
		return NavigationErrorMsg{
			Error: err,
			From:  r.CurrentRoute(),
			To:    target,
		}
	}

	// Match route
	newRoute, err := r.matchTarget(target)
	if err != nil {
		return NavigationErrorMsg{
			Error: err,
			From:  r.CurrentRoute(),
			To:    target,
		}
	}

	// Check for circular redirect
	if err := tracker.visit(newRoute.Path); err != nil {
		return NavigationErrorMsg{
			Error: fmt.Errorf("%w: %s", err, newRoute.Path),
			From:  r.CurrentRoute(),
			To:    target,
		}
	}

	// Execute before guards
	oldRoute := r.CurrentRoute()
	guardResult := r.executeBeforeGuards(newRoute, oldRoute)

	// Handle guard result
	switch guardResult.action {
	case guardCancel:
		return NavigationErrorMsg{
			Error: ErrNavigationCancelled,
			From:  oldRoute,
			To:    target,
		}

	case guardRedirect:
		// Increment redirect depth
		if err := tracker.incrementDepth(); err != nil {
			return NavigationErrorMsg{
				Error: err,
				From:  oldRoute,
				To:    guardResult.target,
			}
		}

		// Redirect to different route (recursive with tracking)
		return r.pushWithTracking(guardResult.target, tracker)

	case guardContinue:
		// Continue with navigation
	}

	// Update current route (thread-safe)
	r.mu.Lock()
	r.currentRoute = newRoute
	r.mu.Unlock()

	// Execute after hooks
	r.executeAfterHooks(newRoute, oldRoute)

	// Return success message
	return RouteChangedMsg{
		To:   newRoute,
		From: oldRoute,
	}
}

// replaceWithTracking performs replace navigation with redirect tracking.
//
// This is an internal method used by Replace() to track redirects and
// detect circular redirects or excessive redirect depth.
//
// Parameters:
//   - target: The navigation target
//   - tracker: The redirect tracker (nil for initial navigation)
//
// Returns:
//   - NavigationMsg: RouteChangedMsg on success, NavigationErrorMsg on failure
func (r *Router) replaceWithTracking(target *NavigationTarget, tracker *redirectTracker) NavigationMsg {
	// Create tracker for initial navigation
	if tracker == nil {
		tracker = newRedirectTracker()
	}

	// Validate target
	if err := validateTarget(target); err != nil {
		return NavigationErrorMsg{
			Error: err,
			From:  r.CurrentRoute(),
			To:    target,
		}
	}

	// Match route
	newRoute, err := r.matchTarget(target)
	if err != nil {
		return NavigationErrorMsg{
			Error: err,
			From:  r.CurrentRoute(),
			To:    target,
		}
	}

	// Check for circular redirect
	if err := tracker.visit(newRoute.Path); err != nil {
		return NavigationErrorMsg{
			Error: fmt.Errorf("%w: %s", err, newRoute.Path),
			From:  r.CurrentRoute(),
			To:    target,
		}
	}

	// Execute before guards
	oldRoute := r.CurrentRoute()
	guardResult := r.executeBeforeGuards(newRoute, oldRoute)

	// Handle guard result
	switch guardResult.action {
	case guardCancel:
		return NavigationErrorMsg{
			Error: ErrNavigationCancelled,
			From:  oldRoute,
			To:    target,
		}

	case guardRedirect:
		// Increment redirect depth
		if err := tracker.incrementDepth(); err != nil {
			return NavigationErrorMsg{
				Error: err,
				From:  oldRoute,
				To:    guardResult.target,
			}
		}

		// Redirect to different route (recursive with tracking)
		return r.replaceWithTracking(guardResult.target, tracker)

	case guardContinue:
		// Continue with navigation
	}

	// Update current route (thread-safe)
	// Note: No history update - that's the difference from Push()
	r.mu.Lock()
	r.currentRoute = newRoute
	r.mu.Unlock()

	// Execute after hooks
	r.executeAfterHooks(newRoute, oldRoute)

	// Return success message
	return RouteChangedMsg{
		To:   newRoute,
		From: oldRoute,
	}
}

// NavigationMsg is a marker interface for navigation messages.
//
// This interface is implemented by RouteChangedMsg and NavigationErrorMsg
// to allow type-safe handling of navigation results.
type NavigationMsg interface {
	isNavigationMsg()
}

// Implement NavigationMsg interface
func (RouteChangedMsg) isNavigationMsg()    {}
func (NavigationErrorMsg) isNavigationMsg() {}
