package devtools

import (
	"sync"
	"time"
)

// Instrumentor manages instrumentation hooks for collecting debug data from the application.
//
// The instrumentor acts as a bridge between the application code (components, refs, events)
// and the dev tools data collector. It provides a global singleton that application code
// can call at key lifecycle points without needing to know about the collector directly.
//
// When dev tools are disabled (collector is nil), all instrumentation calls become no-ops
// with minimal overhead (just a nil check).
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently from multiple goroutines.
//
// Example:
//
//	// In component lifecycle
//	devtools.NotifyComponentMounted(component.ID())
//
//	// In ref.Set()
//	devtools.NotifyRefChanged(ref.id, oldValue, newValue)
type Instrumentor struct {
	collector *DataCollector
	mu        sync.RWMutex
}

// globalInstrumentor is the singleton instance used by all instrumentation calls.
var globalInstrumentor = &Instrumentor{}

// SetCollector sets the data collector for instrumentation.
//
// When a collector is set, all subsequent instrumentation calls will forward
// to the collector's hooks. When set to nil, instrumentation is disabled.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	collector := devtools.NewDataCollector()
//	devtools.SetCollector(collector)
//
// Parameters:
//   - collector: The data collector to use, or nil to disable instrumentation
func SetCollector(collector *DataCollector) {
	globalInstrumentor.mu.Lock()
	defer globalInstrumentor.mu.Unlock()
	globalInstrumentor.collector = collector
}

// GetCollector returns the current data collector.
//
// Returns nil if instrumentation is disabled.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - *DataCollector: The current collector, or nil if disabled
func GetCollector() *DataCollector {
	globalInstrumentor.mu.RLock()
	defer globalInstrumentor.mu.RUnlock()
	return globalInstrumentor.collector
}

// NotifyComponentCreated notifies the collector that a component was created.
//
// This should be called when a new component instance is created, typically
// in the component builder's Build() method.
//
// If no collector is set, this is a no-op with minimal overhead.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	snapshot := &ComponentSnapshot{
//	    ID:   component.ID(),
//	    Name: component.Name(),
//	}
//	devtools.NotifyComponentCreated(snapshot)
//
// Parameters:
//   - snapshot: The component snapshot containing component state
func NotifyComponentCreated(snapshot *ComponentSnapshot) {
	globalInstrumentor.mu.RLock()
	collector := globalInstrumentor.collector
	globalInstrumentor.mu.RUnlock()

	if collector != nil {
		collector.FireComponentCreated(snapshot)
	}
}

// NotifyComponentMounted notifies the collector that a component was mounted.
//
// This should be called when a component's Init() method completes and the
// component is ready for use.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	devtools.NotifyComponentMounted(component.ID())
//
// Parameters:
//   - id: The component ID
func NotifyComponentMounted(id string) {
	globalInstrumentor.mu.RLock()
	collector := globalInstrumentor.collector
	globalInstrumentor.mu.RUnlock()

	if collector != nil {
		collector.FireComponentMounted(id)
	}
}

// NotifyComponentUpdated notifies the collector that a component was updated.
//
// This should be called when a component's Update() method is called with
// a Bubbletea message.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	devtools.NotifyComponentUpdated(component.ID())
//
// Parameters:
//   - id: The component ID
func NotifyComponentUpdated(id string) {
	globalInstrumentor.mu.RLock()
	collector := globalInstrumentor.collector
	globalInstrumentor.mu.RUnlock()

	if collector != nil {
		collector.FireComponentUpdated(id)
	}
}

// NotifyComponentUnmounted notifies the collector that a component was unmounted.
//
// This should be called when a component is being destroyed and cleaned up.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	devtools.NotifyComponentUnmounted(component.ID())
//
// Parameters:
//   - id: The component ID
func NotifyComponentUnmounted(id string) {
	globalInstrumentor.mu.RLock()
	collector := globalInstrumentor.collector
	globalInstrumentor.mu.RUnlock()

	if collector != nil {
		collector.FireComponentUnmounted(id)
	}
}

// NotifyRefChanged notifies the collector that a ref's value changed.
//
// This should be called in Ref.Set() after the value has been updated.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	devtools.NotifyRefChanged(ref.id, oldValue, newValue)
//
// Parameters:
//   - refID: The ref ID
//   - oldValue: The previous value
//   - newValue: The new value
func NotifyRefChanged(refID string, oldValue, newValue interface{}) {
	globalInstrumentor.mu.RLock()
	collector := globalInstrumentor.collector
	globalInstrumentor.mu.RUnlock()

	if collector != nil {
		collector.FireRefChanged(refID, oldValue, newValue)
	}
}

// NotifyEvent notifies the collector that an event was emitted.
//
// This should be called when a component emits an event via Emit().
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	event := &EventRecord{
//	    ID:        generateID(),
//	    Name:      eventName,
//	    SourceID:  component.ID(),
//	    Timestamp: time.Now(),
//	}
//	devtools.NotifyEvent(event)
//
// Parameters:
//   - event: The event record
func NotifyEvent(event *EventRecord) {
	globalInstrumentor.mu.RLock()
	collector := globalInstrumentor.collector
	globalInstrumentor.mu.RUnlock()

	if collector != nil {
		collector.FireEvent(event)
	}
}

// NotifyRenderComplete notifies the collector that a component finished rendering.
//
// This should be called after a component's View() method completes, with
// the duration of the render operation.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	start := time.Now()
//	output := component.View()
//	duration := time.Since(start)
//	devtools.NotifyRenderComplete(component.ID(), duration)
//
// Parameters:
//   - componentID: The component ID
//   - duration: How long the render took
func NotifyRenderComplete(componentID string, duration time.Duration) {
	globalInstrumentor.mu.RLock()
	collector := globalInstrumentor.collector
	globalInstrumentor.mu.RUnlock()

	if collector != nil {
		collector.FireRenderComplete(componentID, duration)
	}
}
