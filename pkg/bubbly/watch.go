package bubbly

// WatchCallback is a function that is called when a watched value changes.
// It receives both the new value and the old value as parameters.
//
// Example:
//
//	callback := func(newVal, oldVal int) {
//	    fmt.Printf("Value changed from %d to %d\n", oldVal, newVal)
//	}
type WatchCallback[T any] func(newVal, oldVal T)

// WatchCleanup is a function that stops watching when called.
// It should be called when the watcher is no longer needed to prevent memory leaks.
//
// Example:
//
//	cleanup := Watch(ref, callback)
//	defer cleanup()  // Stop watching when done
type WatchCleanup func()

// WatchOption configures watcher behavior.
// This is a function type that modifies WatchOptions.
// Reserved for Task 3.2 where options like WithImmediate() and WithDeep() will be implemented.
type WatchOption func(*WatchOptions)

// Watch creates a watcher that executes the callback whenever the source Ref's value changes.
// It returns a cleanup function that should be called to stop watching and prevent memory leaks.
//
// The callback receives both the new value and the old value, allowing you to react to changes
// and compare states. The callback is executed synchronously after the value is set.
//
// Multiple watchers can be registered on the same Ref, and they will all be notified
// independently when the value changes.
//
// Type parameter T must match the type of the Ref being watched, ensuring compile-time
// type safety for the callback parameters.
//
// Example:
//
//	count := NewRef(0)
//	cleanup := Watch(count, func(newVal, oldVal int) {
//	    fmt.Printf("Count changed: %d → %d\n", oldVal, newVal)
//	})
//	defer cleanup()  // Stop watching when done
//
//	count.Set(5)   // Prints: Count changed: 0 → 5
//	count.Set(10)  // Prints: Count changed: 5 → 10
//
// Options (reserved for Task 3.2):
//
//	Watch(ref, callback, WithImmediate())  // Execute callback immediately
//	Watch(ref, callback, WithDeep())       // Watch nested changes
func Watch[T any](
	source *Ref[T],
	callback WatchCallback[T],
	options ...WatchOption,
) WatchCleanup {
	// Create internal watcher
	w := &watcher[T]{
		callback: callback,
		options:  WatchOptions{}, // Options will be populated in Task 3.2
	}

	// Register watcher with the Ref
	source.addWatcher(w)

	// Return cleanup function that removes the watcher
	return func() {
		source.removeWatcher(w)
	}
}
