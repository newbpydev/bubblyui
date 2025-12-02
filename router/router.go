// Package router provides Vue Router-inspired navigation for BubblyUI.
//
// The router enables SPA-style navigation in terminal applications with
// named routes, path parameters, query strings, navigation guards,
// nested routes, and history management.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/router,
// providing a cleaner import path for users.
//
// # Core Features
//
//   - Route matching with path parameters (/user/:id)
//   - Named routes for programmatic navigation
//   - Query string handling
//   - Navigation guards (beforeEach, afterEach)
//   - Nested/child routes
//   - History navigation (back/forward)
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/router"
//
//	func main() {
//	    r := router.NewRouterBuilder().
//	        Route("/", router.WithComponent(HomeComponent), router.WithName("home")).
//	        Route("/user/:id", router.WithComponent(UserComponent), router.WithName("user")).
//	        Route("/settings",
//	            router.WithComponent(SettingsComponent),
//	            router.WithGuard(authGuard),
//	        ).
//	        Build()
//
//	    // Navigate programmatically
//	    r.Push(&router.NavigationTarget{Name: "user", Params: map[string]string{"id": "123"}})
//	}
package router

import "github.com/newbpydev/bubblyui/pkg/bubbly/router"

// =============================================================================
// Router Builder
// =============================================================================

// NewRouterBuilder creates a new router builder for fluent configuration.
var NewRouterBuilder = router.NewRouterBuilder

// Builder provides fluent API for router configuration.
type Builder = router.Builder

// =============================================================================
// Route Configuration Options
// =============================================================================

// RouteOption configures route behavior.
type RouteOption = router.RouteOption

// WithComponent sets the component for a route.
var WithComponent = router.WithComponent

// WithName sets the name for a route (for programmatic navigation).
var WithName = router.WithName

// WithGuard adds a navigation guard to a route.
var WithGuard = router.WithGuard

// WithMeta adds metadata to a route.
var WithMeta = router.WithMeta

// WithChildren adds nested child routes.
var WithChildren = router.WithChildren

// =============================================================================
// Core Types
// =============================================================================

// Router is the main router instance.
type Router = router.Router

// Route represents a matched route with all its information.
type Route = router.Route

// NewRoute creates a new route instance.
var NewRoute = router.NewRoute

// RouteRecord defines a route configuration.
type RouteRecord = router.RouteRecord

// RouteMatch represents a successful route match.
type RouteMatch = router.RouteMatch

// =============================================================================
// Navigation
// =============================================================================

// NavigationTarget specifies where to navigate (by path or name).
type NavigationTarget = router.NavigationTarget

// NavigationGuard is a function called before navigation.
type NavigationGuard = router.NavigationGuard

// NextFunc controls navigation flow in guards.
type NextFunc = router.NextFunc

// AfterNavigationHook is called after navigation completes.
type AfterNavigationHook = router.AfterNavigationHook

// =============================================================================
// History
// =============================================================================

// History manages navigation history for back/forward.
type History = router.History

// HistoryEntry represents a single history entry.
type HistoryEntry = router.HistoryEntry

// =============================================================================
// Route Matching
// =============================================================================

// RouteMatcher handles route pattern matching.
type RouteMatcher = router.RouteMatcher

// NewRouteMatcher creates a new route matcher.
var NewRouteMatcher = router.NewRouteMatcher

// RoutePattern represents a parsed route pattern.
type RoutePattern = router.RoutePattern

// =============================================================================
// Query Parsing
// =============================================================================

// QueryParser handles query string parsing.
type QueryParser = router.QueryParser

// NewQueryParser creates a new query parser.
var NewQueryParser = router.NewQueryParser

// =============================================================================
// Router View
// =============================================================================

// View renders the current route's component.
type View = router.View

// NewRouterView creates a router view for rendering routes.
var NewRouterView = router.NewRouterView

// =============================================================================
// Composables
// =============================================================================

// ProvideRouter provides a router to child components.
var ProvideRouter = router.ProvideRouter

// UseRoute returns a reactive reference to the current route.
var UseRoute = router.UseRoute

// =============================================================================
// Messages (Bubbletea Integration)
// =============================================================================

// NavigationMsg is the base interface for navigation messages.
type NavigationMsg = router.NavigationMsg

// RouteChangedMsg is sent when the route changes.
type RouteChangedMsg = router.RouteChangedMsg

// NavigationErrorMsg is sent when navigation fails.
type NavigationErrorMsg = router.NavigationErrorMsg

// =============================================================================
// Errors
// =============================================================================

// Error represents a router error with context.
type Error = router.Error

// ErrorCode identifies the type of router error.
type ErrorCode = router.ErrorCode

// Error code constants.
const (
	ErrCodeRouteNotFound    = router.ErrCodeRouteNotFound
	ErrCodeInvalidTarget    = router.ErrCodeInvalidTarget
	ErrCodeGuardRejected    = router.ErrCodeGuardRejected
	ErrCodeCircularRedirect = router.ErrCodeCircularRedirect
)

// NewRouteNotFoundError creates a route not found error.
var NewRouteNotFoundError = router.NewRouteNotFoundError

// NewInvalidTargetError creates an invalid target error.
var NewInvalidTargetError = router.NewInvalidTargetError

// NewGuardRejectedError creates a guard rejected error.
var NewGuardRejectedError = router.NewGuardRejectedError

// NewCircularRedirectError creates a circular redirect error.
var NewCircularRedirectError = router.NewCircularRedirectError

// Standard errors.
var (
	ErrEmptyPath          = router.ErrEmptyPath
	ErrNoMatch            = router.ErrNoMatch
	ErrNilTarget          = router.ErrNilTarget
	ErrNavigationCanceled = router.ErrNavigationCanceled
	ErrCircularRedirect   = router.ErrCircularRedirect
)

// =============================================================================
// Component Guards Interface
// =============================================================================

// ComponentGuards interface for components with navigation guards.
type ComponentGuards = router.ComponentGuards
