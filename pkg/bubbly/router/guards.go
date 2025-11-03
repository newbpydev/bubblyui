package router

import (
	"errors"
)

var (
	// ErrNavigationCancelled is returned when navigation is cancelled by a guard
	ErrNavigationCancelled = errors.New("navigation cancelled by guard")
)

// guardAction represents the action a guard wants to take
type guardAction int

const (
	guardContinue guardAction = iota // Continue to next guard
	guardCancel                      // Cancel navigation
	guardRedirect                    // Redirect to different route
)

// guardResult represents the result of guard execution
type guardResult struct {
	action guardAction
	target *NavigationTarget
}

// BeforeEach registers a global before guard.
//
// Before guards execute before every navigation and can inspect the target
// route, current route, and control navigation flow via the next() function.
//
// Parameters:
//   - guard: The guard function to register
//
// Guard Execution:
// Guards are executed in the order they are registered. Each guard must
// call next() to continue the guard chain. If a guard doesn't call next(),
// navigation will hang (this is a programming error).
//
// Guard Actions:
//   - next(nil): Allow navigation, continue to next guard
//   - next(&NavigationTarget{}): Cancel navigation (empty target)
//   - next(&NavigationTarget{Path: "/other"}): Redirect to different route
//
// Thread Safety:
// This method acquires a write lock and is safe for concurrent use.
// However, guards are typically registered during router setup, not
// during navigation.
//
// Example:
//
//	router.BeforeEach(func(to, from *Route, next NextFunc) {
//		if to.Meta["requiresAuth"] == true && !isAuthenticated() {
//			// Redirect to login
//			next(&NavigationTarget{
//				Path: "/login",
//				Query: map[string]string{"redirect": to.FullPath},
//			})
//		} else {
//			// Allow navigation
//			next(nil)
//		}
//	})
//
// Use Cases:
//   - Authentication checks
//   - Authorization checks
//   - Data fetching before route
//   - Route validation
//   - Analytics tracking
//   - Logging
func (r *Router) BeforeEach(guard NavigationGuard) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.beforeHooks = append(r.beforeHooks, guard)
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
// Hook Execution:
// Hooks are executed in the order they are registered, after the route
// has changed and the RouteChangedMsg has been generated.
//
// Thread Safety:
// This method acquires a write lock and is safe for concurrent use.
//
// Example:
//
//	router.AfterEach(func(to, from *Route) {
//		// Track page view
//		analytics.TrackPageView(to.Path)
//
//		// Update document title
//		if title, ok := to.GetMeta("title"); ok {
//			setWindowTitle(title.(string))
//		}
//
//		// Log navigation
//		log.Printf("Navigated from %v to %s", from, to.Path)
//	})
//
// Use Cases:
//   - Analytics tracking
//   - Logging
//   - Focus management
//   - Document title updates
//   - Breadcrumb updates
//   - State persistence
func (r *Router) AfterEach(hook AfterNavigationHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.afterHooks = append(r.afterHooks, hook)
}

// executeBeforeGuards executes all before guards in order.
//
// Guards are executed sequentially. Each guard must call next() to continue.
// If a guard cancels or redirects, remaining guards are skipped.
//
// Parameters:
//   - to: The target route
//   - from: The current route (nil if no current route)
//
// Returns:
//   - *guardResult: The result of guard execution (continue, cancel, or redirect)
//
// Guard Flow:
//  1. Execute each guard in order
//  2. Guard calls next(nil) → continue to next guard
//  3. Guard calls next(&NavigationTarget{}) → cancel, stop execution
//  4. Guard calls next(&NavigationTarget{Path: "..."}) → redirect, stop execution
//  5. If all guards allow, return guardContinue
//
// Thread Safety:
// This method reads beforeHooks with a read lock.
func (r *Router) executeBeforeGuards(to, from *Route) *guardResult {
	r.mu.RLock()
	guards := make([]NavigationGuard, len(r.beforeHooks))
	copy(guards, r.beforeHooks)
	r.mu.RUnlock()

	// Execute global guards sequentially
	for _, guard := range guards {
		result := &guardResult{action: guardContinue}

		// Create next function that captures the result
		next := func(target *NavigationTarget) {
			if target == nil {
				// Allow navigation
				result.action = guardContinue
			} else if target.Path == "" && target.Name == "" {
				// Empty target = cancel
				result.action = guardCancel
			} else {
				// Redirect
				result.action = guardRedirect
				result.target = target
			}
		}

		// Execute guard
		guard(to, from, next)

		// Check result
		if result.action == guardCancel {
			return &guardResult{action: guardCancel}
		}

		if result.action == guardRedirect {
			return &guardResult{
				action: guardRedirect,
				target: result.target,
			}
		}

		// Continue to next guard
	}

	// Execute route-specific beforeEnter guard if present
	if to != nil && to.Meta != nil {
		if beforeEnter, ok := to.Meta["beforeEnter"].(NavigationGuard); ok {
			result := &guardResult{action: guardContinue}

			next := func(target *NavigationTarget) {
				if target == nil {
					result.action = guardContinue
				} else if target.Path == "" && target.Name == "" {
					result.action = guardCancel
				} else {
					result.action = guardRedirect
					result.target = target
				}
			}

			beforeEnter(to, from, next)

			if result.action == guardCancel {
				return &guardResult{action: guardCancel}
			}

			if result.action == guardRedirect {
				return &guardResult{
					action: guardRedirect,
					target: result.target,
				}
			}
		}
	}

	// Execute component guards
	componentGuardResult := r.executeComponentGuards(to, from)
	if componentGuardResult.action != guardContinue {
		return componentGuardResult
	}

	// All guards passed
	return &guardResult{action: guardContinue}
}

// executeComponentGuards executes component-level navigation guards.
//
// Component guards are executed after global and route-specific guards.
// The execution order is:
//  1. BeforeRouteLeave on old component (if exists)
//  2. BeforeRouteEnter on new component (if exists)
//  3. BeforeRouteUpdate on component (if same component, different params)
//
// Parameters:
//   - to: The target route
//   - from: The current route (nil if no current route)
//
// Returns:
//   - *guardResult: The result of component guard execution
func (r *Router) executeComponentGuards(to, from *Route) *guardResult {
	// Get old and new components
	var oldComponent, newComponent interface{}

	if from != nil && len(from.Matched) > 0 {
		// Get the leaf component (last in matched array)
		oldComponent = from.Matched[len(from.Matched)-1].Component
	}

	if to != nil && len(to.Matched) > 0 {
		// Get the leaf component (last in matched array)
		newComponent = to.Matched[len(to.Matched)-1].Component
	}

	// Check if components implement ComponentGuards
	oldGuards, oldHasGuards := hasComponentGuards(oldComponent)
	newGuards, newHasGuards := hasComponentGuards(newComponent)

	// Determine if component is being reused (same component instance, different params)
	// Use pointer comparison for component reuse detection
	componentReused := oldComponent != nil && newComponent != nil && oldComponent == newComponent

	// Execute BeforeRouteLeave on old component
	if oldHasGuards && oldComponent != newComponent {
		result := &guardResult{action: guardContinue}

		next := func(target *NavigationTarget) {
			if target == nil {
				result.action = guardContinue
			} else if target.Path == "" && target.Name == "" {
				result.action = guardCancel
			} else {
				result.action = guardRedirect
				result.target = target
			}
		}

		oldGuards.BeforeRouteLeave(to, from, next)

		if result.action != guardContinue {
			return result
		}
	}

	// Execute BeforeRouteUpdate if component is reused
	if componentReused && newHasGuards {
		result := &guardResult{action: guardContinue}

		next := func(target *NavigationTarget) {
			if target == nil {
				result.action = guardContinue
			} else if target.Path == "" && target.Name == "" {
				result.action = guardCancel
			} else {
				result.action = guardRedirect
				result.target = target
			}
		}

		newGuards.BeforeRouteUpdate(to, from, next)

		if result.action != guardContinue {
			return result
		}
	}

	// Execute BeforeRouteEnter on new component (if not reused)
	if newHasGuards && !componentReused {
		result := &guardResult{action: guardContinue}

		next := func(target *NavigationTarget) {
			if target == nil {
				result.action = guardContinue
			} else if target.Path == "" && target.Name == "" {
				result.action = guardCancel
			} else {
				result.action = guardRedirect
				result.target = target
			}
		}

		newGuards.BeforeRouteEnter(to, from, next)

		if result.action != guardContinue {
			return result
		}
	}

	return &guardResult{action: guardContinue}
}

// executeAfterHooks executes all after hooks in order.
//
// After hooks are called after navigation completes successfully.
// They cannot affect navigation.
//
// Parameters:
//   - to: The new current route
//   - from: The previous route (nil if no previous route)
//
// Thread Safety:
// This method reads afterHooks with a read lock.
func (r *Router) executeAfterHooks(to, from *Route) {
	r.mu.RLock()
	hooks := make([]AfterNavigationHook, len(r.afterHooks))
	copy(hooks, r.afterHooks)
	r.mu.RUnlock()

	// Execute hooks sequentially
	for _, hook := range hooks {
		hook(to, from)
	}
}
