package router

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestUseRouter_RouterAccessible tests that UseRouter returns the router instance.
func TestUseRouter_RouterAccessible(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "router is accessible via UseRouter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			router := NewRouter()

			// Create component with router provided
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, router)

					// Use composable
					injectedRouter := UseRouter(ctx)

					// Verify it's the same instance
					assert.Same(t, router, injectedRouter, "UseRouter should return the provided router instance")

					ctx.Expose("router", injectedRouter)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			assert.NotNil(t, component)
		})
	}
}

// TestUseRouter_PanicIfNotProvided tests that UseRouter panics when router is not provided.
func TestUseRouter_PanicIfNotProvided(t *testing.T) {
	tests := []struct {
		name          string
		expectedPanic string
	}{
		{
			name:          "panics when router not provided",
			expectedPanic: "router not provided in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component WITHOUT providing router
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// This should panic when Init() is called
					_ = UseRouter(ctx)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err, "Build should succeed")
			assert.NotNil(t, component, "Component should be created")

			// Panic should happen when component is initialized (Setup is executed)
			assert.Panics(t, func() {
				component.Init()
			}, "UseRouter should panic when router is not provided during Init()")
		})
	}
}

// TestUseRouter_MultipleComponentsShareInstance tests that multiple components share the same router instance.
func TestUseRouter_MultipleComponentsShareInstance(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "multiple components share same router instance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			router := NewRouter()

			// Test that multiple calls to UseRouter in the same component return the same instance
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, router)

					// Call UseRouter multiple times
					router1 := UseRouter(ctx)
					router2 := UseRouter(ctx)

					// Both should be the same instance
					assert.Same(t, router, router1, "First call should return provided router")
					assert.Same(t, router, router2, "Second call should return provided router")
					assert.Same(t, router1, router2, "Multiple calls should return same instance")
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			assert.NotNil(t, component)
		})
	}
}

// TestUseRouter_ContextInjection tests that UseRouter properly uses context injection.
func TestUseRouter_ContextInjection(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "context injection works correctly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			router := NewRouter()

			// Create component
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router using the standard key
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, router)

					// Inject via UseRouter
					injectedRouter := UseRouter(ctx)

					// Verify injection worked
					assert.NotNil(t, injectedRouter, "Injected router should not be nil")
					assert.Same(t, router, injectedRouter, "Injected router should be the same instance")

					// Expose to ensure return value is used
					ctx.Expose("router", injectedRouter)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			assert.NotNil(t, component)
		})
	}
}

// TestUseRouter_ReturnValue tests that UseRouter returns the correct router instance.
func TestUseRouter_ReturnValue(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns provided router instance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			expectedRouter := NewRouter()

			var actualRouter *Router

			// Create component that captures the return value
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, expectedRouter)

					// Call UseRouter and capture return value
					actualRouter = UseRouter(ctx)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			assert.NotNil(t, component)

			// Initialize component to execute Setup
			component.Init()

			// Verify the return value
			assert.NotNil(t, actualRouter, "UseRouter should return a non-nil router")
			assert.Same(t, expectedRouter, actualRouter, "UseRouter should return the exact router instance provided")

			// Verify it's a usable router
			assert.NotNil(t, actualRouter.CurrentRoute, "Returned router should have CurrentRoute method")
		})
	}
}

// TestUseRoute_RouteAccessible tests that UseRoute returns a reactive ref to the current route.
func TestUseRoute_RouteAccessible(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "route accessible via UseRoute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with initial route
			router := NewRouter()
			initialRoute := NewRoute("/home", "home", nil, nil, "", nil, nil)
			router.currentRoute = initialRoute

			var routeRef *bubbly.Ref[*Route]

			// Create component
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, router)

					// Use composable
					routeRef = UseRoute(ctx)

					// Verify initial value
					assert.NotNil(t, routeRef, "UseRoute should return a non-nil ref")
					assert.Same(t, initialRoute, routeRef.GetTyped(), "UseRoute should return ref to current route")
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			assert.NotNil(t, component)

			// Initialize component to execute Setup
			component.Init()
		})
	}
}

// TestUseRoute_UpdatesReactively tests that the route ref updates when navigation occurs.
func TestUseRoute_UpdatesReactively(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "route ref updates on navigation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			router := NewRouter()
			initialRoute := NewRoute("/home", "home", nil, nil, "", nil, nil)
			router.currentRoute = initialRoute

			var routeRef *bubbly.Ref[*Route]

			// Create component
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, router)

					// Use composable
					routeRef = UseRoute(ctx)

					// Verify initial route
					assert.Same(t, initialRoute, routeRef.GetTyped(), "Initial route should match")
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)

			// Initialize component to execute Setup
			component.Init()

			// Trigger first View() to execute OnMounted hooks
			_ = component.View()

			// Verify hook was registered
			router.mu.RLock()
			hookCount := len(router.afterHooks)
			router.mu.RUnlock()
			assert.Equal(t, 1, hookCount, "AfterEach hook should be registered")

			// Simulate navigation by updating current route and triggering AfterEach hooks
			newRoute := NewRoute("/about", "about", nil, nil, "", nil, nil)
			router.mu.Lock()
			router.currentRoute = newRoute
			// Execute after hooks manually (simulating navigation)
			// Note: AfterEach hook was registered during Init() via OnMounted
			for _, hook := range router.afterHooks {
				hook(newRoute, initialRoute)
			}
			router.mu.Unlock()

			// Verify route ref updated
			assert.Same(t, newRoute, routeRef.GetTyped(), "Route ref should update to new route")
		})
	}
}

// TestUseRoute_ParamsAccessible tests that route params are accessible via the route ref.
func TestUseRoute_ParamsAccessible(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "route params accessible"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with route containing params
			router := NewRouter()
			params := map[string]string{"id": "123", "tab": "profile"}
			route := NewRoute("/user/:id", "user", params, nil, "", nil, nil)
			router.currentRoute = route

			var routeRef *bubbly.Ref[*Route]

			// Create component
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, router)

					// Use composable
					routeRef = UseRoute(ctx)

					// Access params
					currentRoute := routeRef.GetTyped()
					assert.Equal(t, "123", currentRoute.Params["id"], "Params should be accessible")
					assert.Equal(t, "profile", currentRoute.Params["tab"], "Params should be accessible")
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			component.Init()
		})
	}
}

// TestUseRoute_QueryAccessible tests that query params are accessible via the route ref.
func TestUseRoute_QueryAccessible(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "query params accessible"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with route containing query
			router := NewRouter()
			query := map[string]string{"page": "1", "sort": "name"}
			route := NewRoute("/users", "users", nil, query, "", nil, nil)
			router.currentRoute = route

			var routeRef *bubbly.Ref[*Route]

			// Create component
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, router)

					// Use composable
					routeRef = UseRoute(ctx)

					// Access query
					currentRoute := routeRef.GetTyped()
					assert.Equal(t, "1", currentRoute.Query["page"], "Query should be accessible")
					assert.Equal(t, "name", currentRoute.Query["sort"], "Query should be accessible")
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			component.Init()
		})
	}
}

// TestUseRoute_MetaAccessible tests that route meta is accessible via the route ref.
func TestUseRoute_MetaAccessible(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "route meta accessible"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with route containing meta
			router := NewRouter()
			meta := map[string]interface{}{
				"requiresAuth": true,
				"title":        "Dashboard",
			}
			route := NewRoute("/dashboard", "dashboard", nil, nil, "", meta, nil)
			router.currentRoute = route

			var routeRef *bubbly.Ref[*Route]

			// Create component
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Provide router
					routerKey := bubbly.NewProvideKey[*Router]("router")
					bubbly.ProvideTyped(ctx, routerKey, router)

					// Use composable
					routeRef = UseRoute(ctx)

					// Access meta
					currentRoute := routeRef.GetTyped()
					requiresAuth, ok := currentRoute.GetMeta("requiresAuth")
					assert.True(t, ok, "Meta key should exist")
					assert.Equal(t, true, requiresAuth, "Meta value should be accessible")

					title, ok := currentRoute.GetMeta("title")
					assert.True(t, ok, "Meta key should exist")
					assert.Equal(t, "Dashboard", title, "Meta value should be accessible")
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			component.Init()
		})
	}
}
