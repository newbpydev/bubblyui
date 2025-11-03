package router

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// routerKey is the standard provide/inject key for the router instance.
// This key is used internally by UseRouter to retrieve the router from context.
var routerKey = bubbly.NewProvideKey[*Router]("router")

// UseRouter provides access to the router instance via context injection.
//
// This composable retrieves the router that was provided by an ancestor component
// using the provide/inject pattern. It enables components to access router methods
// for programmatic navigation without prop drilling.
//
// The router must be provided by a parent component using:
//
//	routerKey := bubbly.NewProvideKey[*Router]("router")
//	bubbly.ProvideTyped(ctx, routerKey, router)
//
// Parameters:
//   - ctx: The component context (required for all composables)
//
// Returns:
//   - *Router: The router instance provided by an ancestor component
//
// Panics:
//   - If no router is provided in the component tree
//   - If the provided value is not a *Router (type assertion failure)
//
// Example:
//
//	// In parent component (e.g., App root)
//	Setup(func(ctx *bubbly.Context) {
//	    router := router.NewRouter()
//	    routerKey := bubbly.NewProvideKey[*Router]("router")
//	    bubbly.ProvideTyped(ctx, routerKey, router)
//	})
//
//	// In child component
//	Setup(func(ctx *bubbly.Context) {
//	    router := router.UseRouter(ctx)
//
//	    ctx.On("navigate", func(data interface{}) {
//	        path := data.(string)
//	        router.Push(&router.NavigationTarget{Path: path})
//	    })
//	})
//
// Navigation Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    router := router.UseRouter(ctx)
//
//	    ctx.On("goToUser", func(data interface{}) {
//	        userID := data.(string)
//	        router.Push(&router.NavigationTarget{
//	            Name:   "user-detail",
//	            Params: map[string]string{"id": userID},
//	        })
//	    })
//
//	    ctx.On("goBack", func(_ interface{}) {
//	        router.Back()
//	    })
//	})
//
// Multiple Components:
//
//	// All components in the tree share the same router instance
//	// Component A
//	routerA := router.UseRouter(ctx)
//	routerA.Push(&router.NavigationTarget{Path: "/page1"})
//
//	// Component B (sibling or child)
//	routerB := router.UseRouter(ctx)
//	routerB.Push(&router.NavigationTarget{Path: "/page2"})
//	// routerA and routerB are the same instance
//
// Thread Safety:
// The router instance itself is thread-safe. Multiple components can
// safely call UseRouter concurrently and use the returned router from
// different goroutines.
//
// Best Practices:
//   - Provide router at the root component level
//   - Use UseRouter in any component that needs navigation
//   - Don't pass router as props - use this composable instead
//   - Call UseRouter in Setup(), not in Template()
func UseRouter(ctx *bubbly.Context) *Router {
	// Inject router from context using the standard key
	router := bubbly.InjectTyped(ctx, routerKey, (*Router)(nil))

	// Panic if router was not provided
	if router == nil {
		panic("router not provided in context")
	}

	return router
}
