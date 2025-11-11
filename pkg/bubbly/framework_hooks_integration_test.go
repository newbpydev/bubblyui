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

// Task 8.8: Integration tests for Ref → Watch and Computed → Watch cascades

// TestFrameworkHooks_RefWatch verifies hooks are called when Ref watchers execute
func TestFrameworkHooks_RefWatch(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref and watch it
	ref := NewRef(10)

	watcherCalled := false
	Watch(ref, func(newVal, oldVal int) {
		watcherCalled = true
	})

	// Change ref value - should trigger watch callback hook
	ref.Set(20)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.watchCalls.Load())
	hook.mu.RLock()
	assert.Contains(t, hook.lastWatchID, "watch-0x")
	assert.Equal(t, 20, hook.lastWatchNew)
	assert.Equal(t, 10, hook.lastWatchOld)
	hook.mu.RUnlock()

	// Verify watcher callback was also called
	assert.True(t, watcherCalled)
}

// TestFrameworkHooks_ComputedWatch verifies hooks are called when Computed watchers execute
func TestFrameworkHooks_ComputedWatch(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref and computed value
	ref := NewRef(10)
	computed := NewComputed(func() int {
		return ref.Get().(int) * 2
	})

	watcherCalled := false
	Watch(computed, func(newVal, oldVal int) {
		watcherCalled = true
	})

	// Initial computation
	computed.Get()

	// Change ref value - should trigger computed change AND watch callback hooks
	ref.Set(15)

	// Verify watch callback hook was called
	assert.Equal(t, int32(1), hook.watchCalls.Load())
	hook.mu.RLock()
	assert.Contains(t, hook.lastWatchID, "watch-0x")
	assert.Equal(t, 30, hook.lastWatchNew)
	assert.Equal(t, 20, hook.lastWatchOld)
	hook.mu.RUnlock()

	// Verify watcher callback was also called
	assert.True(t, watcherCalled)
}

// TestFrameworkHooks_WatchWithImmediate verifies hooks fire when immediate watcher is triggered
func TestFrameworkHooks_WatchWithImmediate(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref
	ref := NewRef(10)

	// Watch with immediate option
	// Note: The immediate callback in Watch() bypasses notifyWatcher,
	// so hook won't fire until a value change triggers notifyWatcher
	Watch(ref, func(newVal, oldVal int) {
		// Watcher callback
	}, WithImmediate())

	// Hook not called yet (immediate callback bypasses notifyWatcher)
	assert.Equal(t, int32(0), hook.watchCalls.Load())

	// Now change the value - this will trigger notifyWatcher and the hook
	ref.Set(20)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.watchCalls.Load())
	hook.mu.RLock()
	assert.Contains(t, hook.lastWatchID, "watch-0x")
	assert.Equal(t, 20, hook.lastWatchNew)
	assert.Equal(t, 10, hook.lastWatchOld)
	hook.mu.RUnlock()
}

// TestFrameworkHooks_WatchWithDeep verifies hooks respect deep watching mode
func TestFrameworkHooks_WatchWithDeep(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	type User struct {
		Name string
		Age  int
	}

	// Create a ref with deep watching
	ref := NewRef(User{Name: "John", Age: 30})

	Watch(ref, func(newVal, oldVal User) {
		// Watcher callback
	}, WithDeep())

	// Set to same value - with deep watching, should NOT trigger callback
	ref.Set(User{Name: "John", Age: 30})

	// Hook is called BEFORE deep comparison, so it fires
	// But the callback itself won't execute due to deep equal check
	assert.Equal(t, int32(1), hook.watchCalls.Load())

	// Set to different value - should trigger callback
	ref.Set(User{Name: "Jane", Age: 31})

	// Hook called again
	assert.Equal(t, int32(2), hook.watchCalls.Load())
}

// TestFrameworkHooks_WatchFlushModes verifies hooks work with different flush modes
func TestFrameworkHooks_WatchFlushModes(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Test sync mode (default)
	ref1 := NewRef(10)
	Watch(ref1, func(newVal, oldVal int) {}, WithFlush("sync"))
	ref1.Set(20)

	assert.Equal(t, int32(1), hook.watchCalls.Load())

	// Test post mode (queued)
	ref2 := NewRef(10)
	Watch(ref2, func(newVal, oldVal int) {}, WithFlush("post"))
	ref2.Set(20)

	// Hook is called immediately (before queueing)
	assert.Equal(t, int32(2), hook.watchCalls.Load())

	// Flush queued callbacks
	FlushWatchers()

	// Hook count stays same (already called before queueing)
	assert.Equal(t, int32(2), hook.watchCalls.Load())
}

// TestFrameworkHooks_WatchThreadSafe verifies concurrent watch callbacks are safe
func TestFrameworkHooks_WatchThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create multiple refs with watchers
	ref1 := NewRef(0)
	ref2 := NewRef(0)

	Watch(ref1, func(newVal, oldVal int) {})
	Watch(ref2, func(newVal, oldVal int) {})

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
	assert.Equal(t, int32(100), hook.watchCalls.Load())
}

// TestFrameworkHooks_FullCascade verifies complete Ref → Computed → Watch cascade
func TestFrameworkHooks_FullCascade(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create Ref → Computed → Watch cascade
	ref := NewRef(10)
	computed := NewComputed(func() int {
		return ref.Get().(int) * 2
	})

	Watch(computed, func(newVal, oldVal int) {
		// Watcher callback
	})

	// Initial computation
	computed.Get()

	// Reset counters to track only the cascade
	hook.refChangeCalls.Store(0)
	hook.computedCalls.Store(0)
	hook.watchCalls.Store(0)

	// Trigger cascade: Ref change → Computed recompute → Watch callback
	ref.Set(15)

	// Verify all hooks in cascade were called
	assert.Equal(t, int32(1), hook.refChangeCalls.Load(), "Ref change hook should fire")
	assert.Equal(t, int32(1), hook.computedCalls.Load(), "Computed change hook should fire")
	assert.Equal(t, int32(1), hook.watchCalls.Load(), "Watch callback hook should fire")

	// Verify correct values in watch callback
	hook.mu.RLock()
	assert.Equal(t, 30, hook.lastWatchNew)
	assert.Equal(t, 20, hook.lastWatchOld)
	hook.mu.RUnlock()
}

// Task 8.9: Integration tests for WatchEffect instrumentation

// TestFrameworkHooks_EffectRun_InitialRun verifies hook fires on initial effect run
func TestFrameworkHooks_EffectRun_InitialRun(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a WatchEffect
	effectRan := false
	cleanup := WatchEffect(func() {
		effectRan = true
	})
	defer cleanup()

	// Verify hook was called for initial run
	assert.Equal(t, int32(1), hook.effectCalls.Load())
	hook.mu.RLock()
	assert.Contains(t, hook.lastEffectID, "effect-0x")
	hook.mu.RUnlock()

	// Verify effect actually ran
	assert.True(t, effectRan)
}

// TestFrameworkHooks_EffectRun_DependencyChange verifies hook fires on dependency changes
func TestFrameworkHooks_EffectRun_DependencyChange(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref
	ref := NewRef(10)

	// Create a WatchEffect that depends on the ref
	runCount := 0
	cleanup := WatchEffect(func() {
		_ = ref.Get()
		runCount++
	})
	defer cleanup()

	// Initial run
	assert.Equal(t, int32(1), hook.effectCalls.Load())
	assert.Equal(t, 1, runCount)

	// Change ref value - should trigger effect re-run
	ref.Set(20)

	// Verify hook was called again
	assert.Equal(t, int32(2), hook.effectCalls.Load())
	assert.Equal(t, 2, runCount)

	// Change again
	ref.Set(30)

	// Verify hook called third time
	assert.Equal(t, int32(3), hook.effectCalls.Load())
	assert.Equal(t, 3, runCount)
}

// TestFrameworkHooks_EffectRun_MultipleDependencies verifies hook fires for multiple dependency changes
func TestFrameworkHooks_EffectRun_MultipleDependencies(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create multiple refs
	ref1 := NewRef(10)
	ref2 := NewRef(20)

	// Create a WatchEffect that depends on both refs
	runCount := 0
	cleanup := WatchEffect(func() {
		_ = ref1.Get()
		_ = ref2.Get()
		runCount++
	})
	defer cleanup()

	// Initial run
	assert.Equal(t, int32(1), hook.effectCalls.Load())

	// Change first ref
	ref1.Set(15)
	assert.Equal(t, int32(2), hook.effectCalls.Load())

	// Change second ref
	ref2.Set(25)
	assert.Equal(t, int32(3), hook.effectCalls.Load())

	// Change both (should trigger twice)
	ref1.Set(100)
	ref2.Set(200)
	assert.Equal(t, int32(5), hook.effectCalls.Load())
}

// TestFrameworkHooks_EffectRun_EffectIDFormat verifies effect ID format is correct
func TestFrameworkHooks_EffectRun_EffectIDFormat(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a WatchEffect
	cleanup := WatchEffect(func() {
		// Effect body
	})
	defer cleanup()

	// Verify effect ID format: "effect-0xHEX"
	hook.mu.RLock()
	effectID := hook.lastEffectID
	hook.mu.RUnlock()

	assert.Contains(t, effectID, "effect-0x")
	assert.Greater(t, len(effectID), len("effect-0x"))
}

// TestFrameworkHooks_EffectRun_StoppedEffect verifies hook doesn't fire when effect stopped
func TestFrameworkHooks_EffectRun_StoppedEffect(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create a ref
	ref := NewRef(10)

	// Create a WatchEffect
	cleanup := WatchEffect(func() {
		_ = ref.Get()
	})

	// Initial run
	assert.Equal(t, int32(1), hook.effectCalls.Load())

	// Stop the effect
	cleanup()

	// Change ref - should NOT trigger effect
	ref.Set(20)

	// Verify hook was NOT called again
	assert.Equal(t, int32(1), hook.effectCalls.Load())
}

// TestFrameworkHooks_EffectRun_NoHook verifies no panic when no hook registered
func TestFrameworkHooks_EffectRun_NoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Create a ref
	ref := NewRef(10)

	// Create a WatchEffect without hook registered
	runCount := 0
	cleanup := WatchEffect(func() {
		_ = ref.Get()
		runCount++
	})
	defer cleanup()

	// Should not panic
	assert.Equal(t, 1, runCount)

	// Change ref - should not panic
	ref.Set(20)
	assert.Equal(t, 2, runCount)
}

// TestFrameworkHooks_EffectRun_ThreadSafe verifies concurrent effect runs are safe
func TestFrameworkHooks_EffectRun_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create multiple refs
	ref1 := NewRef(0)
	ref2 := NewRef(0)

	// Create multiple effects
	cleanup1 := WatchEffect(func() {
		_ = ref1.Get()
	})
	defer cleanup1()

	cleanup2 := WatchEffect(func() {
		_ = ref2.Get()
	})
	defer cleanup2()

	// Initial runs (2 effects)
	initialCalls := hook.effectCalls.Load()
	assert.Equal(t, int32(2), initialCalls)

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

	// Verify hooks were called (should be 2 initial + 100 updates = 102)
	assert.Equal(t, int32(102), hook.effectCalls.Load())
}

// TestFrameworkHooks_EffectRun_RefComputedEffectCascade verifies Ref → Computed → Effect cascade
func TestFrameworkHooks_EffectRun_RefComputedEffectCascade(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create Ref → Computed → Effect cascade
	ref := NewRef(10)
	computed := NewComputed(func() int {
		return ref.Get().(int) * 2
	})

	effectRuns := 0
	cleanup := WatchEffect(func() {
		_ = computed.Get()
		effectRuns++
	})
	defer cleanup()

	// Initial run
	assert.Equal(t, int32(1), hook.effectCalls.Load())
	assert.Equal(t, 1, effectRuns)

	// Reset counters to track only the cascade
	hook.refChangeCalls.Store(0)
	hook.computedCalls.Store(0)
	hook.effectCalls.Store(0)

	// Trigger cascade: Ref change → Computed invalidation → Effect re-run
	ref.Set(15)

	// Verify hooks in cascade were called
	assert.Equal(t, int32(1), hook.refChangeCalls.Load(), "Ref change hook should fire")
	// Note: Computed change hook does NOT fire because WatchEffect accesses computed
	// directly without using Watch(), so computed doesn't recompute until accessed
	// in the effect. The computed is invalidated but not recomputed yet.
	assert.Equal(t, int32(1), hook.effectCalls.Load(), "Effect run hook should fire")

	// Verify effect actually re-ran
	assert.Equal(t, 2, effectRuns)
}

// TestFrameworkHooks_EffectRun_ConditionalDependencies verifies hook fires for conditional dependencies
func TestFrameworkHooks_EffectRun_ConditionalDependencies(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create refs for conditional dependencies
	toggle := NewRef(true)
	valueA := NewRef(1)
	valueB := NewRef(100)

	// Create effect with conditional dependencies
	cleanup := WatchEffect(func() {
		if toggle.Get().(bool) {
			_ = valueA.Get() // Only tracks valueA when toggle is true
		} else {
			_ = valueB.Get() // Only tracks valueB when toggle is false
		}
	})
	defer cleanup()

	// Initial run (effect runs once)
	initialCalls := hook.effectCalls.Load()
	assert.Equal(t, int32(1), initialCalls)

	// Change valueA - should trigger (toggle is true, tracking valueA and toggle)
	valueA.Set(2)
	assert.Greater(t, hook.effectCalls.Load(), initialCalls, "valueA change should trigger effect")

	// Reset counter
	beforeToggle := hook.effectCalls.Load()

	// Change valueB - should NOT trigger (toggle is true, not tracking valueB)
	valueB.Set(200)
	assert.Equal(t, beforeToggle, hook.effectCalls.Load(), "valueB change should NOT trigger when toggle is true")

	// Toggle to false - should trigger (tracking toggle)
	toggle.Set(false)
	assert.Greater(t, hook.effectCalls.Load(), beforeToggle, "toggle change should trigger effect")

	// Reset counter
	afterToggle := hook.effectCalls.Load()

	// Now change valueB - should trigger (toggle is false, now tracking valueB)
	valueB.Set(300)
	assert.Greater(t, hook.effectCalls.Load(), afterToggle, "valueB change should trigger when toggle is false")
}
