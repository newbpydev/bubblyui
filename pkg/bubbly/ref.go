// Package bubbly provides a Vue-inspired reactive state management system for Go TUI applications.
// It offers type-safe reactive primitives built on generics that integrate seamlessly with
// the Bubbletea framework's Elm architecture.
package bubbly

import (
	"fmt"
	"sync"
)

// watcher represents an internal watcher that observes changes to a Ref.
// It is unexported as it's an implementation detail used by the Watch function.
// WatchOptions is defined in watch.go.
type watcher[T any] struct {
	callback  func(newVal, oldVal T)
	options   WatchOptions
	prevValue *T // Stores previous value for deep comparison (nil if not deep watching)
}

// notifyWatcher handles notification logic for a single watcher.
// It performs deep comparison if enabled and handles flush modes.
// This is a shared helper to avoid code duplication between Ref and Computed.
func notifyWatcher[T any](w *watcher[T], newVal, oldVal T) {
	// Task 8.8: Notify framework hook BEFORE callback execution
	// This allows dev tools to track the reactive cascade
	watcherID := fmt.Sprintf("watch-%p", w)
	notifyHookWatchCallback(watcherID, newVal, oldVal)

	shouldNotify := true

	// If deep watching is enabled, check if value actually changed
	if w.options.Deep {
		// Get custom comparator if provided
		var compareFn DeepCompareFunc[T]
		if w.options.DeepCompare != nil {
			if fn, ok := w.options.DeepCompare.(DeepCompareFunc[T]); ok {
				compareFn = fn
			}
		}

		// Compare with previous value if available
		if w.prevValue != nil {
			// Use deep comparison to check if value changed
			shouldNotify = hasChanged(*w.prevValue, newVal, compareFn)
		}

		// Update previous value for next comparison
		prevCopy := newVal
		w.prevValue = &prevCopy
	}

	// Only trigger callback if value changed (or not deep watching)
	if shouldNotify {
		// Check flush mode
		if w.options.Flush == "post" {
			// Queue callback for later execution
			// Capture values in closure to avoid race conditions
			watcher := w
			newValue := newVal
			oldValue := oldVal
			globalScheduler.enqueue(watcher, func() {
				watcher.callback(newValue, oldValue)
			})
		} else {
			// Execute immediately (sync mode)
			w.callback(newVal, oldVal)
		}
	}
}

// Ref is a type-safe reactive reference that holds a mutable value of type T.
// It provides thread-safe read and write operations using a read-write mutex.
// Ref supports watchers that are notified when the value changes.
//
// Ref implements the Dependency interface, allowing it to participate in
// automatic dependency tracking for computed values.
//
// Example usage:
//
//	count := bubbly.NewRef(0)
//	value := count.GetTyped()  // Read current value
//	count.Set(42)         // Update value and notify watchers
type Ref[T any] struct {
	mu              sync.RWMutex
	value           T
	watchers        []*watcher[T]
	dependents      []Dependency
	setHook         func(oldValue, newValue T) // Optional hook called after Set()
	templateChecker func() bool                // Optional checker for template context
}

// NewRef creates a new reactive reference with the given initial value.
// The reference is thread-safe and can be safely accessed from multiple goroutines.
//
// Type parameter T can be any Go type including primitives, structs, slices,
// maps, pointers, and interfaces.
//
// Example:
//
//	intRef := NewRef(42)
//	stringRef := NewRef("hello")
//	structRef := NewRef(User{Name: "John"})
func NewRef[T any](value T) *Ref[T] {
	return &Ref[T]{
		value: value,
	}
}

// Get returns the current value as any, implementing the Dependency interface.
// This allows Ref to be used polymorphically with other reactive types.
// For type-safe access, use GetTyped() instead.
//
// This operation is thread-safe and uses a read lock, allowing multiple
// concurrent readers. When called during computed value evaluation, Get
// automatically registers this Ref as a dependency.
//
// Example:
//
//	ref := NewRef(42)
//	value := ref.GetTyped().(int)  // Returns 42, requires type assertion
func (r *Ref[T]) Get() any {
	return r.GetTyped()
}

// GetTyped returns the current value with full type safety.
// This is the preferred method for direct access when the type is known.
// Use Get() any when working with the Dependency interface.
//
// This operation is thread-safe and uses a read lock, allowing multiple
// concurrent readers. When called during computed value evaluation, GetTyped
// automatically registers this Ref as a dependency.
//
// Example:
//
//	ref := NewRef(42)
//	value := ref.GetTyped()  // Returns 42 as int, no type assertion needed
func (r *Ref[T]) GetTyped() T {
	// Track this Ref as a dependency if tracking is active
	if globalTracker.IsTracking() {
		globalTracker.Track(r)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// Set updates the value of the reference and notifies all registered watchers.
// This operation is thread-safe and uses a write lock.
// Watchers are notified outside the lock to prevent deadlocks.
//
// When the value changes, all dependent computed values are invalidated,
// causing them to recompute on their next Get() call.
//
// If a setHook is registered (for automatic command generation), it will be
// called after the value is updated and watchers are notified.
//
// Template Context Safety:
// If a templateChecker is registered and indicates we're currently inside a template,
// Set() will panic with a clear error message. Templates must be pure functions with
// no side effects - state mutations should only occur in event handlers or lifecycle hooks.
//
// Example:
//
//	ref := NewRef(10)
//	ref.Set(20)  // Updates value to 20 and notifies watchers
func (r *Ref[T]) Set(value T) {
	// Check if we're in a template context (if checker is registered)
	// This must be done BEFORE acquiring the lock to avoid deadlock
	r.mu.RLock()
	checker := r.templateChecker
	r.mu.RUnlock()

	if checker != nil && checker() {
		panic("Cannot call Ref.Set() in template - templates must be pure functions with no side effects. " +
			"Move state updates to event handlers (ctx.On) or lifecycle hooks (onMounted, onUpdated).")
	}

	r.mu.Lock()
	oldValue := r.value
	r.value = value

	// Copy watchers slice and hook while holding the lock
	// Use exact length to avoid over-allocation
	var watchersCopy []*watcher[T]
	if len(r.watchers) > 0 {
		watchersCopy = make([]*watcher[T], len(r.watchers))
		copy(watchersCopy, r.watchers)
	}
	hook := r.setHook
	r.mu.Unlock()

	// Notify framework hooks of ref change (use memory address as ID)
	refID := fmt.Sprintf("ref-%p", r)
	notifyHookRefChange(refID, oldValue, value)

	// Invalidate all dependent computed values
	r.Invalidate()

	// Notify watchers outside the lock (only if there are watchers)
	if len(watchersCopy) > 0 {
		r.notifyWatchers(value, oldValue, watchersCopy)
	}

	// Call setHook if registered (for automatic command generation)
	if hook != nil {
		hook(oldValue, value)
	}
}

// addWatcher registers a new watcher to be notified of value changes.
// This is an internal method used by the public Watch function.
func (r *Ref[T]) addWatcher(w *watcher[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Preallocate watchers slice with small initial capacity on first watcher
	if r.watchers == nil {
		r.watchers = make([]*watcher[T], 0, 4) // Most refs have 1-4 watchers
	}

	r.watchers = append(r.watchers, w)
}

// removeWatcher unregisters a watcher so it no longer receives notifications.
// This is an internal method used by the cleanup function returned by Watch.
// Removing a non-existent watcher is safe and does nothing.
func (r *Ref[T]) removeWatcher(w *watcher[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find and remove the watcher using pointer comparison
	for i, watcher := range r.watchers {
		if watcher == w {
			// Remove by replacing with last element and truncating
			r.watchers[i] = r.watchers[len(r.watchers)-1]
			r.watchers = r.watchers[:len(r.watchers)-1]
			return
		}
	}
}

// notifyWatchers calls all watcher callbacks with the new and old values.
// This method is called after Set() updates the value.
// It handles deep watching and flush modes by delegating to the shared notifyWatcher helper.
func (r *Ref[T]) notifyWatchers(newVal, oldVal T, watchers []*watcher[T]) {
	for _, w := range watchers {
		notifyWatcher(w, newVal, oldVal)
	}
}

// Invalidate marks all dependents (computed values) as needing recomputation.
// This is called when the Ref's value changes.
// Implements the Dependency interface.
func (r *Ref[T]) Invalidate() {
	r.mu.RLock()
	deps := make([]Dependency, len(r.dependents))
	copy(deps, r.dependents)
	r.mu.RUnlock()

	// Invalidate all dependents outside the lock
	for _, dep := range deps {
		dep.Invalidate()
	}
}

// AddDependent registers a computed value that depends on this Ref.
// When the Ref changes, all dependents will be invalidated.
// Implements the Dependency interface.
func (r *Ref[T]) AddDependent(dep Dependency) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Avoid duplicate dependents
	for _, d := range r.dependents {
		if d == dep {
			return
		}
	}

	r.dependents = append(r.dependents, dep)
}
