package bubbly

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockHook is a test implementation of FrameworkHook
type mockHook struct {
	mountCalls           atomic.Int32
	updateCalls          atomic.Int32
	unmountCalls         atomic.Int32
	refChangeCalls       atomic.Int32
	eventCalls           atomic.Int32
	renderCalls          atomic.Int32
	computedCalls        atomic.Int32
	watchCalls           atomic.Int32
	effectCalls          atomic.Int32
	childAddedCalls      atomic.Int32
	childRemovedCalls    atomic.Int32
	refExposedCalls      atomic.Int32
	lastMountID          string
	lastMountName        string
	lastUpdateID         string
	lastUpdateMsg        interface{}
	lastUnmountID        string
	lastRefID            string
	lastRefOld           interface{}
	lastRefNew           interface{}
	lastEventCompID      string
	lastEventName        string
	lastEventData        interface{}
	lastRenderID         string
	lastRenderDur        time.Duration
	lastComputedID       string
	lastComputedOld      interface{}
	lastComputedNew      interface{}
	lastWatchID          string
	lastWatchNew         interface{}
	lastWatchOld         interface{}
	lastEffectID         string
	lastParentID         string
	lastChildID          string
	lastRefExposedCompID string
	lastRefExposedRefID  string
	lastRefExposedName   string
	mu                   sync.RWMutex
}

func (m *mockHook) OnComponentMount(id, name string) {
	m.mountCalls.Add(1)
	m.mu.Lock()
	m.lastMountID = id
	m.lastMountName = name
	m.mu.Unlock()
}

func (m *mockHook) OnComponentUpdate(id string, msg interface{}) {
	m.updateCalls.Add(1)
	m.mu.Lock()
	m.lastUpdateID = id
	m.lastUpdateMsg = msg
	m.mu.Unlock()
}

func (m *mockHook) OnComponentUnmount(id string) {
	m.unmountCalls.Add(1)
	m.mu.Lock()
	m.lastUnmountID = id
	m.mu.Unlock()
}

func (m *mockHook) OnRefChange(id string, oldValue, newValue interface{}) {
	m.refChangeCalls.Add(1)
	m.mu.Lock()
	m.lastRefID = id
	m.lastRefOld = oldValue
	m.lastRefNew = newValue
	m.mu.Unlock()
}

func (m *mockHook) OnEvent(componentID, eventName string, data interface{}) {
	m.eventCalls.Add(1)
	m.mu.Lock()
	m.lastEventCompID = componentID
	m.lastEventName = eventName
	m.lastEventData = data
	m.mu.Unlock()
}

func (m *mockHook) OnRenderComplete(componentID string, duration time.Duration) {
	m.renderCalls.Add(1)
	m.mu.Lock()
	m.lastRenderID = componentID
	m.lastRenderDur = duration
	m.mu.Unlock()
}

func (m *mockHook) OnComputedChange(id string, oldValue, newValue interface{}) {
	m.computedCalls.Add(1)
	m.mu.Lock()
	m.lastComputedID = id
	m.lastComputedOld = oldValue
	m.lastComputedNew = newValue
	m.mu.Unlock()
}

func (m *mockHook) OnWatchCallback(watcherID string, newValue, oldValue interface{}) {
	m.watchCalls.Add(1)
	m.mu.Lock()
	m.lastWatchID = watcherID
	m.lastWatchNew = newValue
	m.lastWatchOld = oldValue
	m.mu.Unlock()
}

func (m *mockHook) OnEffectRun(effectID string) {
	m.effectCalls.Add(1)
	m.mu.Lock()
	m.lastEffectID = effectID
	m.mu.Unlock()
}

func (m *mockHook) OnChildAdded(parentID, childID string) {
	m.childAddedCalls.Add(1)
	m.mu.Lock()
	m.lastParentID = parentID
	m.lastChildID = childID
	m.mu.Unlock()
}

func (m *mockHook) OnChildRemoved(parentID, childID string) {
	m.childRemovedCalls.Add(1)
	m.mu.Lock()
	m.lastParentID = parentID
	m.lastChildID = childID
	m.mu.Unlock()
}

func (m *mockHook) OnRefExposed(componentID, refID, refName string) {
	m.refExposedCalls.Add(1)
	m.mu.Lock()
	m.lastRefExposedCompID = componentID
	m.lastRefExposedRefID = refID
	m.lastRefExposedName = refName
	m.mu.Unlock()
}

func TestRegisterHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	err := RegisterHook(hook)

	assert.NoError(t, err)
	assert.True(t, IsHookRegistered())
}

func TestUnregisterHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)
	assert.True(t, IsHookRegistered())

	err := UnregisterHook()
	assert.NoError(t, err)
	assert.False(t, IsHookRegistered())
}

func TestRegisterHook_Replaces(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook1 := &mockHook{}
	hook2 := &mockHook{}

	RegisterHook(hook1)
	notifyHookComponentMount("comp-1", "Test")
	assert.Equal(t, int32(1), hook1.mountCalls.Load())
	assert.Equal(t, int32(0), hook2.mountCalls.Load())

	// Replace with hook2
	RegisterHook(hook2)
	notifyHookComponentMount("comp-2", "Test2")
	assert.Equal(t, int32(1), hook1.mountCalls.Load()) // No new calls
	assert.Equal(t, int32(1), hook2.mountCalls.Load()) // Gets the call
}

func TestIsHookRegistered(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	assert.False(t, IsHookRegistered())

	hook := &mockHook{}
	RegisterHook(hook)
	assert.True(t, IsHookRegistered())

	UnregisterHook()
	assert.False(t, IsHookRegistered())
}

func TestNotifyHookComponentMount(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	notifyHookComponentMount("comp-123", "Counter")

	assert.Equal(t, int32(1), hook.mountCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "comp-123", hook.lastMountID)
	assert.Equal(t, "Counter", hook.lastMountName)
	hook.mu.RUnlock()
}

func TestNotifyHookComponentUpdate(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	msg := "test message"
	notifyHookComponentUpdate("comp-456", msg)

	assert.Equal(t, int32(1), hook.updateCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "comp-456", hook.lastUpdateID)
	assert.Equal(t, msg, hook.lastUpdateMsg)
	hook.mu.RUnlock()
}

func TestNotifyHookComponentUnmount(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	notifyHookComponentUnmount("comp-789")

	assert.Equal(t, int32(1), hook.unmountCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "comp-789", hook.lastUnmountID)
	hook.mu.RUnlock()
}

func TestNotifyHookRefChange(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	notifyHookRefChange("ref-0x123", 10, 20)

	assert.Equal(t, int32(1), hook.refChangeCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "ref-0x123", hook.lastRefID)
	assert.Equal(t, 10, hook.lastRefOld)
	assert.Equal(t, 20, hook.lastRefNew)
	hook.mu.RUnlock()
}

func TestNotifyHookEvent(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	data := map[string]string{"key": "value"}
	notifyHookEvent("comp-999", "click", data)

	assert.Equal(t, int32(1), hook.eventCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "comp-999", hook.lastEventCompID)
	assert.Equal(t, "click", hook.lastEventName)
	assert.Equal(t, data, hook.lastEventData)
	hook.mu.RUnlock()
}

func TestNotifyHookRenderComplete(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	duration := 100 * time.Millisecond
	notifyHookRenderComplete("comp-111", duration)

	assert.Equal(t, int32(1), hook.renderCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "comp-111", hook.lastRenderID)
	assert.Equal(t, duration, hook.lastRenderDur)
	hook.mu.RUnlock()
}

func TestNotifyHooks_NoHookRegistered(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Should not panic when no hook registered
	notifyHookComponentMount("comp-1", "Test")
	notifyHookComponentUpdate("comp-1", nil)
	notifyHookComponentUnmount("comp-1")
	notifyHookRefChange("ref-1", nil, nil)
	notifyHookEvent("comp-1", "test", nil)
	notifyHookRenderComplete("comp-1", time.Millisecond)

	// No assertions needed - just verify no panic
}

func TestHooks_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Concurrent calls to notify functions
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(6)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookComponentMount("comp", "Test")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookComponentUpdate("comp", nil)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookComponentUnmount("comp")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookRefChange("ref", nil, nil)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookEvent("comp", "event", nil)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookRenderComplete("comp", time.Millisecond)
		}
	}()

	wg.Wait()

	// Verify all calls were made
	assert.Equal(t, int32(iterations), hook.mountCalls.Load())
	assert.Equal(t, int32(iterations), hook.updateCalls.Load())
	assert.Equal(t, int32(iterations), hook.unmountCalls.Load())
	assert.Equal(t, int32(iterations), hook.refChangeCalls.Load())
	assert.Equal(t, int32(iterations), hook.eventCalls.Load())
	assert.Equal(t, int32(iterations), hook.renderCalls.Load())
}

func TestHookRegistration_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Concurrent register/unregister operations
	var wg sync.WaitGroup
	iterations := 50

	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			hook := &mockHook{}
			RegisterHook(hook)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			UnregisterHook()
			time.Sleep(time.Microsecond) // Small delay
		}
	}()

	wg.Wait()

	// No assertions needed - just verify no races or panics
}

// Task 8.7: Tests for OnComputedChange hook

func TestNotifyHookComputedChange(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Notify computed change
	notifyHookComputedChange("computed-0x123", 10, 20)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.computedCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "computed-0x123", hook.lastComputedID)
	assert.Equal(t, 10, hook.lastComputedOld)
	assert.Equal(t, 20, hook.lastComputedNew)
	hook.mu.RUnlock()
}

func TestNotifyHookComputedChange_NoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Should not panic when no hook registered
	notifyHookComputedChange("computed-0x123", 10, 20)
}

func TestNotifyHookComputedChange_MultipleValues(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	tests := []struct {
		name     string
		id       string
		oldValue interface{}
		newValue interface{}
	}{
		{
			name:     "int values",
			id:       "computed-0x1",
			oldValue: 10,
			newValue: 20,
		},
		{
			name:     "string values",
			id:       "computed-0x2",
			oldValue: "old",
			newValue: "new",
		},
		{
			name:     "struct values",
			id:       "computed-0x3",
			oldValue: struct{ X int }{X: 1},
			newValue: struct{ X int }{X: 2},
		},
		{
			name:     "nil to value",
			id:       "computed-0x4",
			oldValue: nil,
			newValue: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset call count
			hook.computedCalls.Store(0)

			notifyHookComputedChange(tt.id, tt.oldValue, tt.newValue)

			assert.Equal(t, int32(1), hook.computedCalls.Load())
			hook.mu.RLock()
			assert.Equal(t, tt.id, hook.lastComputedID)
			assert.Equal(t, tt.oldValue, hook.lastComputedOld)
			assert.Equal(t, tt.newValue, hook.lastComputedNew)
			hook.mu.RUnlock()
		})
	}
}

func TestNotifyHookComputedChange_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Concurrent computed change notifications
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookComputedChange("computed-0x123", i, i+1)
		}
	}()

	wg.Wait()

	// Verify all calls were made
	assert.Equal(t, int32(iterations), hook.computedCalls.Load())
}

// Task 8.8: Tests for OnWatchCallback hook

func TestNotifyHookWatchCallback(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Notify watch callback
	notifyHookWatchCallback("watch-0x123", 20, 10)

	// Verify hook was called
	assert.Equal(t, int32(1), hook.watchCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "watch-0x123", hook.lastWatchID)
	assert.Equal(t, 20, hook.lastWatchNew)
	assert.Equal(t, 10, hook.lastWatchOld)
	hook.mu.RUnlock()
}

func TestNotifyHookWatchCallback_NoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Should not panic when no hook registered
	notifyHookWatchCallback("watch-0x123", 20, 10)
}

func TestNotifyHookWatchCallback_MultipleValues(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	tests := []struct {
		name     string
		id       string
		newValue interface{}
		oldValue interface{}
	}{
		{
			name:     "int values",
			id:       "watch-0x1",
			newValue: 20,
			oldValue: 10,
		},
		{
			name:     "string values",
			id:       "watch-0x2",
			newValue: "new",
			oldValue: "old",
		},
		{
			name:     "struct values",
			id:       "watch-0x3",
			newValue: struct{ X int }{X: 2},
			oldValue: struct{ X int }{X: 1},
		},
		{
			name:     "nil to value",
			id:       "watch-0x4",
			newValue: 42,
			oldValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset call count
			hook.watchCalls.Store(0)

			notifyHookWatchCallback(tt.id, tt.newValue, tt.oldValue)

			assert.Equal(t, int32(1), hook.watchCalls.Load())
			hook.mu.RLock()
			assert.Equal(t, tt.id, hook.lastWatchID)
			assert.Equal(t, tt.newValue, hook.lastWatchNew)
			assert.Equal(t, tt.oldValue, hook.lastWatchOld)
			hook.mu.RUnlock()
		})
	}
}

func TestNotifyHookWatchCallback_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Concurrent watch callback notifications
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookWatchCallback("watch-0x123", i+1, i)
		}
	}()

	wg.Wait()

	// Verify all calls were made
	assert.Equal(t, int32(iterations), hook.watchCalls.Load())
}

// Task 8.9: Tests for OnEffectRun hook

func TestNotifyHookEffectRun(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Notify effect run
	notifyHookEffectRun("effect-0x123")

	// Verify hook was called
	assert.Equal(t, int32(1), hook.effectCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "effect-0x123", hook.lastEffectID)
	hook.mu.RUnlock()
}

func TestNotifyHookEffectRun_NoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Should not panic when no hook registered
	notifyHookEffectRun("effect-0x123")
}

func TestNotifyHookEffectRun_MultipleEffects(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	tests := []struct {
		name     string
		effectID string
	}{
		{
			name:     "effect 1",
			effectID: "effect-0x1",
		},
		{
			name:     "effect 2",
			effectID: "effect-0x2",
		},
		{
			name:     "effect 3",
			effectID: "effect-0x3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset call count
			hook.effectCalls.Store(0)

			notifyHookEffectRun(tt.effectID)

			assert.Equal(t, int32(1), hook.effectCalls.Load())
			hook.mu.RLock()
			assert.Equal(t, tt.effectID, hook.lastEffectID)
			hook.mu.RUnlock()
		})
	}
}

func TestNotifyHookEffectRun_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Concurrent effect run notifications
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookEffectRun("effect-0x123")
		}
	}()

	wg.Wait()

	// Verify all calls were made
	assert.Equal(t, int32(iterations), hook.effectCalls.Load())
}

// Task 8.10: Tests for OnChildAdded and OnChildRemoved hooks

func TestNotifyHookChildAdded(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Notify child added
	notifyHookChildAdded("parent-123", "child-456")

	// Verify hook was called
	assert.Equal(t, int32(1), hook.childAddedCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "parent-123", hook.lastParentID)
	assert.Equal(t, "child-456", hook.lastChildID)
	hook.mu.RUnlock()
}

func TestNotifyHookChildRemoved(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Notify child removed
	notifyHookChildRemoved("parent-789", "child-012")

	// Verify hook was called
	assert.Equal(t, int32(1), hook.childRemovedCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "parent-789", hook.lastParentID)
	assert.Equal(t, "child-012", hook.lastChildID)
	hook.mu.RUnlock()
}

func TestNotifyHookChildAdded_NoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Should not panic when no hook registered
	notifyHookChildAdded("parent-1", "child-1")
}

func TestNotifyHookChildRemoved_NoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Should not panic when no hook registered
	notifyHookChildRemoved("parent-1", "child-1")
}

func TestNotifyHookChildMutations_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Concurrent child mutation notifications
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookChildAdded("parent", "child")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookChildRemoved("parent", "child")
		}
	}()

	wg.Wait()

	// Verify all calls were made
	assert.Equal(t, int32(iterations), hook.childAddedCalls.Load())
	assert.Equal(t, int32(iterations), hook.childRemovedCalls.Load())
}

// CRITICAL FIX 1: Tests for OnRefExposed hook (TDD - these will FAIL initially)

func TestContext_Expose_NotifiesHookForRef(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	tests := []struct {
		name     string
		refName  string
		refValue interface{}
	}{
		{
			name:     "expose int ref",
			refName:  "count",
			refValue: NewRef(42),
		},
		{
			name:     "expose string ref",
			refName:  "message",
			refValue: NewRef("hello"),
		},
		{
			name:     "expose bool ref",
			refName:  "isActive",
			refValue: NewRef(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Register mock hook
			hook := &mockHook{}
			RegisterHook(hook)
			defer UnregisterHook()

			// Create component with context
			component := newComponentImpl("TestComponent")
			ctx := &Context{component: component}

			// Expose the ref - this should call OnRefExposed hook
			ctx.Expose(tt.refName, tt.refValue)

			// Verify hook was called
			assert.Equal(t, int32(1), hook.refExposedCalls.Load(),
				"OnRefExposed should be called when Ref is exposed")

			hook.mu.RLock()
			assert.Equal(t, component.id, hook.lastRefExposedCompID,
				"Component ID should match")
			assert.Equal(t, tt.refName, hook.lastRefExposedName,
				"Ref name should match")
			assert.NotEmpty(t, hook.lastRefExposedRefID,
				"Ref ID should not be empty")
			hook.mu.RUnlock()
		})
	}
}

func TestContext_Expose_DoesNotNotifyForNonRef(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "plain int",
			key:   "count",
			value: 42,
		},
		{
			name:  "plain string",
			key:   "message",
			value: "hello",
		},
		{
			name:  "plain struct",
			key:   "data",
			value: struct{ X int }{X: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Register mock hook
			hook := &mockHook{}
			RegisterHook(hook)
			defer UnregisterHook()

			// Create component with context
			component := newComponentImpl("TestComponent")
			ctx := &Context{component: component}

			// Expose non-ref value - should NOT call OnRefExposed
			ctx.Expose(tt.key, tt.value)

			// Verify hook was NOT called
			assert.Equal(t, int32(0), hook.refExposedCalls.Load(),
				"OnRefExposed should NOT be called for non-Ref values")
		})
	}
}

func TestNotifyHookRefExposed(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Call notifyHookRefExposed directly
	notifyHookRefExposed("comp-123", "ref-0xABC", "myRef")

	// Verify hook was called
	assert.Equal(t, int32(1), hook.refExposedCalls.Load())
	hook.mu.RLock()
	assert.Equal(t, "comp-123", hook.lastRefExposedCompID)
	assert.Equal(t, "ref-0xABC", hook.lastRefExposedRefID)
	assert.Equal(t, "myRef", hook.lastRefExposedName)
	hook.mu.RUnlock()
}

func TestNotifyHookRefExposed_NoHook(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Should not panic when no hook registered
	notifyHookRefExposed("comp-1", "ref-1", "test")
}

func TestNotifyHookRefExposed_ThreadSafe(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Concurrent ref exposed notifications
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			notifyHookRefExposed("comp", "ref", "test")
		}
	}()

	wg.Wait()

	// Verify all calls were made
	assert.Equal(t, int32(iterations), hook.refExposedCalls.Load())
}

// CRITICAL FIX 2: Tests for ExposeComponent calling OnChildAdded hook

func TestContext_ExposeComponent_NotifiesHookChildAdded(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	hook := &mockHook{}
	RegisterHook(hook)

	// Create parent and child components
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")
	ctx := &Context{component: parent}

	// ExposeComponent should call AddChild which notifies hook
	err := ctx.ExposeComponent("myChild", child)
	assert.NoError(t, err)

	// Verify OnChildAdded hook was called
	assert.Equal(t, int32(1), hook.childAddedCalls.Load(),
		"OnChildAdded should be called when component is exposed")

	hook.mu.RLock()
	assert.Equal(t, parent.id, hook.lastParentID,
		"Parent ID should match")
	assert.Equal(t, child.id, hook.lastChildID,
		"Child ID should match")
	hook.mu.RUnlock()
}

func TestContext_ExposeComponent_EstablishesParentChildRelationship(t *testing.T) {
	// Clean up
	defer UnregisterHook()

	// Create parent and child components
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")
	ctx := &Context{component: parent}

	// Initially, parent should have no children
	assert.Len(t, parent.Children(), 0)

	// ExposeComponent should add child to parent
	err := ctx.ExposeComponent("myChild", child)
	assert.NoError(t, err)

	// Verify parent-child relationship
	children := parent.Children()
	assert.Len(t, children, 1, "Parent should have 1 child")
	assert.Equal(t, child.ID(), children[0].ID(), "Child should be in parent's children")
}
