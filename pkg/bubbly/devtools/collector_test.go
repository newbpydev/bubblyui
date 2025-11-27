package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestNewDataCollector_CreatesInstance tests that NewDataCollector creates a valid instance
func TestNewDataCollector_CreatesInstance(t *testing.T) {
	collector := NewDataCollector()

	assert.NotNil(t, collector, "NewDataCollector should return non-nil instance")
}

// TestAddComponentHook_AddsHook tests adding a component hook
func TestAddComponentHook_AddsHook(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockComponentHook{}

	collector.AddComponentHook(hook)

	// Verify hook was added by firing an event
	snapshot := &ComponentSnapshot{ID: "test-1", Name: "TestComponent"}
	collector.FireComponentCreated(snapshot)

	assert.True(t, hook.createdCalled, "Hook should be called when event fires")
	assert.Equal(t, snapshot, hook.createdSnapshot, "Hook should receive correct snapshot")
}

// TestRemoveComponentHook_RemovesHook tests removing a component hook
func TestRemoveComponentHook_RemovesHook(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockComponentHook{}

	collector.AddComponentHook(hook)
	collector.RemoveComponentHook(hook)

	// Verify hook was removed - should not be called
	snapshot := &ComponentSnapshot{ID: "test-1", Name: "TestComponent"}
	collector.FireComponentCreated(snapshot)

	assert.False(t, hook.createdCalled, "Hook should not be called after removal")
}

// TestFireComponentCreated_CallsAllHooks tests that all registered hooks are called
func TestFireComponentCreated_CallsAllHooks(t *testing.T) {
	collector := NewDataCollector()
	hook1 := &mockComponentHook{}
	hook2 := &mockComponentHook{}
	hook3 := &mockComponentHook{}

	collector.AddComponentHook(hook1)
	collector.AddComponentHook(hook2)
	collector.AddComponentHook(hook3)

	snapshot := &ComponentSnapshot{ID: "test-1", Name: "TestComponent"}
	collector.FireComponentCreated(snapshot)

	assert.True(t, hook1.createdCalled, "Hook 1 should be called")
	assert.True(t, hook2.createdCalled, "Hook 2 should be called")
	assert.True(t, hook3.createdCalled, "Hook 3 should be called")
}

// TestFireComponentMounted_CallsHooks tests FireComponentMounted
func TestFireComponentMounted_CallsHooks(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockComponentHook{}

	collector.AddComponentHook(hook)
	collector.FireComponentMounted("component-123")

	assert.True(t, hook.mountedCalled, "OnComponentMounted should be called")
	assert.Equal(t, "component-123", hook.mountedID, "Should receive correct ID")
}

// TestFireComponentUpdated_CallsHooks tests FireComponentUpdated
func TestFireComponentUpdated_CallsHooks(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockComponentHook{}

	collector.AddComponentHook(hook)
	collector.FireComponentUpdated("component-456")

	assert.True(t, hook.updatedCalled, "OnComponentUpdated should be called")
	assert.Equal(t, "component-456", hook.updatedID, "Should receive correct ID")
}

// TestFireComponentUnmounted_CallsHooks tests FireComponentUnmounted
func TestFireComponentUnmounted_CallsHooks(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockComponentHook{}

	collector.AddComponentHook(hook)
	collector.FireComponentUnmounted("component-789")

	assert.True(t, hook.unmountedCalled, "OnComponentUnmounted should be called")
	assert.Equal(t, "component-789", hook.unmountedID, "Should receive correct ID")
}

// TestStateHooks_WorkCorrectly tests state hook functionality
func TestStateHooks_WorkCorrectly(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockStateHook{}

	collector.AddStateHook(hook)
	collector.FireRefChanged("ref-1", "old", "new")

	assert.True(t, hook.refChangedCalled, "OnRefChanged should be called")
	assert.Equal(t, "ref-1", hook.refID, "Should receive correct ref ID")
	assert.Equal(t, "old", hook.oldValue, "Should receive correct old value")
	assert.Equal(t, "new", hook.newValue, "Should receive correct new value")
}

// TestEventHooks_WorkCorrectly tests event hook functionality
func TestEventHooks_WorkCorrectly(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockEventHook{}

	collector.AddEventHook(hook)
	event := &EventRecord{ID: "evt-1", Name: "click"}
	collector.FireEvent(event)

	assert.True(t, hook.eventCalled, "OnEvent should be called")
	assert.Equal(t, event, hook.event, "Should receive correct event")
}

// TestPerformanceHooks_WorkCorrectly tests performance hook functionality
func TestPerformanceHooks_WorkCorrectly(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockPerformanceHook{}

	collector.AddPerformanceHook(hook)
	collector.FireRenderComplete("comp-1", 5*time.Millisecond)

	assert.True(t, hook.renderCalled, "OnRenderComplete should be called")
	assert.Equal(t, "comp-1", hook.componentID, "Should receive correct component ID")
	assert.Equal(t, 5*time.Millisecond, hook.duration, "Should receive correct duration")
}

// TestHookPanic_DoesNotCrashCollector tests that panicking hooks don't crash the collector
func TestHookPanic_DoesNotCrashCollector(t *testing.T) {
	collector := NewDataCollector()
	panicHook := &panicComponentHook{}
	normalHook := &mockComponentHook{}

	collector.AddComponentHook(panicHook)
	collector.AddComponentHook(normalHook)

	// Should not panic
	assert.NotPanics(t, func() {
		snapshot := &ComponentSnapshot{ID: "test-1", Name: "TestComponent"}
		collector.FireComponentCreated(snapshot)
	}, "Collector should not panic when hook panics")

	// Normal hook should still be called
	assert.True(t, normalHook.createdCalled, "Other hooks should still be called after panic")
}

// TestCollector_ConcurrentAccess_ThreadSafe tests thread-safe concurrent access
func TestCollector_ConcurrentAccess_ThreadSafe(t *testing.T) {
	collector := NewDataCollector()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent hook additions
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hook := &mockComponentHook{}
			collector.AddComponentHook(hook)
		}()
	}

	// Concurrent hook firings
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			snapshot := &ComponentSnapshot{ID: "test", Name: "Test"}
			collector.FireComponentCreated(snapshot)
		}()
	}

	// Concurrent hook removals
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hook := &mockComponentHook{}
			collector.RemoveComponentHook(hook)
		}()
	}

	wg.Wait()

	// Should not panic or deadlock
	assert.NotNil(t, collector, "Collector should still be valid after concurrent access")
}

// Mock implementations for testing

type mockComponentHook struct {
	mu              sync.Mutex
	createdCalled   bool
	createdSnapshot *ComponentSnapshot
	mountedCalled   bool
	mountedID       string
	updatedCalled   bool
	updatedID       string
	unmountedCalled bool
	unmountedID     string
}

func (m *mockComponentHook) OnComponentCreated(snapshot *ComponentSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createdCalled = true
	m.createdSnapshot = snapshot
}

func (m *mockComponentHook) OnComponentMounted(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mountedCalled = true
	m.mountedID = id
}

func (m *mockComponentHook) OnComponentUpdated(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updatedCalled = true
	m.updatedID = id
}

func (m *mockComponentHook) OnComponentUnmounted(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.unmountedCalled = true
	m.unmountedID = id
}

type panicComponentHook struct{}

func (p *panicComponentHook) OnComponentCreated(snapshot *ComponentSnapshot) {
	panic("test panic in hook")
}

func (p *panicComponentHook) OnComponentMounted(id string)   {}
func (p *panicComponentHook) OnComponentUpdated(id string)   {}
func (p *panicComponentHook) OnComponentUnmounted(id string) {}

type mockStateHook struct {
	refChangedCalled bool
	refID            string
	oldValue         interface{}
	newValue         interface{}
}

func (m *mockStateHook) OnRefChanged(refID string, oldValue, newValue interface{}) {
	m.refChangedCalled = true
	m.refID = refID
	m.oldValue = oldValue
	m.newValue = newValue
}

func (m *mockStateHook) OnComputedEvaluated(computedID string, value interface{}, duration time.Duration) {
}
func (m *mockStateHook) OnWatcherTriggered(watcherID string, value interface{}) {}

type mockEventHook struct {
	eventCalled bool
	event       *EventRecord
}

func (m *mockEventHook) OnEvent(event *EventRecord) {
	m.eventCalled = true
	m.event = event
}

type mockPerformanceHook struct {
	renderCalled bool
	componentID  string
	duration     time.Duration
}

func (m *mockPerformanceHook) OnRenderComplete(componentID string, duration time.Duration) {
	m.renderCalled = true
	m.componentID = componentID
	m.duration = duration
}

// TestRemoveStateHook_RemovesHook tests removing a state hook
func TestRemoveStateHook_RemovesHook(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockStateHook{}

	collector.AddStateHook(hook)
	collector.RemoveStateHook(hook)

	// Verify hook was removed - should not be called
	collector.FireRefChanged("ref-1", "old", "new")

	assert.False(t, hook.refChangedCalled, "Hook should not be called after removal")
}

// TestRemoveStateHook_RemovesCorrectHook tests removing specific hook from multiple
func TestRemoveStateHook_RemovesCorrectHook(t *testing.T) {
	collector := NewDataCollector()
	hook1 := &mockStateHook{}
	hook2 := &mockStateHook{}
	hook3 := &mockStateHook{}

	collector.AddStateHook(hook1)
	collector.AddStateHook(hook2)
	collector.AddStateHook(hook3)

	// Remove middle hook
	collector.RemoveStateHook(hook2)

	// Fire event
	collector.FireRefChanged("ref-1", "old", "new")

	assert.True(t, hook1.refChangedCalled, "Hook 1 should still be called")
	assert.False(t, hook2.refChangedCalled, "Hook 2 should not be called after removal")
	assert.True(t, hook3.refChangedCalled, "Hook 3 should still be called")
}

// TestRemoveStateHook_NonExistentHook tests removing a hook that was never added
func TestRemoveStateHook_NonExistentHook(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockStateHook{}

	// Should not panic when removing non-existent hook
	assert.NotPanics(t, func() {
		collector.RemoveStateHook(hook)
	}, "Removing non-existent hook should not panic")
}

// TestRemoveEventHook_RemovesHook tests removing an event hook
func TestRemoveEventHook_RemovesHook(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockEventHook{}

	collector.AddEventHook(hook)
	collector.RemoveEventHook(hook)

	// Verify hook was removed - should not be called
	event := &EventRecord{ID: "evt-1", Name: "click"}
	collector.FireEvent(event)

	assert.False(t, hook.eventCalled, "Hook should not be called after removal")
}

// TestRemoveEventHook_RemovesCorrectHook tests removing specific hook from multiple
func TestRemoveEventHook_RemovesCorrectHook(t *testing.T) {
	collector := NewDataCollector()
	hook1 := &mockEventHook{}
	hook2 := &mockEventHook{}
	hook3 := &mockEventHook{}

	collector.AddEventHook(hook1)
	collector.AddEventHook(hook2)
	collector.AddEventHook(hook3)

	// Remove middle hook
	collector.RemoveEventHook(hook2)

	// Fire event
	event := &EventRecord{ID: "evt-1", Name: "click"}
	collector.FireEvent(event)

	assert.True(t, hook1.eventCalled, "Hook 1 should still be called")
	assert.False(t, hook2.eventCalled, "Hook 2 should not be called after removal")
	assert.True(t, hook3.eventCalled, "Hook 3 should still be called")
}

// TestRemoveEventHook_NonExistentHook tests removing a hook that was never added
func TestRemoveEventHook_NonExistentHook(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockEventHook{}

	// Should not panic when removing non-existent hook
	assert.NotPanics(t, func() {
		collector.RemoveEventHook(hook)
	}, "Removing non-existent hook should not panic")
}

// TestRemovePerformanceHook_RemovesHook tests removing a performance hook
func TestRemovePerformanceHook_RemovesHook(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockPerformanceHook{}

	collector.AddPerformanceHook(hook)
	collector.RemovePerformanceHook(hook)

	// Verify hook was removed - should not be called
	collector.FireRenderComplete("comp-1", 5*time.Millisecond)

	assert.False(t, hook.renderCalled, "Hook should not be called after removal")
}

// TestRemovePerformanceHook_RemovesCorrectHook tests removing specific hook from multiple
func TestRemovePerformanceHook_RemovesCorrectHook(t *testing.T) {
	collector := NewDataCollector()
	hook1 := &mockPerformanceHook{}
	hook2 := &mockPerformanceHook{}
	hook3 := &mockPerformanceHook{}

	collector.AddPerformanceHook(hook1)
	collector.AddPerformanceHook(hook2)
	collector.AddPerformanceHook(hook3)

	// Remove middle hook
	collector.RemovePerformanceHook(hook2)

	// Fire event
	collector.FireRenderComplete("comp-1", 5*time.Millisecond)

	assert.True(t, hook1.renderCalled, "Hook 1 should still be called")
	assert.False(t, hook2.renderCalled, "Hook 2 should not be called after removal")
	assert.True(t, hook3.renderCalled, "Hook 3 should still be called")
}

// TestRemovePerformanceHook_NonExistentHook tests removing a hook that was never added
func TestRemovePerformanceHook_NonExistentHook(t *testing.T) {
	collector := NewDataCollector()
	hook := &mockPerformanceHook{}

	// Should not panic when removing non-existent hook
	assert.NotPanics(t, func() {
		collector.RemovePerformanceHook(hook)
	}, "Removing non-existent hook should not panic")
}

// TestHookPanic_WithErrorReporter tests panic recovery with error reporter configured
func TestHookPanic_WithErrorReporter(t *testing.T) {
	collector := NewDataCollector()

	// Create a mock reporter to verify panic is reported
	mockReporter := &mockPanicReporter{}
	observability.SetErrorReporter(mockReporter)
	defer observability.SetErrorReporter(nil)

	// Add a panicking hook
	panicHook := &panicComponentHook{}
	collector.AddComponentHook(panicHook)

	// Should not panic and reporter should receive the panic
	assert.NotPanics(t, func() {
		snapshot := &ComponentSnapshot{ID: "test-1", Name: "TestComponent"}
		collector.FireComponentCreated(snapshot)
	}, "Collector should recover from panic")

	// Reporter should have received the panic
	assert.True(t, mockReporter.reportedPanic, "Error reporter should receive panic report")
}

// mockPanicReporter implements observability.ErrorReporter for testing panic reporting
type mockPanicReporter struct {
	mu            sync.Mutex
	reportedPanic bool
	panicValue    interface{}
}

func (m *mockPanicReporter) ReportError(err error, ctx *observability.ErrorContext) {
	// Not used in this test
}

func (m *mockPanicReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reportedPanic = true
	m.panicValue = err.PanicValue
}

func (m *mockPanicReporter) Flush(timeout time.Duration) error {
	return nil
}
