package bubbly

import (
	"reflect"
	"sync"
)

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
// Computed also implements the Watchable interface (Task 6.2), allowing watchers to be
// registered that will be notified when the computed value changes.
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
//
//	// Watch computed value changes (Task 6.2)
//	Watch(doubled, func(newVal, oldVal int) {
//	    fmt.Printf("Doubled changed: %d → %d\n", oldVal, newVal)
//	})
type Computed[T any] struct {
	mu         sync.RWMutex
	fn         func() T
	cache      T
	dirty      bool
	deps       []Dependency
	dependents []Dependency
	watchers   []*watcher[T] // Task 6.2: Support watching computed values
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

// Get returns the current value as any, implementing the Dependency interface.
// This allows Computed to be used polymorphically with other reactive types.
// For type-safe access, use GetTyped() instead.
//
// This operation is thread-safe. If the cache is dirty, it triggers recomputation.
//
// Example:
//
//	count := NewRef(5)
//	computed := NewComputed(func() int { return count.GetTyped() * 2 })
//	value := computed.Get().(int)  // Returns 10, requires type assertion
func (c *Computed[T]) Get() any {
	return c.GetTyped()
}

// GetTyped returns the current value with full type safety.
// This is the preferred method for direct access when the type is known.
// Use Get() any when working with the Dependency interface.
//
// If the cache is dirty (dependencies have changed), it re-evaluates the computation function.
// Otherwise, it returns the cached value without re-evaluation.
//
// During evaluation, GetTyped automatically tracks all dependencies (Refs and other Computed values)
// accessed by the computation function. When any dependency changes, the cache is invalidated.
//
// This operation is thread-safe and uses a read-write lock. Multiple goroutines can safely
// call GetTyped() concurrently. The computation function is guaranteed to be called at most once
// per invalidation, even under concurrent access.
//
// Task 6.2: When the computed value changes after recomputation, all registered watchers
// are notified with the new and old values.
//
// Example:
//
//	count := NewRef(5)
//	computed := NewComputed(func() int { return count.GetTyped() * 2 })
//	value := computed.GetTyped()  // Evaluates function, tracks count, returns 10
//	value2 := computed.GetTyped() // Returns cached 10 (no re-evaluation)
//	count.Set(10)                 // Invalidates cache
//	value3 := computed.GetTyped() // Re-evaluates, returns 20, notifies watchers
func (c *Computed[T]) GetTyped() T {
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

	// Double-check dirty flag in case another goroutine computed it
	if !c.dirty {
		val := c.cache
		c.mu.Unlock()
		return val
	}

	// Store old value for watcher notification (Task 6.2)
	oldValue := c.cache

	// Begin tracking dependencies for this computed value
	err := globalTracker.BeginTracking(c)
	if err != nil {
		c.mu.Unlock()
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

	// Check if we have watchers before unlocking
	hasWatchers := len(c.watchers) > 0
	c.mu.Unlock()

	// Task 6.2: Notify watchers if value changed
	// Only notify if there are watchers and value actually changed
	// Use reflect.DeepEqual for comparison to handle all types correctly
	if hasWatchers && !reflect.DeepEqual(oldValue, result) {
		c.notifyWatchers(result, oldValue)
	}

	return result
}

// Invalidate marks this computed value as dirty, requiring recomputation on next Get().
// It also recursively invalidates all dependents (other computed values that depend on this one).
// Implements the Dependency interface.
//
// Task 6.2: If this computed value has watchers, it triggers immediate recomputation
// to notify watchers of the value change. This ensures watchers are notified even if
// no one explicitly calls Get() after invalidation.
func (c *Computed[T]) Invalidate() {
	c.mu.Lock()
	c.dirty = true
	hasWatchers := len(c.watchers) > 0
	deps := make([]Dependency, len(c.dependents))
	copy(deps, c.dependents)
	c.mu.Unlock()

	// Invalidate all dependents outside the lock
	for _, dep := range deps {
		dep.Invalidate()
	}

	// Task 6.2: If we have watchers, trigger recomputation to notify them
	// This ensures watchers are called even if no one explicitly calls Get()
	if hasWatchers {
		c.Get()
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

// ============================================================================
// Task 6.2: Watchable Interface Implementation
// ============================================================================

// addWatcher registers a new watcher to be notified of computed value changes.
// This is an internal method used by the public Watch function.
// Implements the Watchable interface.
//
// Task 6.2: When the first watcher is added, the computed value is evaluated
// to establish dependency relationships. This ensures that when dependencies
// change, the computed will be notified and can trigger its watchers.
//
//nolint:unused // Interface method - used via Watchable interface
func (c *Computed[T]) addWatcher(w *watcher[T]) {
	c.mu.Lock()
	isFirstWatcher := len(c.watchers) == 0
	needsInitialEval := c.dirty // Check if never evaluated

	// Preallocate watchers slice with small initial capacity on first watcher
	if c.watchers == nil {
		c.watchers = make([]*watcher[T], 0, 4) // Most computed have 1-4 watchers
	}

	c.watchers = append(c.watchers, w)
	c.mu.Unlock()

	// Task 6.2: Evaluate computed value on first watcher to establish dependencies
	// This ensures the computed registers itself with its dependencies (Refs/Computed)
	// so that when they change, this computed will be invalidated and can notify watchers
	// Only evaluate if this is the first watcher AND the computed has never been evaluated
	if isFirstWatcher && needsInitialEval {
		// Temporarily remove watchers to prevent notification during initial evaluation
		c.mu.Lock()
		savedWatchers := c.watchers
		c.watchers = nil
		c.mu.Unlock()

		// Evaluate to establish dependencies
		c.Get()

		// Restore watchers
		c.mu.Lock()
		c.watchers = savedWatchers
		c.mu.Unlock()
	}
}

// removeWatcher unregisters a watcher so it no longer receives notifications.
// This is an internal method used by the cleanup function returned by Watch.
// Removing a non-existent watcher is safe and does nothing.
// Implements the Watchable interface.
//
//nolint:unused // Interface method - used via Watchable interface
func (c *Computed[T]) removeWatcher(w *watcher[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find and remove the watcher using pointer comparison
	for i, watcher := range c.watchers {
		if watcher == w {
			// Remove by replacing with last element and truncating
			c.watchers[i] = c.watchers[len(c.watchers)-1]
			c.watchers = c.watchers[:len(c.watchers)-1]
			return
		}
	}
}

// notifyWatchers calls all watcher callbacks with the new and old values.
// This method is called when the computed value changes after recomputation.
// It handles deep watching and flush modes by delegating to the shared notifyWatcher helper.
func (c *Computed[T]) notifyWatchers(newVal, oldVal T) {
	c.mu.RLock()
	// Copy watchers slice while holding the read lock
	var watchersCopy []*watcher[T]
	if len(c.watchers) > 0 {
		watchersCopy = make([]*watcher[T], len(c.watchers))
		copy(watchersCopy, c.watchers)
	}
	c.mu.RUnlock()

	// Notify watchers outside the lock (only if there are watchers)
	if len(watchersCopy) == 0 {
		return
	}

	for _, w := range watchersCopy {
		notifyWatcher(w, newVal, oldVal)
	}
}
