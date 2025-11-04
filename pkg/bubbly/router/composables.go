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

// UseRoute provides reactive access to the current route.
//
// This composable returns a reactive reference to the current route that
// automatically updates when navigation occurs. The returned Ref can be
// used to access route information (path, params, query, meta) and will
// trigger reactivity when the route changes.
//
// The router must be provided by an ancestor component before calling UseRoute.
//
// Parameters:
//   - ctx: The component context (required for all composables)
//
// Returns:
//   - *bubbly.Ref[*Route]: A reactive reference to the current route
//
// Panics:
//   - If no router is provided in the component tree (via UseRouter)
//
// Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    route := router.UseRoute(ctx)
//
//	    // Access route information
//	    ctx.OnMounted(func() {
//	        currentRoute := route.GetTyped()
//	        fmt.Printf("Current path: %s\n", currentRoute.Path)
//	        fmt.Printf("Params: %v\n", currentRoute.Params)
//	        fmt.Printf("Query: %v\n", currentRoute.Query)
//	    })
//
//	    // Watch for route changes
//	    ctx.Watch(route, func(newVal, oldVal interface{}) {
//	        newRoute := newVal.(*Route)
//	        fmt.Printf("Navigated to: %s\n", newRoute.Path)
//	    })
//
//	    ctx.Expose("route", route)
//	})
//
// Accessing Route Data:
//
//	Setup(func(ctx *bubbly.Context) {
//	    route := router.UseRoute(ctx)
//
//	    // Access params
//	    userID := route.GetTyped().Params["id"]
//
//	    // Access query
//	    page := route.GetTyped().Query["page"]
//
//	    // Access meta
//	    requiresAuth, _ := route.GetTyped().GetMeta("requiresAuth")
//	})
//
// Reactive Updates:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    route := ctx.Get("route").(*bubbly.Ref[*Route])
//	    currentRoute := route.GetTyped()
//
//	    return fmt.Sprintf("Current page: %s", currentRoute.Path)
//	    // This will automatically re-render when route changes
//	})
//
// Thread Safety:
// The returned Ref is thread-safe. Multiple goroutines can safely call
// GetTyped() concurrently. The route update via AfterEach hook is also
// thread-safe as it uses the Ref's Set method.
//
// Best Practices:
//   - Use UseRoute when you need reactive access to route state
//   - Use UseRouter when you only need navigation methods
//   - Access route data via GetTyped() for type safety
//   - Watch the route ref for side effects on navigation
func UseRoute(ctx *bubbly.Context) *bubbly.Ref[*Route] {
	// Get router instance (will panic if not provided)
	router := UseRouter(ctx)

	// Create reactive ref with current route
	routeRef := bubbly.NewRef(router.CurrentRoute())

	// Register AfterEach hook to update ref on navigation
	ctx.OnMounted(func() {
		router.AfterEach(func(to, from *Route) {
			routeRef.Set(to)
		})
	})

	return routeRef
}
