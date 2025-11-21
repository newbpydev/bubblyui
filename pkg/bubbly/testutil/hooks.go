package testutil

import (
	"sync"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestHooks provides callbacks for tracking component behavior during tests.
// These hooks are installed into components to monitor state changes, events,
// and updates, enabling comprehensive testing of component behavior.
//
// TestHooks is designed to be non-intrusive and can be installed/removed
// dynamically during tests. All callbacks are optional and thread-safe.
//
// Example usage:
//
//	harness := testutil.NewHarness(t)
//	component := createMyComponent()
//
//	// Install hooks to track behavior
//	harness.installHooks(component)
//
//	// Perform test actions
//	component.Update(someMessage)
//
//	// Remove hooks when done
//	harness.removeHooks()
//
// Note: This is a foundational implementation. Full integration with component
// state tracking and event system will be completed in future tasks as those
// systems expose the necessary hooks.
type TestHooks struct {
	// onStateChange is called when a ref or computed value changes.
	// Parameters: name (string), newValue (interface{})
	onStateChange func(string, interface{})

	// onEvent is called when an event is emitted.
	// Parameters: eventName (string), payload (interface{})
	onEvent func(string, interface{})

	// onUpdate is called when the component's Update() method is called.
	onUpdate func()

	// mu protects concurrent access to callbacks
	mu sync.RWMutex
}

// NewTestHooks creates a new TestHooks instance with no callbacks set.
// Callbacks can be set after creation using the Set* methods.
//
// Returns:
//   - *TestHooks: A new test hooks instance
//
// Example:
//
//	hooks := NewTestHooks()
//	hooks.SetOnStateChange(func(name string, value interface{}) {
//	    fmt.Printf("State changed: %s = %v\n", name, value)
//	})
func NewTestHooks() *TestHooks {
	return &TestHooks{}
}

// SetOnStateChange sets the callback for state changes.
// The callback receives the name of the changed ref/computed and its new value.
//
// This method is thread-safe and can be called while hooks are active.
//
// Parameters:
//   - fn: Callback function(name string, value interface{})
//
// Example:
//
//	hooks.SetOnStateChange(func(name string, value interface{}) {
//	    stateChanges = append(stateChanges, name)
//	})
func (h *TestHooks) SetOnStateChange(fn func(string, interface{})) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onStateChange = fn
}

// SetOnEvent sets the callback for event emissions.
// The callback receives the event name and payload.
//
// This method is thread-safe and can be called while hooks are active.
//
// Parameters:
//   - fn: Callback function(eventName string, payload interface{})
//
// Example:
//
//	hooks.SetOnEvent(func(eventName string, payload interface{}) {
//	    eventsFired = append(eventsFired, eventName)
//	})
func (h *TestHooks) SetOnEvent(fn func(string, interface{})) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onEvent = fn
}

// SetOnUpdate sets the callback for component updates.
// The callback is called whenever the component's Update() method is invoked.
//
// This method is thread-safe and can be called while hooks are active.
//
// Parameters:
//   - fn: Callback function()
//
// Example:
//
//	hooks.SetOnUpdate(func() {
//	    updateCount++
//	})
func (h *TestHooks) SetOnUpdate(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onUpdate = fn
}

// TriggerStateChange invokes the onStateChange callback if set.
// This is called internally when state changes are detected.
//
// This method is thread-safe and safe to call even if no callback is set.
//
// Parameters:
//   - name: Name of the changed ref/computed
//   - value: New value
func (h *TestHooks) TriggerStateChange(name string, value interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.onStateChange != nil {
		h.onStateChange(name, value)
	}
}

// TriggerEvent invokes the onEvent callback if set.
// This is called internally when events are emitted.
//
// This method is thread-safe and safe to call even if no callback is set.
//
// Parameters:
//   - eventName: Name of the emitted event
//   - payload: Event payload
func (h *TestHooks) TriggerEvent(eventName string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.onEvent != nil {
		h.onEvent(eventName, payload)
	}
}

// TriggerUpdate invokes the onUpdate callback if set.
// This is called internally when Update() is called.
//
// This method is thread-safe and safe to call even if no callback is set.
func (h *TestHooks) TriggerUpdate() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.onUpdate != nil {
		h.onUpdate()
	}
}

// Clear removes all callbacks from the hooks.
// This is useful for resetting hooks between tests or when removing hooks.
//
// This method is thread-safe.
//
// Example:
//
//	hooks.Clear()
//	// All callbacks are now nil
func (h *TestHooks) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.onStateChange = nil
	h.onEvent = nil
	h.onUpdate = nil
}

// installHooks installs test hooks into the component.
// This enables tracking of state changes, events, and updates during testing.
//
// The hooks are stored in the harness and can be accessed via the hooks field.
// Call removeHooks() to clean up when done.
//
// Note: Current implementation creates the hooks infrastructure. Full integration
// with component internals will be completed when the component interface exposes
// the necessary hook points for state and event tracking.
//
// Parameters:
//   - component: The component to install hooks into
//
// Example:
//
//	harness := NewHarness(t)
//	component := createMyComponent()
//	harness.installHooks(component)
//	// Hooks are now active
func (h *TestHarness) installHooks(component bubbly.Component) {
	// Create hooks if not already created
	if h.hooks == nil {
		h.hooks = NewTestHooks()
	}

	// Store component reference for hook integration
	h.component = component

	// TODO: When component interface exposes hook points:
	// - Register state change listeners on refs/computed
	// - Register event emission interceptors
	// - Wrap Update() to trigger onUpdate callback
	//
	// For now, hooks infrastructure is ready for future integration
}

// removeHooks removes test hooks from the component.
// This cleans up all hook callbacks and restores normal component behavior.
//
// This method is idempotent - calling it multiple times is safe.
//
// Example:
//
//	harness.removeHooks()
//	// Hooks are now removed
func (h *TestHarness) removeHooks() {
	if h.hooks != nil {
		h.hooks.Clear()
		h.hooks = nil
	}

	// TODO: When component interface exposes hook points:
	// - Unregister state change listeners
	// - Unregister event interceptors
	// - Restore original Update() behavior
}
