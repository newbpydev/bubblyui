// Package bubbly provides a Vue-inspired reactive state management system for Go TUI applications.
// It offers type-safe reactive primitives built on generics that integrate seamlessly with
// the Bubbletea framework's Elm architecture.
package bubbly

import "sync"

// Ref is a type-safe reactive reference that holds a mutable value of type T.
// It provides thread-safe read and write operations using a read-write mutex.
// Ref is the foundation of the reactivity system and will support watchers and
// dependency tracking in future iterations.
//
// Example usage:
//
//	count := bubbly.NewRef(0)
//	value := count.Get()  // Read current value
//	count.Set(42)         // Update value
type Ref[T any] struct {
	mu    sync.RWMutex
	value T
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

// Get returns the current value of the reference.
// This operation is thread-safe and uses a read lock, allowing multiple
// concurrent readers.
//
// In future iterations, Get will also participate in dependency tracking
// when called within computed value evaluation.
//
// Example:
//
//	ref := NewRef(42)
//	value := ref.Get()  // Returns 42
func (r *Ref[T]) Get() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// Set updates the value of the reference.
// This operation is thread-safe and uses a write lock.
//
// In future iterations, Set will trigger registered watchers and invalidate
// dependent computed values.
//
// Example:
//
//	ref := NewRef(10)
//	ref.Set(20)  // Updates value to 20
func (r *Ref[T]) Set(value T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.value = value
}
