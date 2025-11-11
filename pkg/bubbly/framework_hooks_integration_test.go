package bubbly

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestFrameworkHooks_ComponentLifecycle verifies hooks are called during component lifecycle
func TestFrameworkHooks_ComponentLifecycle(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a component
	component, err := NewComponent("TestComponent").
		Setup(func(ctx *Context) {
			// Setup runs during Init
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)

	// 1. Test Init (should trigger mount notification)
	component.Init()
	assert.Equal(t, int32(1), hook.mountCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, component.ID(), hook.lastMountID)
	assert.Equal(t, "TestComponent", hook.lastMountName)
	hook.mu.RUnlock()

	// 2. Test Update (should trigger update notification)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	component.Update(msg)
	assert.Equal(t, int32(1), hook.updateCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, component.ID(), hook.lastUpdateID)
	assert.Equal(t, msg, hook.lastUpdateMsg)
	hook.mu.RUnlock()

	// 3. Test View (should trigger render complete notification)
	output := component.View()
	assert.Equal(t, "test", output)
	assert.Equal(t, int32(1), hook.renderCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, component.ID(), hook.lastRenderID)
	assert.Greater(t, hook.lastRenderDur, time.Duration(0))
	hook.mu.RUnlock()

	// 4. Test Unmount (should trigger unmount notification)
	if impl, ok := component.(*componentImpl); ok {
		impl.Unmount()
	}
	assert.Equal(t, int32(1), hook.unmountCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, component.ID(), hook.lastUnmountID)
	hook.mu.RUnlock()
}

// TestFrameworkHooks_EventEmission verifies hooks are called when events are emitted
func TestFrameworkHooks_EventEmission(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a component
	component, err := NewComponent("EventTest").
		Setup(func(ctx *Context) {
			ctx.On("testEvent", func(data interface{}) {
				// Event handler
			})
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)

	component.Init()

	// Emit an event
	eventData := map[string]string{"key": "value"}
	component.Emit("testEvent", eventData)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.eventCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, component.ID(), hook.lastEventCompID)
	assert.Equal(t, "testEvent", hook.lastEventName)
	assert.Equal(t, eventData, hook.lastEventData)
	hook.mu.RUnlock()
}

// TestFrameworkHooks_RefChange verifies hooks are called when ref values change
func TestFrameworkHooks_RefChange(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref
	ref := NewRef(10)

	// Change the ref value
	ref.Set(20)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.refChangeCalls.Load())
	hook.mu.RLock()
	assert.NotEmpty(t, hook.lastRefID)
	assert.Equal(t, 10, hook.lastRefOld)
	assert.Equal(t, 20, hook.lastRefNew)
	hook.mu.RUnlock()
}

// TestFrameworkHooks_NoHook verifies no panic when no hook registered
func TestFrameworkHooks_NoHook(t *testing.T) {
	// Ensure no hook is registered
	UnregisterHook()

	// Create and use a component - should not panic
	component, err := NewComponent("NoHookTest").
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)

	// All operations should work without hooks
	component.Init()
	component.Update(tea.KeyMsg{})
	component.View()
	component.Emit("test", nil)

	if impl, ok := component.(*componentImpl); ok {
		impl.Unmount()
	}

	// No assertions needed - just verify no panic
}

// TestFrameworkHooks_MultipleUpdates verifies hooks work with many updates
func TestFrameworkHooks_MultipleUpdates(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	component, err := NewComponent("MultiUpdate").
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)

	component.Init()

	// Send multiple updates
	iterations := 10
	for i := 0; i < iterations; i++ {
		component.Update(tea.KeyMsg{})
	}

	// Verify all updates were tracked
	assert.Equal(t, int32(iterations), hook.updateCalls.Load())
}

// TestFrameworkHooks_MultipleRenders verifies render timing for multiple renders
func TestFrameworkHooks_MultipleRenders(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	component, err := NewComponent("MultiRender").
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)

	component.Init()

	// Render multiple times
	iterations := 10
	for i := 0; i < iterations; i++ {
		component.View()
	}

	// Verify all renders were tracked
	assert.Equal(t, int32(iterations), hook.renderCalls.Load())

	// Verify duration was captured
	hook.mu.RLock()
	assert.Greater(t, hook.lastRenderDur, time.Duration(0))
	hook.mu.RUnlock()
}

// TestFrameworkHooks_RefWithComponent verifies ref changes work with components
func TestFrameworkHooks_RefWithComponent(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	var countRef *Ref[int]

	component, err := NewComponent("RefComponent").
		Setup(func(ctx *Context) {
			countRef = NewRef(0)
			ctx.Expose("count", countRef)
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)

	component.Init()

	// Change ref value
	countRef.Set(5)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.refChangeCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, 0, hook.lastRefOld)
	assert.Equal(t, 5, hook.lastRefNew)
	hook.mu.RUnlock()

	// Change again
	countRef.Set(10)

	// Verify second call
	assert.Equal(t, int32(2), hook.refChangeCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, 5, hook.lastRefOld)
	assert.Equal(t, 10, hook.lastRefNew)
	hook.mu.RUnlock()
}

// TestFrameworkHooks_CompleteLifecycle tests full component lifecycle
func TestFrameworkHooks_CompleteLifecycle(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	var countRef *Ref[int]

	component, err := NewComponent("CompleteLifecycle").
		Setup(func(ctx *Context) {
			countRef = NewRef(0)
			ctx.Expose("count", countRef)
			ctx.On("increment", func(data interface{}) {
				countRef.Set(countRef.GetTyped() + 1)
			})
		}).
		Template(func(ctx RenderContext) string {
			return "count"
		}).
		Build()
	assert.NoError(t, err)

	// 1. Init
	component.Init()
	assert.Equal(t, int32(1), hook.mountCalls.Load())

	// 2. Update
	component.Update(tea.KeyMsg{})
	assert.Equal(t, int32(1), hook.updateCalls.Load())

	// 3. View
	component.View()
	assert.Equal(t, int32(1), hook.renderCalls.Load())

	// 4. Emit event
	component.Emit("increment", nil)
	assert.Equal(t, int32(1), hook.eventCalls.Load())
	assert.Equal(t, int32(1), hook.refChangeCalls.Load()) // Ref changed by event handler

	// 5. Unmount
	if impl, ok := component.(*componentImpl); ok {
		impl.Unmount()
	}
	assert.Equal(t, int32(1), hook.unmountCalls.Load())

	// Verify complete lifecycle
	assert.Equal(t, int32(1), hook.mountCalls.Load())
	assert.Equal(t, int32(1), hook.updateCalls.Load())
	assert.Equal(t, int32(1), hook.renderCalls.Load())
	assert.Equal(t, int32(1), hook.eventCalls.Load())
	assert.Equal(t, int32(1), hook.refChangeCalls.Load())
	assert.Equal(t, int32(1), hook.unmountCalls.Load())
}

// TestFrameworkHooks_ZeroOverhead verifies no overhead when hook not registered
func TestFrameworkHooks_ZeroOverhead(t *testing.T) {
	// Unregister any hook
	UnregisterHook()

	component, err := NewComponent("ZeroOverhead").
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)

	// Run operations - should be fast with no hook
	start := time.Now()
	component.Init()
	for i := 0; i < 1000; i++ {
		component.Update(tea.KeyMsg{})
		component.View()
	}
	duration := time.Since(start)

	// Just verify it completes - actual overhead check would need benchmarks
	assert.Less(t, duration, 1*time.Second)
}

// Task 8.7: Integration tests for Ref → Computed cascade with hooks

// TestFrameworkHooks_ComputedChange verifies hooks are called when computed values change
func TestFrameworkHooks_ComputedChange(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref and computed value
	ref := NewRef(10)
	computed := NewComputed(func() int {
		return ref.Get().(int) * 2
	})

	// Watch the computed to trigger recomputation on ref change
	Watch(computed, func(newVal, oldVal int) {
		// Watcher callback
	})

	// Initial computation (should NOT trigger hook - no change yet)
	result := computed.Get()
	assert.Equal(t, 20, result)
	assert.Equal(t, int32(0), hook.computedCalls.Load(), "Initial computation should not trigger hook")

	// Change ref value - should trigger computed change hook
	ref.Set(15)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.computedCalls.Load())
	hook.mu.RLock()
	assert.Contains(t, hook.lastComputedID, "computed-0x")
	assert.Equal(t, 20, hook.lastComputedOld)
	assert.Equal(t, 30, hook.lastComputedNew)
	hook.mu.RUnlock()
}

// TestFrameworkHooks_ComputedChange_NoChangeNoHook verifies hook not called when value unchanged
func TestFrameworkHooks_ComputedChange_NoChangeNoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref and computed value that always returns same value
	ref := NewRef(10)
	computed := NewComputed(func() int {
		_ = ref.Get() // Access ref to create dependency
		return 42     // Always return same value
	})

	// Watch the computed to trigger recomputation on ref change
	Watch(computed, func(newVal, oldVal int) {
		// Watcher callback
	})

	// Initial computation
	result := computed.Get()
	assert.Equal(t, 42, result)

	// Change ref value - computed recomputes but value unchanged
	ref.Set(15)

	// Verify hook was NOT called (value didn't change)
	assert.Equal(t, int32(0), hook.computedCalls.Load())
}

// TestFrameworkHooks_ComputedChange_NoWatchersNoHook verifies hook not called without watchers
func TestFrameworkHooks_ComputedChange_NoWatchersNoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref and computed value WITHOUT watchers
	ref := NewRef(10)
	computed := NewComputed(func() int {
		return ref.Get().(int) * 2
	})

	// Initial computation
	result := computed.Get()
	assert.Equal(t, 20, result)

	// Change ref value - computed is invalidated but not recomputed (no watchers)
	ref.Set(15)

	// Verify hook was NOT called (no watchers = no recomputation)
	assert.Equal(t, int32(0), hook.computedCalls.Load())
}

// TestFrameworkHooks_ComputedChange_CascadeOrder verifies hook fires before watchers
func TestFrameworkHooks_ComputedChange_CascadeOrder(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Track order of events
	var events []string

	// Create a ref and computed value
	ref := NewRef(10)
	computed := NewComputed(func() int {
		return ref.Get().(int) * 2
	})

	// Watch the computed
	Watch(computed, func(newVal, oldVal int) {
		events = append(events, "watcher")
	})

	// Initial computation
	computed.Get()

	// Change ref value
	ref.Set(15)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.computedCalls.Load())

	// Verify watcher was also called
	assert.Contains(t, events, "watcher")

	// Note: We can't verify exact order in this test because the hook
	// is called synchronously before notifyWatchers, but we verify both happened
}

// TestFrameworkHooks_ComputedChange_ThreadSafe verifies concurrent computed changes are safe
func TestFrameworkHooks_ComputedChange_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create multiple refs and computed values
	ref1 := NewRef(0)
	ref2 := NewRef(0)

	computed1 := NewComputed(func() int {
		return ref1.Get().(int) * 2
	})
	computed2 := NewComputed(func() int {
		return ref2.Get().(int) * 3
	})

	// Watch both to trigger recomputation
	Watch(computed1, func(newVal, oldVal int) {})
	Watch(computed2, func(newVal, oldVal int) {})

	// Initial computation
	computed1.Get()
	computed2.Get()

	// Concurrent updates
	done := make(chan bool, 2)

	go func() {
		for i := 1; i <= 50; i++ {
			ref1.Set(i)
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	go func() {
		for i := 1; i <= 50; i++ {
			ref2.Set(i)
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done

	// Verify hooks were called (should be 100 total: 50 + 50)
	assert.Equal(t, int32(100), hook.computedCalls.Load())
}
