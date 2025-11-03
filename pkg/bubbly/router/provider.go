package router

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ProvideRouter injects a router instance into the component tree.
//
// This helper function provides the router to all descendant components
// using the standard provide/inject pattern. Child components can access
// the router via UseRouter(ctx) or UseRoute(ctx) composables.
//
// The router is provided using a type-safe key (routerKey) that is
// shared between ProvideRouter and UseRouter, ensuring type safety
// and preventing accidental type mismatches.
//
// Parameters:
//   - ctx: The component context where the router should be provided
//   - router: The router instance to inject into the component tree
//
// Usage:
//
//	// In root/parent component
//	Setup(func(ctx *bubbly.Context) {
//	    router := router.NewRouter()
//	    router.ProvideRouter(ctx, router)
//	})
//
//	// In any child component
//	Setup(func(ctx *bubbly.Context) {
//	    router := router.UseRouter(ctx)
//	    router.Push(&router.NavigationTarget{Path: "/home"})
//	})
//
// Best Practices:
//   - Call ProvideRouter at the root component level
//   - Provide router before any child components are initialized
//   - Only provide one router per component tree
//   - Use UseRouter/UseRoute in children to access the router
//
// Example with RouterBuilder:
//
//	Setup(func(ctx *bubbly.Context) {
//	    router, err := router.NewRouterBuilder().
//	        Route("/", "home").
//	        Route("/about", "about").
//	        Build()
//	    if err != nil {
//	        panic(err)
//	    }
//	    router.ProvideRouter(ctx, router)
//	})
//
// Multiple Routers:
// Different component trees can have different routers. Each tree
// maintains its own router instance via provide/inject:
//
//	// Tree 1
//	tree1Root.Setup(func(ctx *bubbly.Context) {
//	    router1 := router.NewRouter()
//	    router.ProvideRouter(ctx, router1)
//	})
//
//	// Tree 2
//	tree2Root.Setup(func(ctx *bubbly.Context) {
//	    router2 := router.NewRouter()
//	    router.ProvideRouter(ctx, router2)
//	})
//
// Thread Safety:
// The router instance itself is thread-safe. Multiple components can
// safely access the provided router concurrently via UseRouter.
//
// See Also:
//   - UseRouter: Access the provided router instance
//   - UseRoute: Access reactive current route state
func ProvideRouter(ctx *bubbly.Context, router *Router) {
	bubbly.ProvideTyped(ctx, routerKey, router)
}
