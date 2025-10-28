package bubbly

import (
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// hookIDCounter is an atomic counter for generating unique hook IDs.
var hookIDCounter atomic.Uint64

// maxUpdateDepth is the maximum number of consecutive update cycles allowed
// before an infinite loop is detected. This prevents runaway updates where
// onUpdated hooks continuously trigger more updates.
//
// The value of 100 is chosen to be high enough for legitimate use cases
// but low enough to catch infinite loops quickly.
const maxUpdateDepth = 100

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
// It stores the cleanup function that will be called when the component unmounts.
type watcherCleanup struct {
	// cleanup is the function to call to stop watching
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
// If the hook panics, the panic is caught, reported to observability, and execution continues.
// This ensures that one failing hook doesn't crash the component or prevent
// other hooks from executing.
//
// The method:
//   - Uses defer/recover to catch panics
//   - Reports panic to observability system (Sentry, console, etc.)
//   - Captures stack trace and context for debugging
//   - Allows execution to continue normally
//
// Example:
//
//	lm.safeExecuteHook("mounted", hook)
func (lm *LifecycleManager) safeExecuteHook(hookType string, hook lifecycleHook) {
	defer func() {
		if r := recover(); r != nil {
			// Report panic to observability system
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: lm.component.name,
					EventName:     fmt.Sprintf("lifecycle:%s", hookType),
					PanicValue:    r,
				}

				ctx := &observability.ErrorContext{
					ComponentName: lm.component.name,
					ComponentID:   lm.component.id,
					EventName:     fmt.Sprintf("lifecycle:%s", hookType),
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"hook_type": hookType,
						"hook_id":   hook.id,
					},
					Extra: map[string]interface{}{
						"hook_order":       hook.order,
						"has_dependencies": len(hook.dependencies) > 0,
						"dependency_count": len(hook.dependencies),
					},
				}

				reporter.ReportPanic(panicErr, ctx)
			}
		}
	}()

	// Execute the hook callback
	hook.callback()
}

// executeUpdated executes all registered onUpdated hooks with dependency tracking.
// This method should be called after the component updates (after state changes).
//
// The method:
//   - Checks for infinite loop (max update depth exceeded)
//   - Checks if component is mounted (returns early if not)
//   - Increments update counter for loop detection
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
// Infinite loop detection:
//   - Tracks update count to detect infinite loops
//   - Returns early with error if max depth (100) exceeded
//   - Reports error to observability system for monitoring
//
// Example:
//
//	lm.executeUpdated()  // Execute all onUpdated hooks
func (lm *LifecycleManager) executeUpdated() {
	// Check for infinite loop (max update depth exceeded)
	if err := lm.checkUpdateDepth(); err != nil {
		// Report error to observability system
		if reporter := observability.GetErrorReporter(); reporter != nil {
			ctx := &observability.ErrorContext{
				ComponentName: lm.component.name,
				ComponentID:   lm.component.id,
				EventName:     "lifecycle:max_update_depth",
				Timestamp:     time.Now(),
				StackTrace:    debug.Stack(),
				Tags: map[string]string{
					"error_type":   "max_update_depth",
					"update_count": fmt.Sprintf("%d", lm.updateCount),
				},
				Extra: map[string]interface{}{
					"max_depth":    maxUpdateDepth,
					"update_count": lm.updateCount,
					"is_mounted":   lm.mounted,
				},
			}

			// Report as a generic error (not a panic)
			reporter.ReportError(err, ctx)
		}
		// Stop execution to prevent infinite loop
		return
	}

	// Only execute if component is mounted
	if !lm.IsMounted() {
		return
	}

	// Increment update count for infinite loop detection
	lm.updateCount++

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

	// Execute watcher cleanups (before manual cleanups)
	lm.cleanupWatchers()

	// Execute manual cleanup functions
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
// If the cleanup panics, the panic is caught, reported to observability, and execution continues.
// This ensures that one failing cleanup doesn't prevent other cleanups from executing.
//
// The method:
//   - Uses defer/recover to catch panics
//   - Reports panic to observability system (Sentry, console, etc.)
//   - Captures stack trace and context for debugging
//   - Allows execution to continue normally
//
// Example:
//
//	lm.safeExecuteCleanup(cleanup)
func (lm *LifecycleManager) safeExecuteCleanup(cleanup CleanupFunc) {
	defer func() {
		if r := recover(); r != nil {
			// Report panic to observability system
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: lm.component.name,
					EventName:     "lifecycle:cleanup",
					PanicValue:    r,
				}

				ctx := &observability.ErrorContext{
					ComponentName: lm.component.name,
					ComponentID:   lm.component.id,
					EventName:     "lifecycle:cleanup",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"hook_type": "cleanup",
					},
					Extra: map[string]interface{}{
						"cleanup_count": len(lm.cleanups),
						"is_unmounting": lm.unmounting,
					},
				}

				reporter.ReportPanic(panicErr, ctx)
			}
		}
	}()

	// Execute the cleanup function
	cleanup()
}

// checkUpdateDepth checks if the update count has exceeded the maximum depth.
// Returns an error if the maximum depth is exceeded, indicating a potential
// infinite loop.
//
// This method is called at the beginning of executeUpdated() to prevent
// runaway updates where onUpdated hooks continuously trigger more updates.
//
// Example scenario that would trigger this:
//
//	ctx.OnUpdated(func() {
//	    count.Set(count.Get() + 1)  // Infinite loop!
//	})
//
// Returns ErrMaxUpdateDepth if the limit is exceeded.
func (lm *LifecycleManager) checkUpdateDepth() error {
	if lm.updateCount > maxUpdateDepth {
		return ErrMaxUpdateDepth
	}
	return nil
}

// resetUpdateCount resets the update counter to zero.
// This should be called periodically to prevent false positives,
// or manually when recovering from an infinite loop error.
//
// Example:
//
//	lm.resetUpdateCount()  // Reset counter
func (lm *LifecycleManager) resetUpdateCount() {
	lm.updateCount = 0
}

// registerWatcher registers a watcher cleanup function for auto-cleanup on unmount.
// The cleanup function will be called when the component unmounts to stop watching
// and prevent memory leaks.
//
// This method is called internally by Context.Watch() to automatically register
// watchers for cleanup.
//
// Example:
//
//	cleanup := Watch(ref, callback)
//	lm.registerWatcher(cleanup)  // Auto-cleanup on unmount
func (lm *LifecycleManager) registerWatcher(cleanup func()) {
	lm.watchers = append(lm.watchers, watcherCleanup{
		cleanup: cleanup,
	})
}

// cleanupWatchers executes all registered watcher cleanup functions.
// This method is called during component unmount to stop all watchers
// and prevent memory leaks.
//
// The method:
//   - Iterates through all registered watcher cleanups
//   - Executes each cleanup with panic recovery
//   - Continues execution even if individual cleanups panic
//   - Guarantees all cleanups are attempted
//
// Panic recovery ensures that one failing watcher cleanup doesn't prevent
// other watchers from being cleaned up.
//
// Example:
//
//	lm.cleanupWatchers()  // Stop all watchers
func (lm *LifecycleManager) cleanupWatchers() {
	// Execute each watcher cleanup
	for _, watcher := range lm.watchers {
		lm.safeExecuteWatcherCleanup(watcher.cleanup)
	}
}

// safeExecuteWatcherCleanup executes a single watcher cleanup function with panic recovery.
// If the cleanup panics, the panic is caught, reported to observability, and execution continues.
// This ensures that one failing watcher cleanup doesn't prevent other cleanups from executing.
//
// The method:
//   - Uses defer/recover to catch panics
//   - Reports panic to observability system (Sentry, console, etc.)
//   - Captures stack trace and context for debugging
//   - Allows execution to continue normally
//
// Example:
//
//	lm.safeExecuteWatcherCleanup(cleanup)
func (lm *LifecycleManager) safeExecuteWatcherCleanup(cleanup func()) {
	defer func() {
		if r := recover(); r != nil {
			// Report panic to observability system
			if reporter := observability.GetErrorReporter(); reporter != nil {
				panicErr := &observability.HandlerPanicError{
					ComponentName: lm.component.name,
					EventName:     "lifecycle:watcher_cleanup",
					PanicValue:    r,
				}

				ctx := &observability.ErrorContext{
					ComponentName: lm.component.name,
					ComponentID:   lm.component.id,
					EventName:     "lifecycle:watcher_cleanup",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"hook_type": "watcher_cleanup",
					},
					Extra: map[string]interface{}{
						"watcher_count": len(lm.watchers),
						"is_unmounting": lm.unmounting,
					},
				}

				reporter.ReportPanic(panicErr, ctx)
			}
		}
	}()

	// Execute the watcher cleanup function
	cleanup()
}
