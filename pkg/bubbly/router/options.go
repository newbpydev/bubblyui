package router

// RouteOption is a function that configures a RouteRecord.
//
// RouteOption follows the functional options pattern, allowing flexible
// and composable route configuration. Multiple options can be applied
// to a single route to configure different aspects.
//
// Example:
//
//	builder.RouteWithOptions("/dashboard",
//		WithName("dashboard"),
//		WithMeta(map[string]interface{}{
//			"requiresAuth": true,
//		}),
//		WithGuard(authGuard),
//	)
type RouteOption func(*RouteRecord)

// WithName sets the route name.
//
// The route name is used for named navigation, allowing you to navigate
// by name instead of path. Names must be unique across all routes.
//
// Parameters:
//   - name: The route name (e.g., "user-detail", "dashboard")
//
// Returns:
//   - RouteOption: An option function that sets the name
//
// Example:
//
//	builder.RouteWithOptions("/users/:id",
//		WithName("user-detail"),
//	)
func WithName(name string) RouteOption {
	return func(r *RouteRecord) {
		r.Name = name
	}
}

// WithMeta sets or merges route metadata.
//
// Metadata is arbitrary key-value data attached to a route. Common uses
// include authentication requirements, page titles, permissions, and
// custom application-specific data.
//
// If the route already has metadata, the new metadata is merged with
// existing values. New keys are added, and existing keys are overwritten.
//
// Parameters:
//   - meta: Metadata map to attach to the route
//
// Returns:
//   - RouteOption: An option function that sets/merges metadata
//
// Example:
//
//	builder.RouteWithOptions("/dashboard",
//		WithMeta(map[string]interface{}{
//			"requiresAuth": true,
//			"title":        "Dashboard",
//			"roles":        []string{"admin", "user"},
//		}),
//	)
func WithMeta(meta map[string]interface{}) RouteOption {
	return func(r *RouteRecord) {
		if r.Meta == nil {
			r.Meta = make(map[string]interface{})
		}
		// Merge metadata
		for k, v := range meta {
			r.Meta[k] = v
		}
	}
}

// WithGuard sets a per-route navigation guard.
//
// Per-route guards (also called beforeEnter guards) execute only when
// navigating to this specific route. They run after global before guards
// but before component guards.
//
// The guard is stored in the route's metadata under the "beforeEnter" key,
// following Vue Router's convention.
//
// Parameters:
//   - guard: The navigation guard function
//
// Returns:
//   - RouteOption: An option function that sets the guard
//
// Example:
//
//	authGuard := func(to, from *Route, next NextFunc) {
//		if !isAuthenticated() {
//			next(&NavigationTarget{Path: "/login"})
//		} else {
//			next(nil)
//		}
//	}
//
//	builder.RouteWithOptions("/dashboard",
//		WithGuard(authGuard),
//	)
func WithGuard(guard NavigationGuard) RouteOption {
	return func(r *RouteRecord) {
		if r.Meta == nil {
			r.Meta = make(map[string]interface{})
		}
		r.Meta["beforeEnter"] = guard
	}
}

// WithComponent sets the component to render for this route.
//
// The component is used by RouterView to render the matched route.
// The component should implement bubbly.Component interface.
//
// Parameters:
//   - component: The component to render (should be bubbly.Component)
//
// Returns:
//   - RouteOption: An option function that sets the component
//
// Example:
//
//	userComponent := bubbly.NewComponent("User").Build()
//
//	builder.RouteWithOptions("/user/:id",
//		WithName("user"),
//		WithComponent(userComponent),
//	)
func WithComponent(component interface{}) RouteOption {
	return func(r *RouteRecord) {
		r.Component = component
	}
}

// WithChildren sets or appends child routes for nested routing.
//
// Child routes create a hierarchical routing structure where routes
// can have nested sub-routes. This is useful for layouts with nested
// views, such as a dashboard with multiple sections.
//
// If the route already has children, the new children are appended
// to the existing list.
//
// Parameters:
//   - children: One or more child RouteRecords
//
// Returns:
//   - RouteOption: An option function that sets/appends children
//
// Example:
//
//	dashboardOverview := &RouteRecord{
//		Path: "/overview",
//		Name: "dashboard-overview",
//	}
//
//	dashboardSettings := &RouteRecord{
//		Path: "/settings",
//		Name: "dashboard-settings",
//	}
//
//	builder.RouteWithOptions("/dashboard",
//		WithName("dashboard"),
//		WithChildren(dashboardOverview, dashboardSettings),
//	)
func WithChildren(children ...*RouteRecord) RouteOption {
	return func(r *RouteRecord) {
		if r.Children == nil {
			r.Children = make([]*RouteRecord, 0, len(children))
		}
		r.Children = append(r.Children, children...)
	}
}
