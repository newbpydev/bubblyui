// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ProfilerHookAdapter implements bubbly.FrameworkHook to collect profiling data.
//
// It tracks component render times, memory usage, and other metrics
// for the profiler's ComponentTracker.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	prof := profiler.New(profiler.WithEnabled(true))
//	hook := profiler.NewProfilerHookAdapter(prof)
//	bubbly.RegisterHook(hook)
type ProfilerHookAdapter struct {
	profiler         *Profiler
	componentTracker *ComponentTracker
	componentNames   map[string]string // Maps component IDs to names
	mu               sync.RWMutex
}

// Ensure ProfilerHookAdapter implements bubbly.FrameworkHook
var _ bubbly.FrameworkHook = (*ProfilerHookAdapter)(nil)

// NewProfilerHookAdapter creates a new profiler hook adapter.
//
// The adapter integrates with the framework's hook system to automatically
// collect render timing and other metrics for the profiler.
//
// Example:
//
//	prof := profiler.New(profiler.WithEnabled(true))
//	hook := profiler.NewProfilerHookAdapter(prof)
func NewProfilerHookAdapter(prof *Profiler) *ProfilerHookAdapter {
	return &ProfilerHookAdapter{
		profiler:         prof,
		componentTracker: NewComponentTracker(),
		componentNames:   make(map[string]string),
	}
}

// GetComponentTracker returns the component tracker used by this adapter.
func (h *ProfilerHookAdapter) GetComponentTracker() *ComponentTracker {
	return h.componentTracker
}

// OnComponentMount tracks when a component is mounted.
func (h *ProfilerHookAdapter) OnComponentMount(id, name string) {
	h.mu.Lock()
	h.componentNames[id] = name
	h.mu.Unlock()
}

// OnComponentUpdate is called when a component receives an update.
func (h *ProfilerHookAdapter) OnComponentUpdate(id string, msg interface{}) {
	// Track update count if needed
}

// OnComponentUnmount tracks when a component is unmounted.
func (h *ProfilerHookAdapter) OnComponentUnmount(id string) {
	h.mu.Lock()
	delete(h.componentNames, id)
	h.mu.Unlock()
}

// OnRefChange is called when a ref value changes.
func (h *ProfilerHookAdapter) OnRefChange(id string, oldValue, newValue interface{}) {
	// Can track state change frequency if needed
}

// OnEvent is called when an event is emitted.
func (h *ProfilerHookAdapter) OnEvent(componentID, eventName string, data interface{}) {
	// Can track event frequency if needed
}

// OnRenderComplete records render timing for components.
// This is the CRITICAL method for profiler data collection.
func (h *ProfilerHookAdapter) OnRenderComplete(componentID string, duration time.Duration) {
	h.mu.RLock()
	name := h.componentNames[componentID]
	h.mu.RUnlock()

	if name == "" {
		name = "Unknown"
	}

	// Record render timing in component tracker
	// This populates the Components section of the profiler report
	h.componentTracker.RecordRender(componentID, name, duration)
}

// OnComputedChange is called when a computed value changes.
func (h *ProfilerHookAdapter) OnComputedChange(id string, oldValue, newValue interface{}) {}

// OnWatchCallback is called when a watch callback executes.
func (h *ProfilerHookAdapter) OnWatchCallback(watcherID string, newValue, oldValue interface{}) {}

// OnEffectRun is called when an effect runs.
func (h *ProfilerHookAdapter) OnEffectRun(effectID string) {}

// OnChildAdded is called when a child component is added.
func (h *ProfilerHookAdapter) OnChildAdded(parentID, childID string) {}

// OnChildRemoved is called when a child component is removed.
func (h *ProfilerHookAdapter) OnChildRemoved(parentID, childID string) {}

// OnRefExposed is called when a ref is exposed.
func (h *ProfilerHookAdapter) OnRefExposed(componentID, refID, refName string) {}

// =============================================================================
// CompositeHook - Multiplexes events to multiple hooks
// =============================================================================

// CompositeHook multiplexes framework events to multiple hook implementations.
//
// This allows both DevTools and Profiler to receive framework events,
// working around the single-hook limitation.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	// Create individual hooks
//	devtoolsHook := bubbly.GetRegisteredHook() // Get DevTools hook
//	profilerHook := profiler.NewProfilerHookAdapter(prof)
//
//	// Combine them
//	composite := profiler.NewCompositeHook(devtoolsHook, profilerHook)
//	bubbly.RegisterHook(composite)
type CompositeHook struct {
	hooks []bubbly.FrameworkHook
	mu    sync.RWMutex
}

// Ensure CompositeHook implements bubbly.FrameworkHook
var _ bubbly.FrameworkHook = (*CompositeHook)(nil)

// NewCompositeHook creates a new composite hook that forwards to multiple hooks.
//
// Example:
//
//	composite := profiler.NewCompositeHook(hook1, hook2, hook3)
func NewCompositeHook(hooks ...bubbly.FrameworkHook) *CompositeHook {
	return &CompositeHook{
		hooks: hooks,
	}
}

// AddHook adds a hook to the composite.
func (c *CompositeHook) AddHook(hook bubbly.FrameworkHook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hooks = append(c.hooks, hook)
}

// OnComponentMount forwards to all hooks.
func (c *CompositeHook) OnComponentMount(id, name string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnComponentMount(id, name)
		}
	}
}

// OnComponentUpdate forwards to all hooks.
func (c *CompositeHook) OnComponentUpdate(id string, msg interface{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnComponentUpdate(id, msg)
		}
	}
}

// OnComponentUnmount forwards to all hooks.
func (c *CompositeHook) OnComponentUnmount(id string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnComponentUnmount(id)
		}
	}
}

// OnRefChange forwards to all hooks.
func (c *CompositeHook) OnRefChange(id string, oldValue, newValue interface{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnRefChange(id, oldValue, newValue)
		}
	}
}

// OnEvent forwards to all hooks.
func (c *CompositeHook) OnEvent(componentID, eventName string, data interface{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnEvent(componentID, eventName, data)
		}
	}
}

// OnRenderComplete forwards to all hooks.
func (c *CompositeHook) OnRenderComplete(componentID string, duration time.Duration) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnRenderComplete(componentID, duration)
		}
	}
}

// OnComputedChange forwards to all hooks.
func (c *CompositeHook) OnComputedChange(id string, oldValue, newValue interface{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnComputedChange(id, oldValue, newValue)
		}
	}
}

// OnWatchCallback forwards to all hooks.
func (c *CompositeHook) OnWatchCallback(watcherID string, newValue, oldValue interface{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnWatchCallback(watcherID, newValue, oldValue)
		}
	}
}

// OnEffectRun forwards to all hooks.
func (c *CompositeHook) OnEffectRun(effectID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnEffectRun(effectID)
		}
	}
}

// OnChildAdded forwards to all hooks.
func (c *CompositeHook) OnChildAdded(parentID, childID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnChildAdded(parentID, childID)
		}
	}
}

// OnChildRemoved forwards to all hooks.
func (c *CompositeHook) OnChildRemoved(parentID, childID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnChildRemoved(parentID, childID)
		}
	}
}

// OnRefExposed forwards to all hooks.
func (c *CompositeHook) OnRefExposed(componentID, refID, refName string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.hooks {
		if h != nil {
			h.OnRefExposed(componentID, refID, refName)
		}
	}
}
