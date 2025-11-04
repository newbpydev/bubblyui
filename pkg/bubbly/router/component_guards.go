package router

// ComponentGuards is an optional interface that components can implement
// to receive navigation lifecycle hooks.
//
// Component guards execute during navigation after global guards and route-specific
// guards. They provide fine-grained control over navigation at the component level.
//
// Guard Execution Order:
//  1. Global beforeEach guards
//  2. Route-specific beforeEnter guards
//  3. Component BeforeRouteLeave (old component)
//  4. Component BeforeRouteEnter (new component)
//  5. Component BeforeRouteUpdate (if component reused)
//  6. Global afterEach hooks
//
// Use Cases:
//   - Confirm before leaving a form with unsaved changes (BeforeRouteLeave)
//   - Fetch data before entering a route (BeforeRouteEnter)
//   - React to parameter changes in the same component (BeforeRouteUpdate)
//
// Example:
//
//	type MyComponent struct {
//	    // ... component fields
//	}
//
//	func (c *MyComponent) BeforeRouteEnter(to, from *Route, next NextFunc) {
//	    // Fetch data before entering
//	    data := fetchUserData(to.Params["id"])
//	    if data == nil {
//	        next(&NavigationTarget{Path: "/404"})
//	    } else {
//	        next(nil) // Proceed with navigation
//	    }
//	}
//
//	func (c *MyComponent) BeforeRouteLeave(to, from *Route, next NextFunc) {
//	    // Confirm before leaving with unsaved changes
//	    if c.hasUnsavedChanges {
//	        // In a real TUI, you'd show a confirmation dialog
//	        // For now, cancel navigation
//	        next(&NavigationTarget{Path: ""}) // Cancel
//	    } else {
//	        next(nil) // Allow navigation
//	    }
//	}
//
//	func (c *MyComponent) BeforeRouteUpdate(to, from *Route, next NextFunc) {
//	    // React to parameter changes
//	    c.loadData(to.Params["id"])
//	    next(nil)
//	}
type ComponentGuards interface {
	// BeforeRouteEnter is called before the route that renders this component is confirmed.
	//
	// This guard is called BEFORE the component is created, so it does not have access
	// to the component's state. Use this for:
	//   - Fetching data before entering the route
	//   - Checking permissions
	//   - Redirecting based on conditions
	//
	// Parameters:
	//   - to: The target route being navigated to
	//   - from: The current route being navigated away from (nil on first navigation)
	//   - next: Function to control navigation flow
	//
	// Calling next:
	//   - next(nil): Proceed with navigation
	//   - next(&NavigationTarget{Path: "/other"}): Redirect to another route
	//   - next(&NavigationTarget{Path: ""}): Cancel navigation
	//
	// Example:
	//
	//	func (c *MyComponent) BeforeRouteEnter(to, from *Route, next NextFunc) {
	//	    if !isAuthenticated() {
	//	        next(&NavigationTarget{Path: "/login"})
	//	        return
	//	    }
	//	    next(nil)
	//	}
	BeforeRouteEnter(to, from *Route, next NextFunc)

	// BeforeRouteUpdate is called when the route that renders this component has changed,
	// but the component is reused in the new route.
	//
	// This happens when navigating between routes with the same component but different
	// parameters (e.g., /user/1 to /user/2). Use this for:
	//   - Reloading data when parameters change
	//   - Updating component state based on new route
	//   - Validating new parameters
	//
	// Parameters:
	//   - to: The target route being navigated to
	//   - from: The current route being navigated away from
	//   - next: Function to control navigation flow
	//
	// Calling next:
	//   - next(nil): Proceed with navigation
	//   - next(&NavigationTarget{Path: "/other"}): Redirect to another route
	//   - next(&NavigationTarget{Path: ""}): Cancel navigation
	//
	// Example:
	//
	//	func (c *MyComponent) BeforeRouteUpdate(to, from *Route, next NextFunc) {
	//	    // Reload data for new user ID
	//	    c.userData = fetchUser(to.Params["id"])
	//	    next(nil)
	//	}
	BeforeRouteUpdate(to, from *Route, next NextFunc)

	// BeforeRouteLeave is called when the route that renders this component is about
	// to be navigated away from.
	//
	// This guard has access to the component's state and can prevent navigation.
	// Use this for:
	//   - Confirming before leaving with unsaved changes
	//   - Cleaning up resources
	//   - Saving state before leaving
	//
	// Parameters:
	//   - to: The target route being navigated to
	//   - from: The current route being navigated away from
	//   - next: Function to control navigation flow
	//
	// Calling next:
	//   - next(nil): Proceed with navigation
	//   - next(&NavigationTarget{Path: "/other"}): Redirect to another route
	//   - next(&NavigationTarget{Path: ""}): Cancel navigation
	//
	// Example:
	//
	//	func (c *MyComponent) BeforeRouteLeave(to, from *Route, next NextFunc) {
	//	    if c.hasUnsavedChanges {
	//	        // Show confirmation (in real TUI)
	//	        // For now, cancel navigation
	//	        next(&NavigationTarget{Path: ""})
	//	        return
	//	    }
	//	    next(nil)
	//	}
	BeforeRouteLeave(to, from *Route, next NextFunc)
}

// hasComponentGuards checks if a component implements ComponentGuards interface.
//
// This is a helper function used internally by the router to determine
// if component guards should be executed during navigation.
//
// Parameters:
//   - component: The component to check (can be nil or any type)
//
// Returns:
//   - ComponentGuards: The component cast to ComponentGuards interface, or nil
//   - bool: true if the component implements ComponentGuards, false otherwise
//
// Example:
//
//	if guards, ok := hasComponentGuards(component); ok {
//	    guards.BeforeRouteEnter(to, from, next)
//	}
func hasComponentGuards(component interface{}) (ComponentGuards, bool) {
	if component == nil {
		return nil, false
	}
	guards, ok := component.(ComponentGuards)
	return guards, ok
}
