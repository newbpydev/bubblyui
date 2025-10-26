package bubbly

import "sync"

// Computed is a type-safe computed value that lazily evaluates a function and caches the result.
// It provides thread-safe read operations using a read-write mutex.
// The computation function is only called once on the first Get() call, and the result is cached
// for all subsequent calls.
//
// Computed values are read-only and do not support direct modification. To create derived state
// that updates when dependencies change, use Computed in combination with Ref values and the
// dependency tracking system (available in later tasks).
//
// Example usage:
//
//	count := bubbly.NewRef(5)
//	doubled := bubbly.NewComputed(func() int {
//	    return count.Get() * 2
//	})
//	value := doubled.Get()  // Computes and caches: 10
//	value2 := doubled.Get() // Returns cached value: 10
type Computed[T any] struct {
	mu    sync.RWMutex
	fn    func() T
	cache T
	dirty bool
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
	return &Computed[T]{
		fn:    fn,
		dirty: true, // Starts dirty to trigger initial computation
	}
}

// Get returns the computed value. On the first call, it evaluates the computation function
// and caches the result. Subsequent calls return the cached value without re-evaluating
// the function.
//
// This operation is thread-safe and uses a read-write lock. Multiple goroutines can safely
// call Get() concurrently. The computation function is guaranteed to be called at most once,
// even under concurrent access.
//
// In future iterations (Task 2.2+), Get will participate in dependency tracking and the
// cache will be invalidated when dependencies change.
//
// Example:
//
//	computed := NewComputed(func() int { return 42 })
//	value := computed.Get()  // Evaluates function, returns 42
//	value2 := computed.Get() // Returns cached 42 (no re-evaluation)
func (c *Computed[T]) Get() T {
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

	// Evaluate function and cache result
	result := c.fn()
	c.cache = result
	c.dirty = false

	return result
}
