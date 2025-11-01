package composables

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// UseEffectCleanup is a function that performs cleanup operations.
// It is returned by the effect function and executed before the effect
// re-runs or when the component unmounts.
//
// Cleanup functions are optional - an effect can return nil if no cleanup
// is needed.
//
// Example:
//
//	UseEffect(ctx, func() UseEffectCleanup {
//	    // Setup code
//	    timer := time.NewTicker(time.Second)
//
//	    // Return cleanup function
//	    return func() {
//	        timer.Stop()
//	    }
//	})
type UseEffectCleanup func()

// UseEffect manages side effects with automatic cleanup and dependency tracking.
// It is similar to React's useEffect hook, providing a way to perform side effects
// in components with proper lifecycle management.
//
// The effect function is called:
//   - Once when the component mounts (after first render)
//   - When any dependency changes (if dependencies are provided)
//   - On every update (if no dependencies are provided)
//
// The cleanup function (if returned) is called:
//   - Before the effect re-runs
//   - When the component unmounts
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - effect: Function that performs the side effect and optionally returns cleanup
//   - deps: Optional dependencies - effect only re-runs when these change
//
// Dependency Behavior:
//   - No deps: Effect runs on mount and every update
//   - Empty deps []: Effect runs only once on mount
//   - With deps: Effect runs on mount and when any dep changes
//
// Example - Run once on mount:
//
//	UseEffect(ctx, func() UseEffectCleanup {
//	    data := fetchData()
//	    dataRef.Set(data)
//	    return nil
//	}, []bubbly.Dependency{}...) // Empty deps = run once
//
// Example - Run on dependency change:
//
//	userId := bubbly.NewDependency(1)
//	UseEffect(ctx, func() UseEffectCleanup {
//	    user := fetchUser(userId.GetTyped())
//	    userRef.Set(user)
//	    return nil
//	}, userId) // Re-runs when userId changes
//
// Example - With cleanup:
//
//	UseEffect(ctx, func() UseEffectCleanup {
//	    ticker := time.NewTicker(time.Second)
//	    go func() {
//	        for range ticker.C {
//	            count.Set(count.GetTyped() + 1)
//	        }
//	    }()
//
//	    return func() {
//	        ticker.Stop() // Cleanup
//	    }
//	})
//
// Example - Run on every update:
//
//	UseEffect(ctx, func() UseEffectCleanup {
//	    log.Println("Component updated")
//	    return nil
//	}) // No deps = runs every update
//
// Thread Safety:
//
// UseEffect integrates with the component lifecycle system, which handles
// thread safety and panic recovery. Effects and cleanup functions are
// executed in the component's lifecycle context.
//
// Performance:
//
// UseEffect has minimal overhead as it delegates to the existing lifecycle
// hook system. The performance characteristics match those of OnMounted,
// OnUpdated, and OnUnmounted hooks.
func UseEffect(ctx *bubbly.Context, effect func() UseEffectCleanup, deps ...bubbly.Dependency) {
	// Store the cleanup function from the last effect execution
	var cleanup UseEffectCleanup

	// Create a function that executes cleanup (if exists) then runs the effect
	executeEffect := func() {
		// Run cleanup from previous effect execution
		if cleanup != nil {
			cleanup()
		}

		// Execute the effect and store the new cleanup function
		cleanup = effect()
	}

	// Determine how to register the effect based on dependencies
	if len(deps) == 0 {
		// No dependencies provided: run on mount and every update
		ctx.OnMounted(executeEffect)
		ctx.OnUpdated(executeEffect)
	} else {
		// Dependencies provided: run on mount and when dependencies change
		ctx.OnMounted(executeEffect)
		ctx.OnUpdated(executeEffect, deps...)
	}

	// Register cleanup to run on unmount
	ctx.OnUnmounted(func() {
		if cleanup != nil {
			cleanup()
		}
	})
}
