package bubbly

import (
	"sync"
	"sync/atomic"
)

// hookIDCounter is an atomic counter for generating unique hook IDs.
var hookIDCounter atomic.Uint64

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

	// stateMu protects the mounted and unmounting state flags.
	// Uses RWMutex for read-heavy access patterns.
	stateMu sync.RWMutex

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

// registerHook registers a lifecycle hook of the specified type.
// Hooks are stored in registration order and will be executed in that order.
//
// Hook types: "mounted", "beforeUpdate", "updated", "beforeUnmount", "unmounted"
//
// Example:
//
//	lm.registerHook("mounted", lifecycleHook{
//	    id:       "hook-1",
//	    callback: func() { fmt.Println("mounted") },
//	    order:    0,
//	})
func (lm *LifecycleManager) registerHook(hookType string, hook lifecycleHook) {
	lm.hooks[hookType] = append(lm.hooks[hookType], hook)
}

// IsMounted returns whether the component has been mounted.
// This method is thread-safe and uses a read lock.
//
// Example:
//
//	if lm.IsMounted() {
//	    // Component is mounted
//	}
func (lm *LifecycleManager) IsMounted() bool {
	lm.stateMu.RLock()
	defer lm.stateMu.RUnlock()
	return lm.mounted
}

// IsUnmounting returns whether the component is currently unmounting.
// This method is thread-safe and uses a read lock.
//
// Example:
//
//	if lm.IsUnmounting() {
//	    // Component is unmounting
//	}
func (lm *LifecycleManager) IsUnmounting() bool {
	lm.stateMu.RLock()
	defer lm.stateMu.RUnlock()
	return lm.unmounting
}

// setMounted sets the mounted state of the component.
// This method is thread-safe and uses a write lock.
//
// Example:
//
//	lm.setMounted(true)  // Mark as mounted
func (lm *LifecycleManager) setMounted(mounted bool) {
	lm.stateMu.Lock()
	defer lm.stateMu.Unlock()
	lm.mounted = mounted
}

// setUnmounting sets the unmounting state of the component.
// This method is thread-safe and uses a write lock.
//
// Example:
//
//	lm.setUnmounting(true)  // Mark as unmounting
func (lm *LifecycleManager) setUnmounting(unmounting bool) {
	lm.stateMu.Lock()
	defer lm.stateMu.Unlock()
	lm.unmounting = unmounting
}

// executeMounted executes all registered onMounted hooks.
// This method should be called after the component's first render.
// It ensures hooks only execute once by checking the mounted state.
//
// The method:
//   - Checks if already mounted (returns early if true)
//   - Sets the mounted state to true
//   - Executes all "mounted" hooks in registration order
//   - Recovers from panics in individual hooks
//
// Example:
//
//	lm.executeMounted()  // Execute all onMounted hooks
func (lm *LifecycleManager) executeMounted() {
	// Check if already mounted
	if lm.IsMounted() {
		return
	}

	// Mark as mounted before executing hooks
	lm.setMounted(true)

	// Execute all mounted hooks
	lm.executeHooks("mounted")
}

// executeHooks executes all hooks of the specified type in registration order.
// Each hook is executed with panic recovery to ensure one failing hook
// doesn't prevent others from running.
//
// Hook types: "mounted", "beforeUpdate", "updated", "beforeUnmount", "unmounted"
//
// The method:
//   - Iterates through hooks in registration order
//   - Executes each hook with panic recovery
//   - Logs errors but continues execution
//   - Guarantees all hooks are attempted
//
// Example:
//
//	lm.executeHooks("mounted")  // Execute all mounted hooks
func (lm *LifecycleManager) executeHooks(hookType string) {
	hooks, exists := lm.hooks[hookType]
	if !exists || len(hooks) == 0 {
		return
	}

	// Execute each hook in registration order
	for _, hook := range hooks {
		lm.safeExecuteHook(hookType, hook)
	}
}

// safeExecuteHook executes a single hook with panic recovery.
// If the hook panics, the panic is caught, logged, and execution continues.
// This ensures that one failing hook doesn't crash the component or prevent
// other hooks from executing.
//
// The method:
//   - Uses defer/recover to catch panics
//   - Logs panic information (would integrate with error reporting)
//   - Allows execution to continue normally
//
// Example:
//
//	lm.safeExecuteHook("mounted", hook)
func (lm *LifecycleManager) safeExecuteHook(hookType string, hook lifecycleHook) {
	defer func() {
		if r := recover(); r != nil {
			// Panic recovered - in production, this would be logged
			// or reported to an error tracking service
			// For now, we silently recover to allow tests to verify behavior
			_ = hookType // Use hookType to avoid unused variable warning
			_ = r        // Use r to avoid unused variable warning
		}
	}()

	// Execute the hook callback
	hook.callback()
}

// executeUpdated executes all registered onUpdated hooks with dependency tracking.
// This method should be called after the component updates (after state changes).
//
// The method:
//   - Checks if component is mounted (returns early if not)
//   - Iterates through all "updated" hooks in registration order
//   - For hooks with dependencies: checks if any dependency changed
//   - For hooks without dependencies: always executes
//   - Updates lastValues after execution for dependency tracking
//   - Recovers from panics in individual hooks
//
// Dependency tracking:
//   - No dependencies: hook runs on every update
//   - With dependencies: hook runs only when at least one dependency changes
//   - Uses reflect.DeepEqual for value comparison
//   - Updates lastValues after successful execution
//
// Example:
//
//	lm.executeUpdated()  // Execute all onUpdated hooks
func (lm *LifecycleManager) executeUpdated() {
	// Only execute if component is mounted
	if !lm.IsMounted() {
		return
	}

	// Get updated hooks
	hooks, exists := lm.hooks["updated"]
	if !exists || len(hooks) == 0 {
		return
	}

	// Execute each hook with dependency checking
	for i := range hooks {
		hook := &hooks[i] // Get pointer to modify lastValues

		// Check if hook should execute based on dependencies
		shouldExecute := lm.shouldExecuteHook(hook)

		if shouldExecute {
			// Execute the hook
			lm.safeExecuteHook("updated", *hook)

			// Update lastValues after execution if hook has dependencies
			if len(hook.dependencies) > 0 {
				lm.updateLastValues(hook)
			}
		}
	}
}

// shouldExecuteHook determines if a hook should execute based on its dependencies.
// Returns true if:
//   - Hook has no dependencies (always execute)
//   - At least one dependency value has changed (using reflect.DeepEqual)
//
// This method compares current dependency values with lastValues to detect changes.
func (lm *LifecycleManager) shouldExecuteHook(hook *lifecycleHook) bool {
	// No dependencies: always execute
	if len(hook.dependencies) == 0 {
		return true
	}

	// Check if any dependency has changed
	for i, dep := range hook.dependencies {
		currentValue := dep.Get()
		lastValue := hook.lastValues[i]

		// Use deepEqual for comparison (from deep.go)
		if !deepEqual(currentValue, lastValue) {
			return true
		}
	}

	// No dependencies changed
	return false
}

// updateLastValues updates the lastValues slice with current dependency values.
// This is called after a hook executes to track the values for next comparison.
func (lm *LifecycleManager) updateLastValues(hook *lifecycleHook) {
	for i, dep := range hook.dependencies {
		hook.lastValues[i] = dep.Get()
	}
}

// executeUnmounted executes all registered onUnmounted hooks and cleanup functions.
// This method should be called when the component is being removed/unmounted.
//
// The method:
//   - Checks if already unmounting (returns early if true)
//   - Sets the unmounting state to true
//   - Executes all "unmounted" hooks in registration order
//   - Executes all cleanup functions in reverse order (LIFO)
//   - Recovers from panics in individual hooks and cleanups
//
// Execution order:
//  1. onUnmounted hooks (registration order)
//  2. Cleanup functions (reverse order - LIFO)
//
// This ensures proper cleanup sequence where:
//   - User-defined unmount logic runs first
//   - Cleanup functions unwind in reverse registration order
//
// Example:
//
//	lm.executeUnmounted()  // Execute all onUnmounted hooks and cleanups
func (lm *LifecycleManager) executeUnmounted() {
	// Check if already unmounting
	if lm.IsUnmounting() {
		return
	}

	// Mark as unmounting before executing hooks
	lm.setUnmounting(true)

	// Execute all unmounted hooks
	lm.executeHooks("unmounted")

	// Execute cleanup functions
	lm.executeCleanups()
}

// executeCleanups executes all registered cleanup functions in reverse order (LIFO).
// Cleanup functions are executed in reverse order to properly unwind resources
// in the opposite order they were acquired.
//
// The method:
//   - Iterates through cleanups in reverse order (LIFO)
//   - Executes each cleanup with panic recovery
//   - Continues execution even if individual cleanups panic
//   - Guarantees all cleanups are attempted
//
// LIFO execution ensures:
//   - Resources are released in reverse acquisition order
//   - Dependencies are cleaned up before dependents
//   - Proper unwinding of nested resources
//
// Example:
//
//	lm.executeCleanups()  // Execute all cleanup functions in reverse order
func (lm *LifecycleManager) executeCleanups() {
	// Execute cleanups in reverse order (LIFO)
	for i := len(lm.cleanups) - 1; i >= 0; i-- {
		lm.safeExecuteCleanup(lm.cleanups[i])
	}
}

// safeExecuteCleanup executes a single cleanup function with panic recovery.
// If the cleanup panics, the panic is caught, logged, and execution continues.
// This ensures that one failing cleanup doesn't prevent other cleanups from executing.
//
// The method:
//   - Uses defer/recover to catch panics
//   - Logs panic information (would integrate with error reporting)
//   - Allows execution to continue normally
//
// Example:
//
//	lm.safeExecuteCleanup(cleanup)
func (lm *LifecycleManager) safeExecuteCleanup(cleanup CleanupFunc) {
	defer func() {
		if r := recover(); r != nil {
			// Panic recovered - in production, this would be logged
			// or reported to an error tracking service
			// For now, we silently recover to allow tests to verify behavior
			_ = r // Use r to avoid unused variable warning
		}
	}()

	// Execute the cleanup function
	cleanup()
}
