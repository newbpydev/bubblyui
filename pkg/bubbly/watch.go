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

// WatchOptions configures watcher behavior.
// Options can be combined to customize how watchers respond to value changes.
type WatchOptions struct {
	// Immediate causes the callback to execute immediately with the current value
	// when the watcher is created, before any changes occur.
	Immediate bool

	// Deep enables watching of nested changes in complex structures.
	// ⚠️ PLACEHOLDER: Currently accepted but has no effect on behavior.
	// Full implementation planned in Task 3.3 of the reactivity system spec.
	// When implemented, will use reflection or custom comparator for deep comparison.
	// Until then, watchers only trigger when Set() is explicitly called.
	Deep bool

	// Flush controls when the callback is executed relative to the value change.
	// ⚠️ PARTIAL IMPLEMENTATION: Only "sync" mode is fully functional.
	// Valid values:
	//   - "sync" (default): Execute callback immediately (✅ implemented)
	//   - "post": Defer execution to next tick (⚠️ placeholder, behaves as sync)
	// Full async flush implementation planned in Task 3.4 of the reactivity system spec.
	Flush string
}

// WatchOption is a function that configures WatchOptions.
// Options are applied using the functional options pattern.
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
// Options:
//
//	Watch(ref, callback, WithImmediate())  // Execute callback immediately
//	Watch(ref, callback, WithDeep())       // Watch nested changes (placeholder)
//	Watch(ref, callback, WithFlush("post")) // Defer callback execution
//	Watch(ref, callback, WithImmediate(), WithFlush("sync")) // Combine options
func Watch[T any](
	source *Ref[T],
	callback WatchCallback[T],
	options ...WatchOption,
) WatchCleanup {
	// Build options
	opts := WatchOptions{
		Flush: "sync", // Default to synchronous execution
	}
	for _, opt := range options {
		opt(&opts)
	}

	// Create internal watcher
	w := &watcher[T]{
		callback: callback,
		options:  opts,
	}

	// Register watcher with the Ref
	source.addWatcher(w)

	// If Immediate option is set, execute callback with current value
	if opts.Immediate {
		currentVal := source.Get()
		callback(currentVal, currentVal)
	}

	// Return cleanup function that removes the watcher
	return func() {
		source.removeWatcher(w)
	}
}

// WithImmediate returns a WatchOption that causes the callback to execute immediately
// with the current value when the watcher is created.
//
// This is useful for initializing UI state based on the current value without waiting
// for the first change.
//
// Example:
//
//	count := NewRef(5)
//	Watch(count, func(newVal, oldVal int) {
//	    fmt.Printf("Count: %d\n", newVal)
//	}, WithImmediate())
//	// Immediately prints: Count: 5
func WithImmediate() WatchOption {
	return func(opts *WatchOptions) {
		opts.Immediate = true
	}
}

// WithDeep returns a WatchOption that enables deep watching of nested changes.
//
// ⚠️ PLACEHOLDER: This option is currently accepted but has no effect on behavior.
// Full implementation is planned in Task 3.3 of the reactivity system specification.
//
// Current Behavior:
//   - Watchers trigger only when Set() is explicitly called on the Ref
//   - Nested field changes do NOT trigger watchers automatically
//
// Future Implementation (Task 3.3):
//   - Use reflection-based deep comparison (reflect.DeepEqual)
//   - Support custom comparator functions for performance
//   - Detect nested struct, slice, and map changes
//
// Workaround Until Task 3.3:
//
//	  Always call Set() after modifying nested fields:
//
//		user := NewRef(User{Name: "John", Profile: Profile{Age: 30}})
//		Watch(user, func(newVal, oldVal User) {
//		    fmt.Println("User changed")
//		}, WithDeep())
//
//		// Current: Must call Set() to trigger watcher
//		u := user.Get()
//		u.Profile.Age = 31
//		user.Set(u)  // Required to trigger watcher
func WithDeep() WatchOption {
	return func(opts *WatchOptions) {
		opts.Deep = true
	}
}

// WithFlush returns a WatchOption that controls when the callback is executed.
//
// ⚠️ PARTIAL IMPLEMENTATION: Only "sync" mode is fully functional.
// Full async flush implementation is planned in Task 3.4 of the reactivity system specification.
//
// Valid flush modes:
//   - "sync" (default): Execute callback immediately when value changes (✅ implemented)
//   - "post": Defer callback execution to next tick (⚠️ placeholder, currently behaves as sync)
//
// Current Behavior:
//
//	Both "sync" and "post" modes execute callbacks immediately.
//
// Future Implementation (Task 3.4):
//   - "post" mode will queue callbacks and execute after current operation
//   - Batch multiple rapid changes into single callback execution
//   - Integrate with Bubbletea's event loop for optimal rendering
//   - Reduce redundant callback executions for better performance
//
// Example:
//
//	count := NewRef(0)
//	Watch(count, func(newVal, oldVal int) {
//	    fmt.Println("Changed")
//	}, WithFlush("sync"))  // Executes immediately
//
//	// Future (Task 3.4):
//	Watch(count, func(newVal, oldVal int) {
//	    fmt.Println("Changed")
//	}, WithFlush("post"))  // Will defer to next tick
func WithFlush(mode string) WatchOption {
	return func(opts *WatchOptions) {
		opts.Flush = mode
	}
}
