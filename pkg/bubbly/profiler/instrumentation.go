// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// KeyBinding represents a key binding for component instrumentation.
// This is a local copy to avoid circular dependencies with the bubbly package.
type KeyBinding struct {
	// Key is the keyboard key or key combination (e.g., "space", "ctrl+c", "up").
	Key string

	// Event is the event name to emit when the key is pressed.
	Event string

	// Description is a human-readable description for help text.
	Description string

	// Condition is an optional function that returns true if the binding should be active.
	Condition func() bool

	// Data is optional data to include with the emitted event.
	Data interface{}
}

// Component is the interface that BubblyUI components implement.
// This is a subset of the full bubbly.Component interface, containing
// only the methods needed for instrumentation.
//
// This interface allows the profiler package to work with components
// without creating a circular dependency with the bubbly package.
type Component interface {
	tea.Model

	// Name returns the component's name (e.g., "Button", "Counter").
	Name() string

	// ID returns the component's unique instance identifier.
	ID() string

	// Props returns the component's props (configuration data).
	Props() interface{}

	// Emit sends a custom event with associated data.
	Emit(event string, data interface{})

	// On registers an event handler for the specified event name.
	On(event string, handler func(interface{}))

	// KeyBindings returns the component's registered key bindings.
	KeyBindings() map[string][]KeyBinding

	// HelpText generates a formatted help string from registered key bindings.
	HelpText() string

	// IsInitialized returns whether the component has been initialized.
	IsInitialized() bool
}

// Instrumentor provides component instrumentation for performance profiling.
//
// It wraps BubblyUI components to automatically track render and update timing,
// providing detailed performance metrics with minimal overhead.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	inst := NewInstrumentor(profiler)
//	inst.Enable()
//
//	// Option 1: Manual instrumentation
//	stop := inst.InstrumentRender(component)
//	output := component.View()
//	stop()
//
//	// Option 2: Automatic instrumentation via wrapper
//	wrapped := inst.InstrumentComponent(component)
//	output := wrapped.View() // Automatically timed
type Instrumentor struct {
	// profiler is the parent profiler instance
	profiler *Profiler

	// componentTracker tracks per-component metrics
	componentTracker *ComponentTracker

	// collector handles timing collection
	collector *MetricCollector

	// enabled indicates whether instrumentation is active
	enabled atomic.Bool

	// mu protects access to internal state
	mu sync.RWMutex
}

// NewInstrumentor creates a new component instrumentor.
//
// If profiler is nil, a new profiler with default settings is created.
// The instrumentor starts in a disabled state; call Enable() to begin tracking.
//
// Example:
//
//	inst := NewInstrumentor(profiler)
//	inst.Enable()
func NewInstrumentor(profiler *Profiler) *Instrumentor {
	if profiler == nil {
		profiler = New()
	}

	return &Instrumentor{
		profiler:         profiler,
		componentTracker: NewComponentTracker(),
		collector:        NewMetricCollector(),
	}
}

// Enable activates component instrumentation.
//
// When enabled, InstrumentRender, InstrumentUpdate, and InstrumentComponent
// will record timing metrics. When disabled, these methods have minimal overhead.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (i *Instrumentor) Enable() {
	i.enabled.Store(true)
	i.collector.Enable()
}

// Disable deactivates component instrumentation.
//
// When disabled, instrumentation methods return no-op functions with minimal overhead.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (i *Instrumentor) Disable() {
	i.enabled.Store(false)
	i.collector.Disable()
}

// IsEnabled returns whether instrumentation is currently active.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (i *Instrumentor) IsEnabled() bool {
	return i.enabled.Load()
}

// InstrumentRender starts timing a render operation for a component.
//
// Returns a stop function that must be called when the render completes.
// The stop function records the render duration to the component tracker.
//
// If the instrumentor is disabled or component is nil, returns a no-op function.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	stop := inst.InstrumentRender(component)
//	output := component.View()
//	stop()
func (i *Instrumentor) InstrumentRender(component Component) func() {
	// Fast path when disabled
	if !i.enabled.Load() {
		return func() {}
	}

	// Handle nil component
	if component == nil {
		return func() {}
	}

	id := component.ID()
	name := component.Name()
	start := time.Now()

	return func() {
		duration := time.Since(start)
		i.componentTracker.RecordRender(id, name, duration)
	}
}

// InstrumentUpdate starts timing an update operation for a component.
//
// Returns a stop function that must be called when the update completes.
// The stop function records the update duration to the metric collector.
//
// If the instrumentor is disabled or component is nil, returns a no-op function.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	stop := inst.InstrumentUpdate(component)
//	model, cmd := component.Update(msg)
//	stop()
func (i *Instrumentor) InstrumentUpdate(component Component) func() {
	// Fast path when disabled
	if !i.enabled.Load() {
		return func() {}
	}

	// Handle nil component
	if component == nil {
		return func() {}
	}

	id := component.ID()
	metricName := "update." + id
	start := time.Now()

	return func() {
		duration := time.Since(start)
		i.collector.GetTimings().Record(metricName, duration)
	}
}

// InstrumentComponent wraps a component with automatic instrumentation.
//
// The returned component automatically tracks render and update timing
// whenever View() or Update() is called.
//
// Returns nil if component is nil.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	wrapped := inst.InstrumentComponent(component)
//	output := wrapped.View() // Automatically timed
func (i *Instrumentor) InstrumentComponent(component Component) Component {
	if component == nil {
		return nil
	}

	return &instrumentedComponent{
		original:     component,
		instrumentor: i,
	}
}

// GetComponentTracker returns the component tracker for direct access.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (i *Instrumentor) GetComponentTracker() *ComponentTracker {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.componentTracker
}

// GetCollector returns the metric collector for direct access.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (i *Instrumentor) GetCollector() *MetricCollector {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.collector
}

// Reset clears all collected instrumentation data.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (i *Instrumentor) Reset() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.componentTracker.Reset()
	i.collector.Reset()
}

// instrumentedComponent wraps a Component with automatic instrumentation.
//
// It implements the Component interface and delegates all methods to the
// original component while automatically timing View() and Update() calls.
type instrumentedComponent struct {
	original     Component
	instrumentor *Instrumentor
}

// Init implements tea.Model.Init().
// Delegates to the original component's Init method.
func (ic *instrumentedComponent) Init() tea.Cmd {
	return ic.original.Init()
}

// Update implements tea.Model.Update().
// Times the update operation and records metrics.
func (ic *instrumentedComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !ic.instrumentor.enabled.Load() {
		return ic.original.Update(msg)
	}

	id := ic.original.ID()
	metricName := "update." + id
	start := time.Now()

	model, cmd := ic.original.Update(msg)

	duration := time.Since(start)
	ic.instrumentor.collector.GetTimings().Record(metricName, duration)

	return model, cmd
}

// View implements tea.Model.View().
// Times the render operation and records metrics.
func (ic *instrumentedComponent) View() string {
	if !ic.instrumentor.enabled.Load() {
		return ic.original.View()
	}

	id := ic.original.ID()
	name := ic.original.Name()
	start := time.Now()

	result := ic.original.View()

	duration := time.Since(start)
	ic.instrumentor.componentTracker.RecordRender(id, name, duration)

	return result
}

// Name implements Component.Name().
// Delegates to the original component.
func (ic *instrumentedComponent) Name() string {
	return ic.original.Name()
}

// ID implements Component.ID().
// Delegates to the original component.
func (ic *instrumentedComponent) ID() string {
	return ic.original.ID()
}

// Props implements Component.Props().
// Delegates to the original component.
func (ic *instrumentedComponent) Props() interface{} {
	return ic.original.Props()
}

// Emit implements Component.Emit().
// Delegates to the original component.
func (ic *instrumentedComponent) Emit(event string, data interface{}) {
	ic.original.Emit(event, data)
}

// On implements Component.On().
// Delegates to the original component.
func (ic *instrumentedComponent) On(event string, handler func(interface{})) {
	ic.original.On(event, handler)
}

// KeyBindings implements Component.KeyBindings().
// Delegates to the original component.
func (ic *instrumentedComponent) KeyBindings() map[string][]KeyBinding {
	return ic.original.KeyBindings()
}

// HelpText implements Component.HelpText().
// Delegates to the original component.
func (ic *instrumentedComponent) HelpText() string {
	return ic.original.HelpText()
}

// IsInitialized implements Component.IsInitialized().
// Delegates to the original component.
func (ic *instrumentedComponent) IsInitialized() bool {
	return ic.original.IsInitialized()
}

// Ensure instrumentedComponent implements Component interface
var _ Component = (*instrumentedComponent)(nil)
