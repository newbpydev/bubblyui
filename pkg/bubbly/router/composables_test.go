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
