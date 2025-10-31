package bubbly

// Dependency represents a reactive value that can be watched and tracked.
// It serves as the unified interface for all reactive primitives in BubblyUI,
// enabling both dependency tracking (for computed values) and value access
// (for composables like UseEffect).
//
// The Dependency interface is implemented by both Ref[T] and Computed[T],
// allowing them to be used interchangeably in contexts that need to track
// reactive dependencies or access reactive values.
//
// Key capabilities:
//   - Get() returns the current value as any, enabling type-erased access
//   - Invalidate() marks the dependency as needing recomputation
//   - AddDependent() registers dependencies for automatic invalidation
//
// This interface enables several important patterns:
//   - UseEffect can accept any reactive value (Ref or Computed) as dependencies
//   - Computed values can be watched directly (not just Refs)
//   - Dependency chains automatically propagate invalidation
//   - Type-safe refs work seamlessly with composables
//
// Example usage:
//
//	// With UseEffect - typed refs work directly
//	count := bubbly.NewRef(0)  // *Ref[int]
//	UseEffect(ctx, func() UseEffectCleanup {
//	    currentCount := count.Get().(int)
//	    fmt.Printf("Count: %d\n", currentCount)
//	    return nil
//	}, count)  // Works! count implements Dependency
//
//	// With Computed values
//	fullName := ctx.Computed(func() string {
//	    return firstName.Get() + " " + lastName.Get()
//	})
//	UseEffect(ctx, func() UseEffectCleanup {
//	    name := fullName.Get().(string)
//	    fmt.Printf("Name: %s\n", name)
//	    return nil
//	}, fullName)  // Computed as dependency!
//
// Design rationale:
// Go's type system doesn't support covariance, so *Ref[int] cannot be used
// where *Ref[any] is expected. The Dependency interface solves this by providing
// a common contract that all reactive types implement, enabling polymorphic
// usage while maintaining type safety within each implementation.
//
// See also: Ref, Computed, UseEffect, Watch
type Dependency interface {
	// Get returns the current value of the dependency as any.
	// Callers must type assert the result to the expected type.
	//
	// For Ref[T], this returns the stored value.
	// For Computed[T], this triggers evaluation if dirty and returns the cached result.
	//
	// This method enables type-erased access to reactive values, which is necessary
	// for composables like UseEffect that need to work with dependencies of any type.
	//
	// Example:
	//   count := bubbly.NewRef(42)
	//   value := count.Get().(int)  // Type assertion required
	Get() any

	// Invalidate marks this dependency as needing recomputation or re-evaluation.
	// This method is called automatically by the reactivity system when a dependency changes.
	//
	// For Ref[T], this notifies all watchers and dependents.
	// For Computed[T], this marks the cache as dirty and propagates to dependents.
	//
	// The invalidation propagates through the dependency graph, ensuring that
	// all dependent computed values are marked as needing recomputation.
	Invalidate()

	// AddDependent registers another dependency that depends on this one.
	// When this dependency is invalidated, all registered dependents are also invalidated.
	//
	// This method is used internally by the reactivity system to build the dependency graph.
	// It enables automatic invalidation propagation through chains of computed values.
	//
	// Example dependency chain:
	//   ref → computed1 → computed2
	// When ref changes, computed1 and computed2 are both invalidated automatically.
	AddDependent(dep Dependency)
}
