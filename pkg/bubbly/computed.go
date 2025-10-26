package bubbly

import "sync"

// Computed is a type-safe computed value that lazily evaluates a function and caches the result.
// It provides thread-safe read operations using a read-write mutex.
// The computation function is only called once on the first Get() call, and the result is cached
// for all subsequent calls.
//
// Computed automatically tracks dependencies (Refs and other Computed values) accessed during
// evaluation. When any dependency changes, the cache is invalidated and the function will
// recompute on the next Get() call.
//
// Computed implements the Dependency interface, allowing it to be used as a dependency
// for other computed values, enabling chained computations.
//
// Example usage:
//
//	count := bubbly.NewRef(5)
//	doubled := bubbly.NewComputed(func() int {
//	    return count.Get() * 2  // Automatically tracks count as dependency
//	})
//	value := doubled.Get()  // Computes and caches: 10
//	count.Set(10)           // Invalidates doubled's cache
//	value2 := doubled.Get() // Recomputes: 20
type Computed[T any] struct {
	mu         sync.RWMutex
	fn         func() T
	cache      T
	dirty      bool
	deps       []Dependency
	dependents []Dependency
}

// NewComputed creates a new computed value with the given computation function.
// The function is not called immediately; evaluation is deferred until the first Get() call.
// This lazy evaluation strategy improves performance by avoiding unnecessary computations.
//
// Type parameter T can be any Go type including primitives, structs, slices, maps, pointers,
// and interfaces.
//
// Example:
//
//	// Simple computation
//	computed := NewComputed(func() int { return 42 })
//
//	// Computation using Ref values
//	count := NewRef(10)
//	doubled := NewComputed(func() int {
//	    return count.Get() * 2
//	})
//
//	// Chained computed values
//	quadrupled := NewComputed(func() int {
//	    return doubled.Get() * 2
//	})
func NewComputed[T any](fn func() T) *Computed[T] {
	// Validate compute function is not nil
	if fn == nil {
		panic(ErrNilComputeFn)
	}

	return &Computed[T]{
		fn:    fn,
		dirty: true, // Starts dirty to trigger initial computation
	}
}

// Get returns the computed value. On the first call, it evaluates the computation function
// and caches the result. Subsequent calls return the cached value without re-evaluating
// the function.
//
// During evaluation, Get automatically tracks all dependencies (Refs and other Computed values)
// accessed by the computation function. When any dependency changes, the cache is invalidated.
//
// This operation is thread-safe and uses a read-write lock. Multiple goroutines can safely
// call Get() concurrently. The computation function is guaranteed to be called at most once
// per invalidation, even under concurrent access.
//
// Example:
//
//	count := NewRef(5)
//	computed := NewComputed(func() int { return count.Get() * 2 })
//	value := computed.Get()  // Evaluates function, tracks count, returns 10
//	value2 := computed.Get() // Returns cached 10 (no re-evaluation)
//	count.Set(10)            // Invalidates cache
//	value3 := computed.Get() // Re-evaluates, returns 20
func (c *Computed[T]) Get() T {
	// Track this Computed as a dependency if tracking is active
	if globalTracker.IsTracking() {
		globalTracker.Track(c)
	}

	// Fast path: check if cache is valid with read lock
	c.mu.RLock()
	if !c.dirty {
		val := c.cache
		c.mu.RUnlock()
		return val
	}
	c.mu.RUnlock()

	// Slow path: need to compute value
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check dirty flag in case another goroutine computed it
	if !c.dirty {
		return c.cache
	}

	// Begin tracking dependencies for this computed value
	err := globalTracker.BeginTracking(c)
	if err != nil {
		// Panic on circular dependency or max depth exceeded
		// This is a programming error that should be caught during development
		panic(err)
	}

	// Evaluate function (will track accessed Refs/Computed values)
	result := c.fn()

	// End tracking and get collected dependencies
	deps := globalTracker.EndTracking()

	// Register this computed value with its dependencies
	for _, dep := range deps {
		dep.AddDependent(c)
	}

	// Update cache and dependencies
	c.cache = result
	c.dirty = false
	c.deps = deps

	return result
}

// Invalidate marks this computed value as dirty, requiring recomputation on next Get().
// It also recursively invalidates all dependents (other computed values that depend on this one).
// Implements the Dependency interface.
func (c *Computed[T]) Invalidate() {
	c.mu.Lock()
	c.dirty = true
	deps := make([]Dependency, len(c.dependents))
	copy(deps, c.dependents)
	c.mu.Unlock()

	// Invalidate all dependents outside the lock
	for _, dep := range deps {
		dep.Invalidate()
	}
}

// AddDependent registers another computed value that depends on this one.
// When this computed value is invalidated, all dependents will also be invalidated.
// Implements the Dependency interface.
func (c *Computed[T]) AddDependent(dep Dependency) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Avoid duplicate dependents
	for _, d := range c.dependents {
		if d == dep {
			return
		}
	}

	c.dependents = append(c.dependents, dep)
}
