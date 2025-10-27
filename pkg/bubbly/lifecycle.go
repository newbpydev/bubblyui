package bubbly

// CleanupFunc is a function that performs cleanup operations.
// It is called when a component is unmounted to release resources,
// cancel subscriptions, or perform other cleanup tasks.
//
// Example:
//
//	ctx.OnCleanup(func() {
//	    ticker.Stop()
//	    subscription.Unsubscribe()
//	})
type CleanupFunc func()

// lifecycleHook represents a single lifecycle hook registration.
// It stores the callback function, dependencies for change tracking,
// and metadata about the hook.
type lifecycleHook struct {
	// id is a unique identifier for this hook instance
	//nolint:unused // Will be used in Task 1.2 (Hook registration)
	id string

	// callback is the function to execute when the hook fires
	//nolint:unused // Will be used in Task 2.1 (Hook execution)
	callback func()

	// dependencies are the Refs that this hook depends on.
	// The hook only executes if one of these dependencies changes.
	// Empty slice means the hook runs on every trigger.
	//nolint:unused // Will be used in Task 2.2 (Dependency tracking)
	dependencies []*Ref[any]

	// lastValues stores the previous values of dependencies
	// for change detection
	//nolint:unused // Will be used in Task 2.2 (Dependency tracking)
	lastValues []any

	// order is the registration order of this hook.
	// Hooks execute in registration order.
	//nolint:unused // Will be used in Task 2.1 (Hook execution)
	order int
}

// watcherCleanup represents a watcher that needs cleanup on unmount.
// This will be used in Task 4.1 for auto-cleanup integration.
type watcherCleanup struct {
	// cleanup is the function to call to stop watching
	//nolint:unused // Will be used in Task 4.1 (Watcher auto-cleanup)
	cleanup func()
}

// LifecycleManager manages the lifecycle hooks for a component.
// It handles hook registration, execution, and cleanup.
//
// The lifecycle manager is responsible for:
//   - Storing registered hooks by type (mounted, updated, unmounted)
//   - Tracking component lifecycle state (mounted, unmounting)
//   - Executing hooks at appropriate times
//   - Managing cleanup functions and auto-cleanup
//   - Preventing infinite update loops
//
// Lifecycle flow:
//  1. Component created → LifecycleManager created
//  2. Setup() runs → Hooks registered
//  3. First View() → onMounted hooks execute
//  4. State changes → onUpdated hooks execute
//  5. Component unmounts → onUnmounted hooks + cleanup execute
type LifecycleManager struct {
	// component is the component this lifecycle manager belongs to
	component *componentImpl

	// hooks stores registered lifecycle hooks by type.
	// Keys: "mounted", "beforeUpdate", "updated", "beforeUnmount", "unmounted"
	// Values: slices of hooks in registration order
	hooks map[string][]lifecycleHook

	// cleanups stores cleanup functions to execute on unmount.
	// Executed in reverse order (LIFO).
	cleanups []CleanupFunc

	// watchers stores watcher cleanup functions for auto-cleanup.
	// All watchers are automatically cleaned up when component unmounts.
	watchers []watcherCleanup

	// mounted indicates whether the component has been mounted.
	// Set to true after onMounted hooks execute.
	mounted bool

	// unmounting indicates whether the component is currently unmounting.
	// Set to true when unmount process begins.
	unmounting bool

	// updateCount tracks the number of updates to detect infinite loops.
	// Reset periodically to prevent false positives.
	updateCount int
}

// newLifecycleManager creates a new LifecycleManager for the given component.
// It initializes all maps and slices to prevent nil pointer panics.
//
// The lifecycle manager starts in an unmounted state with no registered hooks.
//
// Example:
//
//	lm := newLifecycleManager(component)
//	// lm.mounted == false
//	// lm.hooks is empty but not nil
func newLifecycleManager(c *componentImpl) *LifecycleManager {
	return &LifecycleManager{
		component:   c,
		hooks:       make(map[string][]lifecycleHook),
		cleanups:    []CleanupFunc{},
		watchers:    []watcherCleanup{},
		mounted:     false,
		unmounting:  false,
		updateCount: 0,
	}
}
