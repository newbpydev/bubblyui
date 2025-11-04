package router

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestProvideRouter tests that ProvideRouter correctly provides router to component tree
func TestProvideRouter(t *testing.T) {
	tests := []struct {
		name          string
		setupParent   func(*bubbly.Context, *Router)
		setupChild    func(*bubbly.Context) *Router
		expectPanic   bool
		validateChild func(*testing.T, *Router, *Router)
	}{
		{
			name: "router provided and accessible in child",
			setupParent: func(ctx *bubbly.Context, router *Router) {
				ProvideRouter(ctx, router)
			},
			setupChild: func(ctx *bubbly.Context) *Router {
				return UseRouter(ctx)
			},
			expectPanic: false,
			validateChild: func(t *testing.T, expected, actual *Router) {
				assert.Equal(t, expected, actual, "Child should receive same router instance")
			},
		},
		{
			name: "router not provided causes panic",
			setupParent: func(ctx *bubbly.Context, router *Router) {
				// Don't provide router
			},
			setupChild: func(ctx *bubbly.Context) *Router {
				return UseRouter(ctx)
			},
			expectPanic: true,
			validateChild: func(t *testing.T, expected, actual *Router) {
				// Not called due to panic
			},
		},
		{
			name: "multiple children access same router",
			setupParent: func(ctx *bubbly.Context, router *Router) {
				ProvideRouter(ctx, router)
			},
			setupChild: func(ctx *bubbly.Context) *Router {
				// Call UseRouter twice to simulate multiple children
				router1 := UseRouter(ctx)
				router2 := UseRouter(ctx)
				assert.Equal(t, router1, router2, "Multiple calls should return same instance")
				return router1
			},
			expectPanic: false,
			validateChild: func(t *testing.T, expected, actual *Router) {
				assert.Equal(t, expected, actual, "All children should receive same router instance")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			router := NewRouter()

			// Create parent and child contexts
			var childRouter *Router
			parentCtx := bubbly.NewTestContext()
			childCtx := bubbly.NewTestContext()

			// Set up parent-child relationship
			bubbly.SetParent(childCtx, parentCtx)

			// Setup parent
			tt.setupParent(parentCtx, router)

			// Setup child with panic handling
			if tt.expectPanic {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic
						assert.Contains(t, r.(string), "router not provided")
					}
				}()
			}
			childRouter = tt.setupChild(childCtx)

			// Validate if no panic expected
			if !tt.expectPanic {
				tt.validateChild(t, router, childRouter)
			}
		})
	}
}

// TestProvideRouter_NestedComponents tests router access in deeply nested components
func TestProvideRouter_NestedComponents(t *testing.T) {
	// Create router
	router := NewRouter()

	// Track which components accessed the router
	var grandparentRouter, parentRouter, childRouter *Router

	// Create contexts
	grandparentCtx := bubbly.NewTestContext()
	parentCtx := bubbly.NewTestContext()
	childCtx := bubbly.NewTestContext()

	// Set up hierarchy: grandparent -> parent -> child
	bubbly.SetParent(parentCtx, grandparentCtx)
	bubbly.SetParent(childCtx, parentCtx)

	// Grandparent provides router
	ProvideRouter(grandparentCtx, router)
	grandparentRouter = UseRouter(grandparentCtx)

	// Parent accesses router
	parentRouter = UseRouter(parentCtx)

	// Child accesses router
	childRouter = UseRouter(childCtx)

	// Assert all components access the same router instance
	assert.Equal(t, router, grandparentRouter, "Grandparent should access provided router")
	assert.Equal(t, router, parentRouter, "Parent should access same router")
	assert.Equal(t, router, childRouter, "Child should access same router")
	assert.Equal(t, grandparentRouter, parentRouter, "Parent and grandparent should share router")
	assert.Equal(t, parentRouter, childRouter, "Child and parent should share router")
}

// TestProvideRouter_MultipleRouters tests multiple routers in different component trees
func TestProvideRouter_MultipleRouters(t *testing.T) {
	// Create two separate routers
	router1 := NewRouter()
	router2 := NewRouter()

	// Verify routers are different instances (pointer comparison)
	if router1 == router2 {
		t.Fatalf("NewRouter() returned same instance: %p == %p", router1, router2)
	}

	// Track which router each tree accesses
	var tree1Router, tree2Router *Router

	// Create first component tree
	tree1Root := bubbly.NewTestContext()
	tree1Child := bubbly.NewTestContext()
	bubbly.SetParent(tree1Child, tree1Root)

	ProvideRouter(tree1Root, router1)
	tree1Router = UseRouter(tree1Child)

	// Create second component tree
	tree2Root := bubbly.NewTestContext()
	tree2Child := bubbly.NewTestContext()
	bubbly.SetParent(tree2Child, tree2Root)

	ProvideRouter(tree2Root, router2)
	tree2Router = UseRouter(tree2Child)

	// Assert each tree has its own router
	assert.Equal(t, router1, tree1Router, "Tree 1 should access router1")
	assert.Equal(t, router2, tree2Router, "Tree 2 should access router2")

	// Verify they're different pointers
	if tree1Router == tree2Router {
		t.Errorf("Both trees got same router: %p", tree1Router)
	}
}

// TestProvideRouter_WithRouteComposable tests that ProvideRouter works with UseRoute
func TestProvideRouter_WithRouteComposable(t *testing.T) {
	// Create router
	router := NewRouter()

	// Track route ref
	var routeRef *bubbly.Ref[*Route]

	// Create parent and child contexts
	parentCtx := bubbly.NewTestContext()
	childCtx := bubbly.NewTestContext()
	bubbly.SetParent(childCtx, parentCtx)

	// Parent provides router
	ProvideRouter(parentCtx, router)

	// Child uses UseRoute (which internally calls UseRouter)
	routeRef = UseRoute(childCtx)

	// Assert route ref is accessible
	assert.NotNil(t, routeRef, "UseRoute should return route ref")

	// Current route should be nil initially (no navigation yet)
	currentRoute := routeRef.GetTyped()
	assert.Nil(t, currentRoute, "Route should be nil initially")
}
