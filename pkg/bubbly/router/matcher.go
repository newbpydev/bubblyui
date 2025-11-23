// Package router provides path matching and routing functionality for BubblyUI applications.
//
// The router package implements a high-performance path matching system inspired by Vue Router,
// designed specifically for terminal user interface (TUI) applications. It supports static routes,
// dynamic parameters, optional parameters, and wildcards with intelligent route precedence.
//
// Key Features:
//   - Static route matching (/users, /about)
//   - Dynamic parameter extraction (/user/:id)
//   - Optional parameters (/profile/:id?)
//   - Wildcard matching (/docs/:path*)
//   - Route specificity scoring (most specific wins)
//   - Sub-microsecond performance (< 2μs per match)
//
// Basic Usage:
//
//	matcher := router.NewRouteMatcher()
//	err := matcher.AddRoute("/users/:id", "user-detail")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	match, err := matcher.Match("/users/123")
//	if err != nil {
//		if errors.Is(err, router.ErrNoMatch) {
//			// Handle 404
//		}
//		return err
//	}
//
//	fmt.Printf("Route: %s, User ID: %s\n", match.Route.Name, match.Params["id"])
//
// Route Precedence:
// Routes are scored by specificity, with more specific routes taking precedence:
//  1. More static segments = higher priority
//  2. Fewer parameter segments = higher priority
//  3. Fewer optional segments = higher priority
//  4. Fewer wildcard segments = higher priority
//
// Example precedence:
//
//	/users/new      (2 static)      → beats /users/:id
//	/users/:id      (1 static, 1 param) → beats /:resource/:id
//	/docs/:path*    (1 static, 1 wildcard) → lowest priority
//
// Performance:
// The matcher is optimized for speed with compiled patterns and efficient scoring.
// Benchmarks show ~1μs per match with minimal memory allocation.
package router

import (
	"errors"
	"fmt"
	"sort"
)

var (
	// ErrNoMatch is returned when no route matches the given path.
	// This error should be used to handle 404 scenarios in TUI applications.
	//
	// Example:
	//	match, err := matcher.Match("/unknown/path")
	//	if errors.Is(err, router.ErrNoMatch) {
	//		// Display 404 message or navigate to error route
	//	}
	ErrNoMatch = errors.New("no route matches path")
)

// RouteRecord represents a registered route in the matcher.
//
// Each route contains the original path pattern, a human-readable name,
// and the compiled pattern used for efficient matching. The pattern
// is compiled once during route registration for optimal performance.
//
// Fields:
//   - Path: The original route pattern (e.g., "/users/:id")
//   - Name: Human-readable identifier for the route (e.g., "user-detail")
//   - Component: The component to render for this route (optional, for View)
//   - Meta: Optional metadata map for route-specific data (e.g., auth requirements)
//   - Parent: Reference to parent route for nested routes (nil for top-level routes)
//   - Children: Nested child routes for hierarchical routing
//   - pattern: Compiled pattern containing segments and regex (internal)
type RouteRecord struct {
	Path      string
	Name      string
	Component interface{}            // Component to render (bubbly.Component)
	Meta      map[string]interface{} // Optional metadata (e.g., requiresAuth, title)
	Parent    *RouteRecord           // Parent route for nested routes
	Children  []*RouteRecord         // Nested child routes
	pattern   *RoutePattern
}

// RouteMatcher manages a collection of routes and performs path matching.
//
// The matcher maintains an ordered list of registered routes and uses
// a two-phase matching algorithm: pattern matching followed by specificity
// scoring to determine the best match.
//
// Thread Safety:
// RouteMatcher is NOT thread-safe. If you need concurrent access, wrap
// operations with appropriate synchronization (mutex or channels).
//
// Performance Characteristics:
//   - Route registration: O(1) per route (pattern compilation dominates)
//   - Path matching: O(n) where n is number of registered routes
//   - Memory usage: O(n) for route storage
//
// Example:
//
//	matcher := router.NewRouteMatcher()
//
//	// Register routes
//	matcher.AddRoute("/", "home")
//	matcher.AddRoute("/users", "users-list")
//	matcher.AddRoute("/users/:id", "user-detail")
//	matcher.AddRoute("/docs/:path*", "documentation")
//
//	// Match paths
//	match, err := matcher.Match("/users/123")
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Matched route: %s\n", match.Route.Name)
//	fmt.Printf("User ID: %s\n", match.Params["id"])
type RouteMatcher struct {
	routes []*RouteRecord
}

// RouteMatch represents the result of a successful route match.
//
// When a path matches a registered route, RouteMatch contains the matched
// route record, extracted parameters, and the specificity score used for
// precedence determination.
//
// Fields:
//   - Route: The matched route record
//   - Params: Extracted path parameters (e.g., {"id": "123"})
//   - Score: Specificity score used for route precedence
//   - Matched: Array of matched route records from root to leaf (for nested routes)
//
// Example:
//
//	match, _ := matcher.Match("/user/123")
//	fmt.Printf("Route: %s\n", match.Route.Name)        // "user-detail"
//	fmt.Printf("User ID: %s\n", match.Params["id"])     // "123"
//	fmt.Printf("Score: %+v\n", match.Score)            // {static:1, param:1, ...}
//	fmt.Printf("Matched: %v\n", match.Matched)         // [parentRoute, childRoute]
type RouteMatch struct {
	Route   *RouteRecord
	Params  map[string]string
	Score   matchScore
	Matched []*RouteRecord // Route records from root to matched route
}

// matchScore represents the specificity of a route match for precedence calculation.
//
// Higher scores indicate more specific routes. The scoring algorithm prioritizes
// static segments over dynamic ones, with optional and wildcard segments having
// the lowest specificity.
//
// Scoring Rules:
//   - staticSegments: More static segments = more specific
//   - paramSegments: Fewer parameters = more specific
//   - optionalSegments: Fewer optionals = more specific
//   - wildcardSegments: Fewer wildcards = more specific
//
// Examples:
//
//	/users/new      → {static:2, param:0, optional:0, wildcard:0}    // Most specific
//	/users/:id      → {static:1, param:1, optional:0, wildcard:0}    // Medium
//	/profile/:id?   → {static:1, param:0, optional:1, wildcard:0}    // Less specific
//	/docs/:path*    → {static:1, param:0, optional:0, wildcard:1}    // Least specific
type matchScore struct {
	staticSegments   int
	paramSegments    int
	optionalSegments int
	wildcardSegments int
}

// NewRouteMatcher creates a new route matcher with an empty route registry.
//
// The returned matcher is ready to accept route registrations via AddRoute().
// No routes are pre-registered; you must add all routes explicitly.
//
// Returns:
//   - *RouteMatcher: A new matcher instance with empty route list
//
// Example:
//
//	matcher := router.NewRouteMatcher()
//	fmt.Printf("Routes: %d\n", len(matcher.routes)) // 0
func NewRouteMatcher() *RouteMatcher {
	return &RouteMatcher{
		routes: make([]*RouteRecord, 0),
	}
}

// AddRoute registers a new route with the matcher for subsequent matching.
//
// The path is compiled into a RoutePattern during registration for optimal
// matching performance. Invalid paths will return an error without modifying
// the matcher's state.
//
// Parameters:
//   - path: The route pattern to register (e.g., "/users/:id", "/docs/:path*")
//   - name: Human-readable identifier for the route (e.g., "user-detail")
//
// Returns:
//   - error: nil on success, compilation error on invalid path patterns
//
// Supported Pattern Types:
//   - Static: "/users", "/about/contact"
//   - Dynamic: "/user/:id", "/post/:slug"
//   - Optional: "/profile/:id?", "/search/:query?"
//   - Wildcard: "/docs/:path*", "/files/:name*"
//
// Errors:
//   - Empty path
//   - Path not starting with "/"
//   - Invalid parameter names
//   - Duplicate parameter names
//   - Wildcard not at end of path
//
// Example:
//
//	err := matcher.AddRoute("/users/:id", "user-detail")
//	if err != nil {
//		return fmt.Errorf("failed to add route: %w", err)
//	}
func (rm *RouteMatcher) AddRoute(path, name string) error {
	// Compile the pattern
	pattern, err := CompilePattern(path)
	if err != nil {
		return fmt.Errorf("failed to compile pattern: %w", err)
	}

	// Create route record
	route := &RouteRecord{
		Path:    path,
		Name:    name,
		pattern: pattern,
	}

	rm.routes = append(rm.routes, route)
	return nil
}

// AddRouteRecord registers a RouteRecord (potentially with children) with the matcher.
//
// This method is used for nested routes where a RouteRecord may have child routes.
// It recursively registers the parent route and all its children, establishing
// parent-child relationships and resolving nested paths.
//
// Parameters:
//   - record: The route record to register (may have children)
//
// Returns:
//   - error: nil on success, compilation error on invalid patterns
//
// Example:
//
//	parent := &RouteRecord{
//		Path: "/user/:id",
//		Name: "user",
//		Children: []*RouteRecord{
//			{Path: "/profile", Name: "user-profile"},
//			{Path: "/settings", Name: "user-settings"},
//		},
//	}
//	err := matcher.AddRouteRecord(parent)
func (rm *RouteMatcher) AddRouteRecord(record *RouteRecord) error {
	// Establish parent-child links
	establishParentLinks(record)

	// Register the parent route
	if err := rm.addSingleRoute(record); err != nil {
		return err
	}

	// Recursively register children
	if record.Children != nil {
		for _, child := range record.Children {
			if err := rm.registerNestedRoute(record, child); err != nil {
				return err
			}
		}
	}

	return nil
}

// addSingleRoute registers a single route without processing children.
func (rm *RouteMatcher) addSingleRoute(record *RouteRecord) error {
	// Compile the pattern
	pattern, err := CompilePattern(record.Path)
	if err != nil {
		return fmt.Errorf("failed to compile pattern %s: %w", record.Path, err)
	}

	record.pattern = pattern
	rm.routes = append(rm.routes, record)
	return nil
}

// registerNestedRoute registers a child route with its full resolved path.
func (rm *RouteMatcher) registerNestedRoute(_, child *RouteRecord) error {
	// Build full path by walking up parent chain
	// This handles deeply nested routes (grandchildren, etc.)
	fullPath := buildFullPath(child)

	// Compile pattern for full path
	pattern, err := CompilePattern(fullPath)
	if err != nil {
		return fmt.Errorf("failed to compile nested pattern %s: %w", fullPath, err)
	}

	child.pattern = pattern
	rm.routes = append(rm.routes, child)

	// Recursively register grandchildren
	if child.Children != nil {
		for _, grandchild := range child.Children {
			if err := rm.registerNestedRoute(child, grandchild); err != nil {
				return err
			}
		}
	}

	return nil
}

// Match finds the best matching route for the given path using intelligent precedence.
//
// The matching algorithm works in two phases:
//  1. Pattern matching: Tests each registered route against the path
//  2. Specificity scoring: Ranks matches by route specificity
//
// Parameters:
//   - path: The incoming path to match (e.g., "/users/123", "/docs/guide")
//
// Returns:
//   - *RouteMatch: The best matching route with extracted parameters and score
//   - error: ErrNoMatch if no routes match the path
//
// Path Normalization:
//   - Trailing slashes are normalized ("/users/" → "/users")
//   - Empty paths are treated as root ("" → "/")
//
// Parameter Extraction:
//   - Dynamic parameters: "/user/:id" → {"id": "123"}
//   - Optional parameters: "/profile/:id?" → {"id": "123"} or {}
//   - Wildcards: "/docs/:path*" → {"path": "guide/getting-started"}
//
// Route Precedence:
//
//	Routes are ranked by specificity (most specific wins):
//	  1. More static segments
//	  2. Fewer parameter segments
//	  3. Fewer optional segments
//	  4. Fewer wildcard segments
//
// Example:
//
//	match, err := matcher.Match("/users/123")
//	if err != nil {
//		if errors.Is(err, router.ErrNoMatch) {
//			// Handle 404 - no route matched
//		}
//		return err
//	}
//	fmt.Printf("Route: %s, User ID: %s\n", match.Route.Name, match.Params["id"])
func (rm *RouteMatcher) Match(path string) (*RouteMatch, error) {
	// Collect all matches
	var matches []*RouteMatch

	for _, route := range rm.routes {
		params, ok := route.pattern.Match(path)
		if !ok {
			continue
		}

		// Calculate score
		score := calculateScore(route.pattern.segments)

		// Build matched array (root to leaf for nested routes)
		matched := buildMatchedArray(route)

		matches = append(matches, &RouteMatch{
			Route:   route,
			Params:  params,
			Score:   score,
			Matched: matched,
		})
	}

	// No matches found
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoMatch, path)
	}

	// Sort by specificity (most specific first)
	sort.Slice(matches, func(i, j int) bool {
		// If scores are equal, prefer child routes over parent routes
		// This handles the case where a child has an empty path (same pattern as parent)
		if matches[i].Score == matches[j].Score {
			// Route with parent is a child route - prefer it
			return matches[i].Route.Parent != nil && matches[j].Route.Parent == nil
		}
		return isMoreSpecific(matches[i].Score, matches[j].Score)
	})

	// Return best match
	return matches[0], nil
}

// calculateScore computes the specificity score for a route based on its segments.
//
// The scoring algorithm analyzes each segment type to determine route specificity.
// More specific routes receive higher scores and win precedence in ambiguous matches.
//
// Parameters:
//   - segments: List of segments from the compiled route pattern
//
// Returns:
//   - matchScore: Specificity score with counts for each segment type
//
// Scoring Logic:
//   - staticSegments: Count of static path segments (higher = more specific)
//   - paramSegments: Count of required parameters (lower = more specific)
//   - optionalSegments: Count of optional parameters (lower = more specific)
//   - wildcardSegments: Count of wildcard segments (lower = more specific)
//
// Example Scoring:
//
//	"/users/new"      → {static:2, param:0, optional:0, wildcard:0}
//	"/users/:id"      → {static:1, param:1, optional:0, wildcard:0}
//	"/profile/:id?"   → {static:1, param:0, optional:1, wildcard:0}
//	"/docs/:path*"    → {static:1, param:0, optional:0, wildcard:1}
func calculateScore(segments []Segment) matchScore {
	score := matchScore{}

	for _, seg := range segments {
		switch seg.Kind {
		case SegmentStatic:
			score.staticSegments++
		case SegmentParam:
			score.paramSegments++
		case SegmentOptional:
			score.optionalSegments++
		case SegmentWildcard:
			score.wildcardSegments++
		}
	}

	return score
}

// isMoreSpecific compares two route scores to determine which is more specific.
//
// This function implements the route precedence algorithm, ensuring that
// more specific routes are preferred over generic ones when multiple routes
// could match the same path.
//
// Parameters:
//   - a: First route's specificity score
//   - b: Second route's specificity score
//
// Returns:
//   - bool: true if score 'a' is more specific than score 'b'
//
// Comparison Rules (in order):
//  1. More static segments wins
//  2. If static equal, fewer parameters wins
//  3. If params equal, fewer optionals wins
//  4. If optionals equal, fewer wildcards wins
//  5. If all equal, routes are considered equally specific
//
// Example Comparisons:
//
//	isMoreSpecific({static:2, param:0}, {static:1, param:1}) → true
//	isMoreSpecific({static:1, param:0}, {static:1, param:1}) → true
//	isMoreSpecific({static:1, param:1}, {static:1, param:1}) → false
func isMoreSpecific(a, b matchScore) bool {
	// More static segments = more specific
	if a.staticSegments != b.staticSegments {
		return a.staticSegments > b.staticSegments
	}

	// Fewer param segments = more specific
	if a.paramSegments != b.paramSegments {
		return a.paramSegments < b.paramSegments
	}

	// Fewer optional segments = more specific
	if a.optionalSegments != b.optionalSegments {
		return a.optionalSegments < b.optionalSegments
	}

	// Fewer wildcard segments = more specific
	return a.wildcardSegments < b.wildcardSegments
}
