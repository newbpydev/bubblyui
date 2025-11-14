package bubbly

import (
	"sync"
	"time"
)

// FrameworkHook defines the interface for observing framework lifecycle events.
//
// FrameworkHook provides a high-level interface for integrating with the BubblyUI
// framework's lifecycle. It acts as an adapter over the lower-level instrumentation
// system, providing a clean extension point for custom observers.
//
// The hook methods are called automatically when framework events occur:
//   - OnComponentMount: When a component is initialized via Init()
//   - OnComponentUpdate: When a component receives a Bubbletea message via Update()
//   - OnComponentUnmount: When a component is cleaned up via Unmount()
//   - OnRefChange: When a Ref value changes via Set()
//   - OnEvent: When a component emits an event via Emit()
//   - OnRenderComplete: When a component completes rendering via View()
//   - OnComputedChange: When a computed value re-evaluates to a new value
//   - OnWatchCallback: When a watcher callback is about to execute
//   - OnEffectRun: When a WatchEffect function is about to run
//   - OnChildAdded: When a child component is added to a parent
//   - OnChildRemoved: When a child component is removed from a parent
//
// Thread Safety:
//
//	All hook methods may be called concurrently from multiple goroutines
//	and must be thread-safe.
//
// Example:
//
//	type MyHook struct{}
//
//	func (h *MyHook) OnComponentMount(id, name string) {
//	    fmt.Printf("Component mounted: %s (%s)\n", name, id)
//	}
//
//	// Implement other methods...
//
//	hook := &MyHook{}
//	devtools.RegisterHook(hook)
type FrameworkHook interface {
	// OnComponentMount is called when a component is initialized and mounted.
	//
	// Parameters:
	//   - id: The component's unique ID
	//   - name: The component's name (e.g., "Button", "Counter")
	OnComponentMount(id, name string)

	// OnComponentUpdate is called when a component receives a Bubbletea message.
	//
	// This is called for every Update() call, which can be very frequent.
	// Implementations should be fast and avoid blocking operations.
	//
	// Parameters:
	//   - id: The component's unique ID
	//   - msg: The Bubbletea message (can be any type)
	OnComponentUpdate(id string, msg interface{})

	// OnComponentUnmount is called when a component is unmounted and cleaned up.
	//
	// Parameters:
	//   - id: The component's unique ID
	OnComponentUnmount(id string)

	// OnRefChange is called when a Ref's value changes.
	//
	// Parameters:
	//   - id: The ref's identifier
	//   - oldValue: The previous value
	//   - newValue: The new value
	OnRefChange(id string, oldValue, newValue interface{})

	
	// OnEvent is called when a component emits an event.
	//
	// Parameters:
	//   - componentID: The component that emitted the event
	//   - eventName: The name of the event
	//   - data: The event payload
	OnEvent(componentID, eventName string, data interface{})

	// OnRenderComplete is called when a component completes rendering.
	//
	// Parameters:
	//   - componentID: The component that rendered
	//   - duration: How long the render took
	OnRenderComplete(componentID string, duration time.Duration)

	// OnComputedChange is called when a computed value re-evaluates and the result changes.
	//
	// This is called after the computed value detects a change (via deep equal check)
	// but BEFORE notifying watchers, maintaining proper cascade order.
	//
	// Parameters:
	//   - id: The computed value's identifier (format: "computed-0xHEX")
	//   - oldValue: The previous cached value
	//   - newValue: The new computed value
	OnComputedChange(id string, oldValue, newValue interface{})

	// OnWatchCallback is called when a watcher callback is about to be executed.
	//
	// This is called BEFORE the watcher callback executes, allowing dev tools
	// to track the reactive cascade from source changes to watcher notifications.
	//
	// Parameters:
	//   - watcherID: The watcher's identifier (format: "watch-0xHEX")
	//   - newValue: The new value being passed to the callback
	//   - oldValue: The old value being passed to the callback
	OnWatchCallback(watcherID string, newValue, oldValue interface{})

	// OnEffectRun is called when a WatchEffect function is about to execute.
	//
	// This is called BEFORE the effect function executes, on every run including
	// the initial run and all re-runs triggered by dependency changes.
	//
	// Parameters:
	//   - effectID: The effect's identifier (format: "effect-0xHEX")
	OnEffectRun(effectID string)

	// OnChildAdded is called when a child component is added to a parent.
	//
	// This is called AFTER the child is successfully added to the parent's
	// children slice and the parent reference is set, ensuring tree consistency.
	//
	// Parameters:
	//   - parentID: The parent component's unique ID
	//   - childID: The child component's unique ID
	OnChildAdded(parentID, childID string)

	// OnChildRemoved is called when a child component is removed from a parent.
	//
	// This is called AFTER the child is successfully removed from the parent's
	// children slice and the parent reference is cleared, ensuring tree consistency.
	//
	// Parameters:
	//   - parentID: The parent component's unique ID
	//   - childID: The child component's unique ID
	OnChildRemoved(parentID, childID string)

	// OnRefExposed is called when a component exposes a Ref via ctx.Expose().
	//
	// This hook is CRITICAL for DevTools to track ref ownership - which refs
	// belong to which components. Without this, DevTools cannot accurately
	// display component state or detect reactive dependencies.
	//
	// This is called AFTER the ref is stored in the component's state map,
	// ensuring the ref is accessible via ctx.Get().
	//
	// Parameters:
	//   - componentID: The component that owns the ref
	//   - refID: The unique ref identifier (format: "ref-0xHEX")
	//   - refName: The name given to the ref (e.g., "count", "message")
	OnRefExposed(componentID, refID, refName string)
}

// hookRegistry manages the registered framework hook.
type hookRegistry struct {
	mu   sync.RWMutex
	hook FrameworkHook
}

// globalHookRegistry is the singleton instance for hook registration.
var globalHookRegistry = &hookRegistry{}

// RegisterHook registers a framework hook to observe lifecycle events.
//
// Only one hook can be registered at a time. Calling RegisterHook multiple
// times will replace the previous hook. To unregister, call UnregisterHook().
//
// The hook methods will be called automatically when framework events occur.
// All hook methods must be thread-safe as they may be called concurrently.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	hook := &MyFrameworkHook{}
//	err := devtools.RegisterHook(hook)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - hook: The FrameworkHook implementation to register
//
// Returns:
//   - error: Always returns nil (reserved for future validation)
func RegisterHook(hook FrameworkHook) error {
	globalHookRegistry.mu.Lock()
	defer globalHookRegistry.mu.Unlock()
	globalHookRegistry.hook = hook
	return nil
}

// UnregisterHook removes the currently registered framework hook.
//
// After calling this, no hook methods will be called until a new hook
// is registered via RegisterHook().
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - error: Always returns nil (reserved for future validation)
func UnregisterHook() error {
	globalHookRegistry.mu.Lock()
	defer globalHookRegistry.mu.Unlock()
	globalHookRegistry.hook = nil
	return nil
}

// IsHookRegistered returns whether a framework hook is currently registered.
//
// This can be used to check if hooks are active before performing
// expensive operations that would only be useful for hook observers.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - bool: True if a hook is registered, false otherwise
func IsHookRegistered() bool {
	globalHookRegistry.mu.RLock()
	defer globalHookRegistry.mu.RUnlock()
	return globalHookRegistry.hook != nil
}

// notifyHookComponentMount calls the registered hook's OnComponentMount method.
// This is an internal helper used by framework integration points.
func notifyHookComponentMount(id, name string) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnComponentMount(id, name)
	}
}

// notifyHookComponentUpdate calls the registered hook's OnComponentUpdate method.
// This is an internal helper used by framework integration points.
func notifyHookComponentUpdate(id string, msg interface{}) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnComponentUpdate(id, msg)
	}
}

// notifyHookComponentUnmount calls the registered hook's OnComponentUnmount method.
// This is an internal helper used by framework integration points.
func notifyHookComponentUnmount(id string) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnComponentUnmount(id)
	}
}


// notifyHookRefChange calls the registered hook's OnRefChange method.
// This is an internal helper used by framework integration points.
func notifyHookRefChange(id string, oldValue, newValue interface{}) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnRefChange(id, oldValue, newValue)
	}
}

// notifyHookEvent calls the registered hook's OnEvent method.
// This is an internal helper used by framework integration points.
func notifyHookEvent(componentID, eventName string, data interface{}) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnEvent(componentID, eventName, data)
	}
}

// notifyHookRenderComplete calls the registered hook's OnRenderComplete method.
// This is an internal helper used by framework integration points.
func notifyHookRenderComplete(componentID string, duration time.Duration) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnRenderComplete(componentID, duration)
	}
}

// notifyHookComputedChange calls the registered hook's OnComputedChange method.
// This is an internal helper used by framework integration points.
func notifyHookComputedChange(id string, oldValue, newValue interface{}) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnComputedChange(id, oldValue, newValue)
	}
}

// notifyHookWatchCallback calls the registered hook's OnWatchCallback method.
// This is an internal helper used by framework integration points.
func notifyHookWatchCallback(watcherID string, newValue, oldValue interface{}) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnWatchCallback(watcherID, newValue, oldValue)
	}
}

// notifyHookEffectRun calls the registered hook's OnEffectRun method.
// This is an internal helper used by framework integration points.
func notifyHookEffectRun(effectID string) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnEffectRun(effectID)
	}
}

// notifyHookChildAdded calls the registered hook's OnChildAdded method.
// This is an internal helper used by framework integration points.
func notifyHookChildAdded(parentID, childID string) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnChildAdded(parentID, childID)
	}
}

// notifyHookChildRemoved calls the registered hook's OnChildRemoved method.
// This is an internal helper used by framework integration points.
func notifyHookChildRemoved(parentID, childID string) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnChildRemoved(parentID, childID)
	}
}

// notifyHookRefExposed calls the registered hook's OnRefExposed method.
// This is an internal helper used by framework integration points.
func notifyHookRefExposed(componentID, refID, refName string) {
	globalHookRegistry.mu.RLock()
	hook := globalHookRegistry.hook
	globalHookRegistry.mu.RUnlock()

	if hook != nil {
		hook.OnRefExposed(componentID, refID, refName)
	}
}
