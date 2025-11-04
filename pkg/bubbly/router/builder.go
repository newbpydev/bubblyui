package router

import (
	"errors"
	"fmt"
)

// Validation errors returned by Build().
var (
	// ErrEmptyPath is returned when a route has an empty path
	ErrEmptyPath = errors.New("path cannot be empty")

	// ErrDuplicatePath is returned when multiple routes have the same path
	ErrDuplicatePath = errors.New("duplicate path")

	// ErrDuplicateName is returned when multiple routes have the same name
	ErrDuplicateName = errors.New("duplicate name")
)

// Note: RouteRecord is defined in matcher.go and reused here for the builder.
// It contains Path, Name, Meta, and Children fields.

// RouterBuilder provides a fluent API for creating and configuring routers.
//
// The builder pattern makes router configuration readable and type-safe.
// It allows you to chain method calls to configure routes, guards, and hooks
// before building the final router instance.
//
// Architecture:
//   - routes: List of route records to register
//   - beforeHooks: Global before guards to register
//   - afterHooks: Global after hooks to register
//
// Example:
//
//	router, err := NewRouterBuilder().
//		Route("/", "home").
//		Route("/about", "about").
//		RouteWithMeta("/dashboard", "dashboard", map[string]interface{}{
//			"requiresAuth": true,
//		}).
//		BeforeEach(authGuard).
//		AfterEach(analyticsHook).
//		Build()
//
// Thread Safety:
// The builder is NOT thread-safe. It should be used in a single goroutine
// during router setup. The built router IS thread-safe.
type RouterBuilder struct {
	// routes is the list of route records to register
	routes []*RouteRecord

	// beforeHooks is the list of global before guards
	beforeHooks []NavigationGuard

	// afterHooks is the list of global after hooks
	afterHooks []AfterNavigationHook
}

// NewRouterBuilder creates a new RouterBuilder for building a router.
//
// This is the entry point for creating routers using the fluent API.
// The builder starts empty and routes/guards are added via method chaining.
//
// Example:
//
//	builder := NewRouterBuilder()
//	builder.Route("/home", "home").Route("/about", "about")
//
// Returns:
//   - *RouterBuilder: A builder instance ready for configuration
func NewRouterBuilder() *RouterBuilder {
	return &RouterBuilder{
		routes:      make([]*RouteRecord, 0),
		beforeHooks: make([]NavigationGuard, 0),
		afterHooks:  make([]AfterNavigationHook, 0),
	}
}

// Route adds a route to the builder.
//
// This is the primary method for registering routes. It creates a route
// with the given path and name, with no metadata.
//
// Parameters:
//   - path: The route path pattern (e.g., "/users/:id")
//   - name: The route name for named navigation
//
// Returns:
//   - *RouterBuilder: The builder for method chaining
//
// Example:
//
//	builder.Route("/home", "home").
//		Route("/about", "about").
//		Route("/users/:id", "user-detail")
func (rb *RouterBuilder) Route(path, name string) *RouterBuilder {
	rb.routes = append(rb.routes, &RouteRecord{
		Path: path,
		Name: name,
		Meta: nil,
	})
	return rb
}

// RouteWithMeta adds a route with metadata to the builder.
//
// This method is used when you need to attach metadata to a route,
// such as authentication requirements, page titles, or other custom data.
//
// Parameters:
//   - path: The route path pattern
//   - name: The route name
//   - meta: Metadata map
//
// Returns:
//   - *RouterBuilder: The builder for method chaining
//
// Example:
//
//	builder.RouteWithMeta("/dashboard", "dashboard", map[string]interface{}{
//		"requiresAuth": true,
//		"title": "Dashboard",
//		"roles": []string{"admin", "user"},
//	})
func (rb *RouterBuilder) RouteWithMeta(path, name string, meta map[string]interface{}) *RouterBuilder {
	rb.routes = append(rb.routes, &RouteRecord{
		Path: path,
		Name: name,
		Meta: meta,
	})
	return rb
}

// RouteWithOptions adds a route with functional options to the builder.
//
// This method uses the functional options pattern for flexible route
// configuration. Options can be combined to configure different aspects
// of the route (name, metadata, guards, children).
//
// Parameters:
//   - path: The route path pattern
//   - opts: Variable number of RouteOption functions
//
// Returns:
//   - *RouterBuilder: The builder for method chaining
//
// Example:
//
//	builder.RouteWithOptions("/dashboard",
//		WithName("dashboard"),
//		WithMeta(map[string]interface{}{
//			"requiresAuth": true,
//		}),
//		WithGuard(authGuard),
//		WithChildren(overviewRoute, settingsRoute),
//	)
func (rb *RouterBuilder) RouteWithOptions(path string, opts ...RouteOption) *RouterBuilder {
	record := &RouteRecord{
		Path: path,
	}

	// Apply all options
	for _, opt := range opts {
		opt(record)
	}

	rb.routes = append(rb.routes, record)
	return rb
}

// BeforeEach registers a global before guard.
//
// Before guards execute before every navigation and can inspect the target
// route, current route, and control navigation flow via the next() function.
//
// Parameters:
//   - guard: The guard function to register
//
// Returns:
//   - *RouterBuilder: The builder for method chaining
//
// Example:
//
//	builder.BeforeEach(func(to, from *Route, next NextFunc) {
//		if to.Meta["requiresAuth"] == true && !isAuthenticated() {
//			next(&NavigationTarget{Path: "/login"})
//		} else {
//			next(nil)
//		}
//	})
func (rb *RouterBuilder) BeforeEach(guard NavigationGuard) *RouterBuilder {
	rb.beforeHooks = append(rb.beforeHooks, guard)
	return rb
}

// AfterEach registers a global after hook.
//
// After hooks execute after navigation completes successfully. They cannot
// affect navigation but are useful for side effects like analytics, logging,
// focus management, and other post-navigation tasks.
//
// Parameters:
//   - hook: The hook function to register
//
// Returns:
//   - *RouterBuilder: The builder for method chaining
//
// Example:
//
//	builder.AfterEach(func(to, from *Route) {
//		analytics.TrackPageView(to.Path)
//		log.Printf("Navigated to %s", to.Path)
//	})
func (rb *RouterBuilder) AfterEach(hook AfterNavigationHook) *RouterBuilder {
	rb.afterHooks = append(rb.afterHooks, hook)
	return rb
}

// Build creates and configures a router from the builder.
//
// This method:
//  1. Validates the route configuration
//  2. Creates a new router instance
//  3. Registers all routes
//  4. Registers all guards and hooks
//  5. Returns the configured router
//
// Validation:
//   - Checks for empty paths
//   - Checks for duplicate paths
//   - Checks for duplicate names (if name is not empty)
//
// Returns:
//   - *Router: The configured router instance
//   - error: Validation error if configuration is invalid
//
// Example:
//
//	router, err := builder.Build()
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Use router...
func (rb *RouterBuilder) Build() (*Router, error) {
	// Validate routes
	if err := rb.validate(); err != nil {
		return nil, err
	}

	// Create new router
	router := NewRouter()

	// Register routes with full RouteRecord (including Component)
	for _, record := range rb.routes {
		// Compile pattern for the route
		pattern, err := CompilePattern(record.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to compile pattern for %s: %w", record.Path, err)
		}
		record.pattern = pattern
		
		// Add route record directly to registry (preserving Component field)
		router.registry.mu.Lock()
		router.registry.routes = append(router.registry.routes, record)
		router.registry.byName[record.Name] = record
		router.registry.byPath[record.Path] = record
		router.registry.mu.Unlock()
	}

	// Register guards
	for _, guard := range rb.beforeHooks {
		router.BeforeEach(guard)
	}

	// Register hooks
	for _, hook := range rb.afterHooks {
		router.AfterEach(hook)
	}

	return router, nil
}

// validate checks the route configuration for errors.
//
// Validation rules:
//   - Path cannot be empty
//   - Paths must be unique
//   - Names must be unique (if provided)
//
// Returns:
//   - error: Validation error if configuration is invalid, nil otherwise
func (rb *RouterBuilder) validate() error {
	// Track seen paths and names
	seenPaths := make(map[string]bool)
	seenNames := make(map[string]bool)

	for _, record := range rb.routes {
		// Check for empty path
		if record.Path == "" {
			return ErrEmptyPath
		}

		// Check for duplicate path
		if seenPaths[record.Path] {
			return fmt.Errorf("%w: %s", ErrDuplicatePath, record.Path)
		}
		seenPaths[record.Path] = true

		// Check for duplicate name (only if name is not empty)
		if record.Name != "" {
			if seenNames[record.Name] {
				return fmt.Errorf("%w: %s", ErrDuplicateName, record.Name)
			}
			seenNames[record.Name] = true
		}
	}

	return nil
}
