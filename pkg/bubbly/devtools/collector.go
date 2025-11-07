package devtools

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// ComponentSnapshot captures the state of a component at a specific point in time.
//
// This snapshot includes all relevant information about the component including
// its identity, hierarchy, state, props, and performance metrics. Snapshots are
// immutable and represent a frozen view of the component.
//
// Thread Safety:
//
//	Snapshots are immutable after creation and safe to share across goroutines.
//
// Example:
//
//	snapshot := &ComponentSnapshot{
//	    ID:        "component-123",
//	    Name:      "Counter",
//	    Type:      "bubbly.Component",
//	    Timestamp: time.Now(),
//	}
type ComponentSnapshot struct {
	// ID is the unique identifier of the component instance
	ID string

	// Name is the human-readable name of the component
	Name string

	// Type is the Go type name of the component
	Type string

	// Status is the component's lifecycle status (e.g., "mounted", "unmounted", "updated")
	Status string

	// Parent is a reference to the parent component snapshot (nil for root)
	Parent *ComponentSnapshot

	// Children are snapshots of child components
	Children []*ComponentSnapshot

	// State contains the component's reactive state (refs, computed values)
	State map[string]interface{}

	// Props contains the component's properties
	Props map[string]interface{}

	// Refs are snapshots of reactive references in the component
	Refs []*RefSnapshot

	// Timestamp is when this snapshot was created
	Timestamp time.Time
}

// RefSnapshot captures the state of a reactive reference at a specific point in time.
//
// Thread Safety:
//
//	Snapshots are immutable after creation and safe to share across goroutines.
type RefSnapshot struct {
	// ID is the unique identifier of the ref
	ID string

	// Name is the variable name of the ref
	Name string

	// Type is the Go type of the ref's value
	Type string

	// Value is the current value of the ref
	Value interface{}

	// Watchers is the number of active watchers on this ref
	Watchers int
}

// EventRecord captures information about an event that occurred in the application.
//
// Thread Safety:
//
//	Records are immutable after creation and safe to share across goroutines.
type EventRecord struct {
	// ID is the unique identifier of the event
	ID string

	// Name is the event name (e.g., "click", "submit", "change")
	Name string

	// SourceID is the ID of the component that emitted the event
	SourceID string

	// TargetID is the ID of the component that received the event
	TargetID string

	// Payload is the event data
	Payload interface{}

	// Timestamp is when the event occurred
	Timestamp time.Time

	// Duration is how long the event handler took to execute
	Duration time.Duration
}

// ComponentHook is called when component lifecycle events occur.
//
// Implementations must be thread-safe as hooks can be called concurrently
// from multiple goroutines.
//
// If a hook panics, the panic is recovered and reported via the observability
// system. The panic does not affect other hooks or the host application.
//
// Example:
//
//	type MyHook struct{}
//
//	func (h *MyHook) OnComponentCreated(snapshot *ComponentSnapshot) {
//	    fmt.Printf("Component created: %s\n", snapshot.Name)
//	}
//
//	func (h *MyHook) OnComponentMounted(id string) {
//	    fmt.Printf("Component mounted: %s\n", id)
//	}
//
//	// ... implement other methods
type ComponentHook interface {
	// OnComponentCreated is called when a component is created
	OnComponentCreated(snapshot *ComponentSnapshot)

	// OnComponentMounted is called when a component is mounted
	OnComponentMounted(id string)

	// OnComponentUpdated is called when a component is updated
	OnComponentUpdated(id string)

	// OnComponentUnmounted is called when a component is unmounted
	OnComponentUnmounted(id string)
}

// StateHook is called when reactive state changes occur.
//
// Implementations must be thread-safe as hooks can be called concurrently
// from multiple goroutines.
//
// Example:
//
//	type MyStateHook struct{}
//
//	func (h *MyStateHook) OnRefChanged(refID string, oldValue, newValue interface{}) {
//	    fmt.Printf("Ref %s changed: %v -> %v\n", refID, oldValue, newValue)
//	}
type StateHook interface {
	// OnRefChanged is called when a ref's value changes
	OnRefChanged(refID string, oldValue, newValue interface{})

	// OnComputedEvaluated is called when a computed value is evaluated
	OnComputedEvaluated(computedID string, value interface{}, duration time.Duration)

	// OnWatcherTriggered is called when a watcher is triggered
	OnWatcherTriggered(watcherID string, value interface{})
}

// EventHook is called when events are emitted in the application.
//
// Implementations must be thread-safe as hooks can be called concurrently
// from multiple goroutines.
//
// Example:
//
//	type MyEventHook struct{}
//
//	func (h *MyEventHook) OnEvent(event *EventRecord) {
//	    fmt.Printf("Event: %s from %s\n", event.Name, event.SourceID)
//	}
type EventHook interface {
	// OnEvent is called when an event is emitted
	OnEvent(event *EventRecord)
}

// PerformanceHook is called when performance metrics are collected.
//
// Implementations must be thread-safe as hooks can be called concurrently
// from multiple goroutines.
//
// Example:
//
//	type MyPerfHook struct{}
//
//	func (h *MyPerfHook) OnRenderComplete(componentID string, duration time.Duration) {
//	    fmt.Printf("Component %s rendered in %v\n", componentID, duration)
//	}
type PerformanceHook interface {
	// OnRenderComplete is called when a component finishes rendering
	OnRenderComplete(componentID string, duration time.Duration)
}

// DataCollector manages hooks for collecting data from the application.
//
// The collector maintains lists of registered hooks and provides methods to
// fire events to all registered hooks. It ensures thread-safe access and
// protects the application from panicking hooks.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	collector := NewDataCollector()
//
//	// Register hooks
//	collector.AddComponentHook(&MyComponentHook{})
//	collector.AddStateHook(&MyStateHook{})
//
//	// Fire events
//	snapshot := &ComponentSnapshot{ID: "comp-1", Name: "Counter"}
//	collector.FireComponentCreated(snapshot)
type DataCollector struct {
	// componentHooks are hooks for component lifecycle events
	componentHooks []ComponentHook

	// stateHooks are hooks for reactive state changes
	stateHooks []StateHook

	// eventHooks are hooks for application events
	eventHooks []EventHook

	// perfHooks are hooks for performance metrics
	perfHooks []PerformanceHook

	// mu protects concurrent access to hook slices
	mu sync.RWMutex
}

// NewDataCollector creates a new data collector with empty hook lists.
//
// The returned collector is ready to use and thread-safe.
//
// Example:
//
//	collector := NewDataCollector()
//	collector.AddComponentHook(&MyHook{})
//
// Returns:
//   - *DataCollector: A new data collector instance
func NewDataCollector() *DataCollector {
	return &DataCollector{
		componentHooks: make([]ComponentHook, 0),
		stateHooks:     make([]StateHook, 0),
		eventHooks:     make([]EventHook, 0),
		perfHooks:      make([]PerformanceHook, 0),
	}
}

// AddComponentHook registers a component lifecycle hook.
//
// The hook will be called for all future component lifecycle events.
// Hooks are called in the order they were registered.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	hook := &MyComponentHook{}
//	collector.AddComponentHook(hook)
//
// Parameters:
//   - hook: The hook to register
func (dc *DataCollector) AddComponentHook(hook ComponentHook) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.componentHooks = append(dc.componentHooks, hook)
}

// RemoveComponentHook unregisters a component lifecycle hook.
//
// The hook will no longer be called for future events. If the hook is not
// registered, this is a no-op.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	collector.RemoveComponentHook(hook)
//
// Parameters:
//   - hook: The hook to unregister
func (dc *DataCollector) RemoveComponentHook(hook ComponentHook) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Find and remove the hook
	for i, h := range dc.componentHooks {
		if h == hook {
			dc.componentHooks = append(dc.componentHooks[:i], dc.componentHooks[i+1:]...)
			return
		}
	}
}

// AddStateHook registers a state change hook.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - hook: The hook to register
func (dc *DataCollector) AddStateHook(hook StateHook) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.stateHooks = append(dc.stateHooks, hook)
}

// RemoveStateHook unregisters a state change hook.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - hook: The hook to unregister
func (dc *DataCollector) RemoveStateHook(hook StateHook) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	for i, h := range dc.stateHooks {
		if h == hook {
			dc.stateHooks = append(dc.stateHooks[:i], dc.stateHooks[i+1:]...)
			return
		}
	}
}

// AddEventHook registers an event hook.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - hook: The hook to register
func (dc *DataCollector) AddEventHook(hook EventHook) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.eventHooks = append(dc.eventHooks, hook)
}

// RemoveEventHook unregisters an event hook.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - hook: The hook to unregister
func (dc *DataCollector) RemoveEventHook(hook EventHook) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	for i, h := range dc.eventHooks {
		if h == hook {
			dc.eventHooks = append(dc.eventHooks[:i], dc.eventHooks[i+1:]...)
			return
		}
	}
}

// AddPerformanceHook registers a performance hook.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - hook: The hook to register
func (dc *DataCollector) AddPerformanceHook(hook PerformanceHook) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.perfHooks = append(dc.perfHooks, hook)
}

// RemovePerformanceHook unregisters a performance hook.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - hook: The hook to unregister
func (dc *DataCollector) RemovePerformanceHook(hook PerformanceHook) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	for i, h := range dc.perfHooks {
		if h == hook {
			dc.perfHooks = append(dc.perfHooks[:i], dc.perfHooks[i+1:]...)
			return
		}
	}
}

// FireComponentCreated calls OnComponentCreated on all registered component hooks.
//
// If a hook panics, the panic is recovered and reported via the observability
// system. Other hooks continue to execute normally.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	snapshot := &ComponentSnapshot{
//	    ID:   "comp-1",
//	    Name: "Counter",
//	}
//	collector.FireComponentCreated(snapshot)
//
// Parameters:
//   - snapshot: The component snapshot
func (dc *DataCollector) FireComponentCreated(snapshot *ComponentSnapshot) {
	dc.mu.RLock()
	hooks := make([]ComponentHook, len(dc.componentHooks))
	copy(hooks, dc.componentHooks)
	dc.mu.RUnlock()

	for _, hook := range hooks {
		func(h ComponentHook) {
			defer dc.recoverHookPanic("ComponentHook.OnComponentCreated", snapshot.ID)
			h.OnComponentCreated(snapshot)
		}(hook)
	}
}

// FireComponentMounted calls OnComponentMounted on all registered component hooks.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - id: The component ID
func (dc *DataCollector) FireComponentMounted(id string) {
	dc.mu.RLock()
	hooks := make([]ComponentHook, len(dc.componentHooks))
	copy(hooks, dc.componentHooks)
	dc.mu.RUnlock()

	for _, hook := range hooks {
		func(h ComponentHook) {
			defer dc.recoverHookPanic("ComponentHook.OnComponentMounted", id)
			h.OnComponentMounted(id)
		}(hook)
	}
}

// FireComponentUpdated calls OnComponentUpdated on all registered component hooks.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - id: The component ID
func (dc *DataCollector) FireComponentUpdated(id string) {
	dc.mu.RLock()
	hooks := make([]ComponentHook, len(dc.componentHooks))
	copy(hooks, dc.componentHooks)
	dc.mu.RUnlock()

	for _, hook := range hooks {
		func(h ComponentHook) {
			defer dc.recoverHookPanic("ComponentHook.OnComponentUpdated", id)
			h.OnComponentUpdated(id)
		}(hook)
	}
}

// FireComponentUnmounted calls OnComponentUnmounted on all registered component hooks.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - id: The component ID
func (dc *DataCollector) FireComponentUnmounted(id string) {
	dc.mu.RLock()
	hooks := make([]ComponentHook, len(dc.componentHooks))
	copy(hooks, dc.componentHooks)
	dc.mu.RUnlock()

	for _, hook := range hooks {
		func(h ComponentHook) {
			defer dc.recoverHookPanic("ComponentHook.OnComponentUnmounted", id)
			h.OnComponentUnmounted(id)
		}(hook)
	}
}

// FireRefChanged calls OnRefChanged on all registered state hooks.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - refID: The ref ID
//   - oldValue: The previous value
//   - newValue: The new value
func (dc *DataCollector) FireRefChanged(refID string, oldValue, newValue interface{}) {
	dc.mu.RLock()
	hooks := make([]StateHook, len(dc.stateHooks))
	copy(hooks, dc.stateHooks)
	dc.mu.RUnlock()

	for _, hook := range hooks {
		func(h StateHook) {
			defer dc.recoverHookPanic("StateHook.OnRefChanged", refID)
			h.OnRefChanged(refID, oldValue, newValue)
		}(hook)
	}
}

// FireEvent calls OnEvent on all registered event hooks.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - event: The event record
func (dc *DataCollector) FireEvent(event *EventRecord) {
	dc.mu.RLock()
	hooks := make([]EventHook, len(dc.eventHooks))
	copy(hooks, dc.eventHooks)
	dc.mu.RUnlock()

	for _, hook := range hooks {
		func(h EventHook) {
			defer dc.recoverHookPanic("EventHook.OnEvent", event.ID)
			h.OnEvent(event)
		}(hook)
	}
}

// FireRenderComplete calls OnRenderComplete on all registered performance hooks.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - componentID: The component ID
//   - duration: How long the render took
func (dc *DataCollector) FireRenderComplete(componentID string, duration time.Duration) {
	dc.mu.RLock()
	hooks := make([]PerformanceHook, len(dc.perfHooks))
	copy(hooks, dc.perfHooks)
	dc.mu.RUnlock()

	for _, hook := range hooks {
		func(h PerformanceHook) {
			defer dc.recoverHookPanic("PerformanceHook.OnRenderComplete", componentID)
			h.OnRenderComplete(componentID, duration)
		}(hook)
	}
}

// recoverHookPanic recovers from a panic in a hook and reports it via observability.
//
// This ensures that panicking hooks don't crash the application or affect other hooks.
//
// Parameters:
//   - hookType: The type of hook that panicked (for error context)
//   - contextID: The ID of the component/ref/event (for error context)
func (dc *DataCollector) recoverHookPanic(hookType, contextID string) {
	if r := recover(); r != nil {
		// Report panic to observability system
		if reporter := observability.GetErrorReporter(); reporter != nil {
			err := &observability.HandlerPanicError{
				ComponentName: "DevTools",
				EventName:     hookType,
				PanicValue:    r,
			}

			ctx := &observability.ErrorContext{
				ComponentName: "DataCollector",
				ComponentID:   contextID,
				EventName:     hookType,
				Timestamp:     time.Now(),
				StackTrace:    debug.Stack(),
				Tags: map[string]string{
					"hook_type":  hookType,
					"context_id": contextID,
				},
				Extra: map[string]interface{}{
					"panic_value": r,
				},
			}

			reporter.ReportPanic(err, ctx)
		}
	}
}
