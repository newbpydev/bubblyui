package bubbly

// Watchable is an interface for reactive values that can be watched.
// Both Ref[T] and Computed[T] implement this interface, allowing Watch()
// to accept either type.
//
// This design follows Vue 3's approach where computed values are just
// special refs that can be watched directly.
//
// Note: Uses GetTyped() for type-safe value access. The Get() method
// (which returns any) is part of the Dependency interface for polymorphic usage.
//
// Example:
//
//	count := NewRef(5)
//	doubled := NewComputed(func() int { return count.GetTyped() * 2 })
//
//	// Both work with Watch()
//	Watch(count, callback)    // Watch a Ref
//	Watch(doubled, callback)  // Watch a Computed
type Watchable[T any] interface {
	// GetTyped returns the current value with full type safety
	GetTyped() T
	// addWatcher registers a watcher (internal method)
	addWatcher(w *watcher[T])
	// removeWatcher unregisters a watcher (internal method)
	removeWatcher(w *watcher[T])
}

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
	// When enabled, uses reflection-based deep comparison (reflect.DeepEqual) to detect
	// changes in nested fields, slice elements, and map values.
	//
	// Performance note: Deep watching is 10-100x slower than shallow watching.
	// For performance-critical paths, use DeepCompare with a custom comparator.
	Deep bool

	// DeepCompare is an optional custom comparison function for deep watching.
	// If provided, it overrides the default reflect.DeepEqual behavior.
	// The function should return true if the values are considered equal.
	//
	// This is stored as interface{} to avoid type parameters in WatchOptions,
	// but will be type-asserted to DeepCompareFunc[T] when used.
	DeepCompare interface{}

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

// Watch creates a watcher that executes the callback whenever the source value changes.
// It accepts any Watchable[T] source, including Ref[T] and Computed[T].
// It returns a cleanup function that should be called to stop watching and prevent memory leaks.
//
// The callback receives both the new value and the old value, allowing you to react to changes
// and compare states. The callback is executed synchronously after the value is set.
//
// Multiple watchers can be registered on the same source, and they will all be notified
// independently when the value changes.
//
// Type parameter T must match the type of the source being watched, ensuring compile-time
// type safety for the callback parameters.
//
// Example with Ref:
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
// Example with Computed:
//
//	count := NewRef(5)
//	doubled := NewComputed(func() int { return count.GetTyped() * 2 })
//	cleanup := Watch(doubled, func(newVal, oldVal int) {
//	    fmt.Printf("Doubled changed: %d → %d\n", oldVal, newVal)
//	})
//	defer cleanup()
//
//	count.Set(10)  // Prints: Doubled changed: 10 → 20
//
// Options:
//
//	Watch(source, callback, WithImmediate())  // Execute callback immediately
//	Watch(source, callback, WithDeep())       // Watch nested changes
//	Watch(source, callback, WithFlush("post")) // Defer callback execution
//	Watch(source, callback, WithImmediate(), WithFlush("sync")) // Combine options
func Watch[T any](
	source Watchable[T],
	callback WatchCallback[T],
	options ...WatchOption,
) WatchCleanup {
	// Validate callback is not nil
	if callback == nil {
		panic(ErrNilCallback)
	}

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

	// If deep watching is enabled, initialize prevValue with current value
	if opts.Deep {
		currentVal := source.GetTyped()
		w.prevValue = &currentVal
	}

	// Register watcher with the Ref
	source.addWatcher(w)

	// If Immediate option is set, execute callback with current value
	if opts.Immediate {
		currentVal := source.GetTyped()
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
// When enabled, the watcher uses reflection-based deep comparison (reflect.DeepEqual)
// to determine if the value has actually changed. This allows detecting changes in:
//   - Nested struct fields
//   - Slice elements
//   - Map values
//   - Pointer-referenced values
//
// Performance Impact:
//
//	Deep watching is 10-100x slower than shallow watching due to reflection overhead.
//	For performance-critical paths, use WithDeepCompare() with a custom comparator.
//
// Example:
//
//	type User struct {
//	    Name    string
//	    Profile Profile
//	}
//
//	user := NewRef(User{Name: "John", Profile: Profile{Age: 30}})
//	Watch(user, func(newVal, oldVal User) {
//	    fmt.Println("User changed")
//	}, WithDeep())
//
//	// Without deep watching, this would trigger callback even if values are same
//	user.Set(User{Name: "John", Profile: Profile{Age: 30}})  // No callback (deep equal)
//
//	// With actual change
//	user.Set(User{Name: "John", Profile: Profile{Age: 31}})  // Callback triggered
func WithDeep() WatchOption {
	return func(opts *WatchOptions) {
		opts.Deep = true
	}
}

// WithDeepCompare returns a WatchOption that enables deep watching with a custom comparator.
//
// This allows you to define custom equality logic for performance-critical paths.
// The comparator function should return true if the values are considered equal.
//
// Performance:
//
//	Custom comparators can be as fast as shallow watching if you only compare
//	the fields that matter to your application.
//
// Example:
//
//	type User struct {
//	    ID      int
//	    Name    string
//	    Profile Profile  // Large nested struct
//	}
//
//	// Only compare ID and Name, ignore Profile for performance
//	compareUsers := func(old, new User) bool {
//	    return old.ID == new.ID && old.Name == new.Name
//	}
//
//	user := NewRef(User{ID: 1, Name: "John"})
//	Watch(user, func(newVal, oldVal User) {
//	    fmt.Println("User changed")
//	}, WithDeepCompare(compareUsers))
//
//	// This won't trigger callback (ID and Name are same)
//	user.Set(User{ID: 1, Name: "John", Profile: Profile{Age: 31}})
//
//	// This will trigger callback (Name changed)
//	user.Set(User{ID: 1, Name: "Jane", Profile: Profile{Age: 31}})
func WithDeepCompare[T any](compareFn DeepCompareFunc[T]) WatchOption {
	return func(opts *WatchOptions) {
		opts.Deep = true
		opts.DeepCompare = compareFn
	}
}

// WithFlush returns a WatchOption that controls when the callback is executed.
//
// Valid flush modes:
//   - "sync" (default): Execute callback immediately when value changes
//   - "post": Queue callback for later execution via FlushWatchers()
//
// Post-Flush Mode:
//
// When using "post" mode, callbacks are queued instead of executed immediately.
// This allows batching multiple rapid changes into a single callback execution.
// You must call FlushWatchers() to execute all queued callbacks.
//
// Batching Behavior:
//
// If the same watcher is triggered multiple times before FlushWatchers() is called,
// only the last callback (with the final values) will be executed. This prevents
// redundant callback executions and improves performance.
//
// Integration with Bubbletea:
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case someMsg:
//	        m.count.Set(m.count.GetTyped() + 1)  // Queued if using WithFlush("post")
//	        m.count.Set(m.count.GetTyped() + 1)  // Replaces previous (batching)
//	    }
//
//	    // Execute all queued callbacks before returning
//	    FlushWatchers()
//
//	    return m, nil
//	}
//
// Example:
//
//	count := NewRef(0)
//
//	// Sync mode (default) - executes immediately
//	Watch(count, func(newVal, oldVal int) {
//	    fmt.Println("Sync:", newVal)
//	}, WithFlush("sync"))
//
//	// Post mode - queues for later
//	Watch(count, func(newVal, oldVal int) {
//	    fmt.Println("Post:", newVal)
//	}, WithFlush("post"))
//
//	count.Set(1)  // Sync callback runs, post callback queued
//	count.Set(2)  // Sync callback runs, post callback replaced in queue
//	count.Set(3)  // Sync callback runs, post callback replaced in queue
//
//	FlushWatchers()  // Post callback runs once with final value (3)
func WithFlush(mode string) WatchOption {
	return func(opts *WatchOptions) {
		opts.Flush = mode
	}
}
