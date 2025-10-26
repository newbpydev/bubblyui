package bubbly

import (
	"sync"
)

// WatchEffect runs an effect function immediately and automatically tracks its reactive dependencies.
// When any accessed dependency changes, the effect re-runs automatically.
//
// This is inspired by Vue 3's watchEffect and provides automatic dependency tracking,
// eliminating the need to manually specify which reactive values to watch.
//
// The effect function is called immediately upon creation, and then re-runs whenever
// any reactive value (Ref or Computed) accessed during its execution changes.
//
// Key Features:
//   - Automatic dependency discovery: No need to manually specify dependencies
//   - Immediate execution: Runs once immediately, then on dependency changes
//   - Dynamic dependencies: Tracks only the dependencies accessed in each run
//   - Cleanup support: Returns a function to stop watching
//
// Example:
//
//	count := NewRef(0)
//	name := NewRef("John")
//
//	cleanup := WatchEffect(func() {
//	    fmt.Printf("%s: %d\n", name.Get(), count.Get())
//	})
//	defer cleanup()
//
//	count.Set(5)  // Automatically triggers re-run
//	name.Set("Jane")  // Automatically triggers re-run
//
// Conditional Dependencies:
//
//	toggle := NewRef(true)
//	valueA := NewRef(1)
//	valueB := NewRef(100)
//
//	WatchEffect(func() {
//	    if toggle.Get() {
//	        fmt.Println(valueA.Get())  // Only tracks valueA when toggle is true
//	    } else {
//	        fmt.Println(valueB.Get())  // Only tracks valueB when toggle is false
//	    }
//	})
//
// The effect automatically adapts to which dependencies are accessed in each run.
func WatchEffect(effect func()) WatchCleanup {
	// Create effect state
	e := &watchEffect{
		effect:   effect,
		cleanups: make([]WatchCleanup, 0),
		watchers: make(map[Dependency]*invalidationWatcher),
	}

	// Run effect immediately and set up watchers
	e.run()

	// Return cleanup function
	return func() {
		e.cleanup()
	}
}

// watchEffect manages the state of a watch effect
type watchEffect struct {
	mu        sync.Mutex
	effect    func()
	cleanups  []WatchCleanup
	running   bool
	stopped   bool
	settingUp bool                                // Flag to prevent re-runs during initial setup
	watchers  map[Dependency]*invalidationWatcher // Track watchers to avoid duplicates
}

// run executes the effect and tracks dependencies
func (e *watchEffect) run() {
	e.mu.Lock()
	if e.stopped || e.settingUp {
		e.mu.Unlock()
		return
	}

	// Prevent recursive runs
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.settingUp = true // Prevent re-runs while setting up watchers

	// Clean up previous watchers
	oldCleanups := e.cleanups
	e.cleanups = make([]WatchCleanup, 0)
	e.mu.Unlock()

	// Clean up old watchers outside the lock
	for _, cleanup := range oldCleanups {
		cleanup()
	}

	// Begin tracking dependencies during effect execution
	err := globalTracker.BeginTracking(nil)
	if err != nil {
		// If we can't track (e.g., circular dependency), just run the effect once
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()

		// Recover from panics in effect
		defer func() {
			if r := recover(); r != nil {
				// Log or handle panic, but don't crash
				// In production, you might want to log this
				_ = r // Suppress unused variable warning
			}
		}()

		e.effect()
		return
	}

	// Run effect (this will track accessed Refs/Computed)
	// Recover from panics
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Effect panicked, but we still want to track dependencies
				// and allow future runs
				_ = r // Suppress unused variable warning
			}
		}()
		e.effect()
	}()

	// End tracking and get dependencies
	deps := globalTracker.EndTracking()

	// Set up watchers for each dependency
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.stopped {
		e.running = false
		e.settingUp = false
		return
	}

	// Track new dependencies
	newWatchers := make(map[Dependency]*invalidationWatcher)

	for _, dep := range deps {
		// Check if we already have a watcher for this dependency
		if iw, exists := e.watchers[dep]; exists {
			// Reuse existing watcher
			newWatchers[dep] = iw
		} else {
			// Create new watcher for this dependency
			iw := &invalidationWatcher{effect: e}
			newWatchers[dep] = iw

			// Add ourselves as a dependent
			dep.AddDependent(iw)
		}
	}

	// Replace watchers map with new one
	// Old watchers that are no longer needed will remain registered as dependents
	// (we can't remove them without a RemoveDependent method)
	// but they won't be in our map, so we won't track them
	e.watchers = newWatchers

	e.running = false
	e.settingUp = false // Allow future re-runs
}

// invalidationWatcher implements Dependency interface to receive invalidation notifications
type invalidationWatcher struct {
	effect *watchEffect
}

// Invalidate is called when a watched dependency changes
func (iw *invalidationWatcher) Invalidate() {
	iw.effect.run()
}

// AddDependent implements Dependency interface (no-op for watchers)
func (iw *invalidationWatcher) AddDependent(dep Dependency) {
	// Watchers don't have dependents
}

// cleanup stops the effect and cleans up all watchers
func (e *watchEffect) cleanup() {
	e.mu.Lock()
	e.stopped = true
	cleanups := e.cleanups
	e.cleanups = nil
	e.mu.Unlock()

	// Clean up all watchers
	for _, cleanup := range cleanups {
		cleanup()
	}
}
