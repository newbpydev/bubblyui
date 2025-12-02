package profiler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestHookAdapter_OnComponentMount(t *testing.T) {
	prof := New(WithEnabled(true))
	adapter := NewHookAdapter(prof)

	adapter.OnComponentMount("comp-1", "TestComponent")

	// Verify component name is stored
	assert.Equal(t, "TestComponent", adapter.componentNames["comp-1"])
}

func TestHookAdapter_OnRenderComplete(t *testing.T) {
	prof := New(WithEnabled(true))
	adapter := NewHookAdapter(prof)

	// Mount component first
	adapter.OnComponentMount("comp-1", "TestComponent")

	// Record render
	adapter.OnRenderComplete("comp-1", 10*time.Millisecond)

	// Verify metrics were recorded
	tracker := adapter.GetComponentTracker()
	metrics := tracker.GetMetrics("comp-1")

	require.NotNil(t, metrics)
	assert.Equal(t, "comp-1", metrics.ComponentID)
	assert.Equal(t, "TestComponent", metrics.ComponentName)
	assert.Equal(t, int64(1), metrics.RenderCount)
	assert.Equal(t, 10*time.Millisecond, metrics.TotalRenderTime)
}

func TestHookAdapter_OnRenderComplete_UnknownComponent(t *testing.T) {
	prof := New(WithEnabled(true))
	adapter := NewHookAdapter(prof)

	// Record render without mounting first
	adapter.OnRenderComplete("comp-unknown", 5*time.Millisecond)

	// Should still record with "Unknown" name
	tracker := adapter.GetComponentTracker()
	metrics := tracker.GetMetrics("comp-unknown")

	require.NotNil(t, metrics)
	assert.Equal(t, "Unknown", metrics.ComponentName)
}

func TestCompositeHook_ForwardsToAllHooks(t *testing.T) {
	// Create mock hooks that track calls
	hook1Calls := 0
	hook2Calls := 0

	hook1 := &mockHook{onRenderComplete: func(id string, d time.Duration) {
		hook1Calls++
	}}
	hook2 := &mockHook{onRenderComplete: func(id string, d time.Duration) {
		hook2Calls++
	}}

	// Create composite
	composite := NewCompositeHook(hook1, hook2)

	// Trigger event
	composite.OnRenderComplete("comp-1", 10*time.Millisecond)

	// Both hooks should be called
	assert.Equal(t, 1, hook1Calls)
	assert.Equal(t, 1, hook2Calls)
}

func TestCompositeHook_HandlesNilHooks(t *testing.T) {
	composite := NewCompositeHook(nil, nil)

	// Should not panic
	assert.NotPanics(t, func() {
		composite.OnRenderComplete("comp-1", 10*time.Millisecond)
		composite.OnComponentMount("comp-1", "Test")
	})
}

func TestProfiler_SetHookAdapter(t *testing.T) {
	prof := New(WithEnabled(true))
	adapter := NewHookAdapter(prof)

	prof.SetHookAdapter(adapter)

	// Hook adapter should be set
	assert.NotNil(t, prof.hookAdapter)
}

func TestProfiler_GenerateReport_WithHookAdapter(t *testing.T) {
	prof := New(WithEnabled(true))
	adapter := NewHookAdapter(prof)
	prof.SetHookAdapter(adapter)

	// Simulate component renders
	adapter.OnComponentMount("comp-1", "Counter")
	adapter.OnComponentMount("comp-2", "Button")
	adapter.OnRenderComplete("comp-1", 10*time.Millisecond)
	adapter.OnRenderComplete("comp-2", 5*time.Millisecond)
	adapter.OnRenderComplete("comp-1", 15*time.Millisecond)

	// Generate report
	report := prof.GenerateReport()

	require.NotNil(t, report)
	require.NotNil(t, report.Components)
	assert.Len(t, report.Components, 2)

	// Find Counter metrics
	var counterMetrics *ComponentMetrics
	for _, m := range report.Components {
		if m.ComponentName == "Counter" {
			counterMetrics = m
			break
		}
	}

	require.NotNil(t, counterMetrics)
	assert.Equal(t, int64(2), counterMetrics.RenderCount)
	assert.Equal(t, 25*time.Millisecond, counterMetrics.TotalRenderTime)
}

// mockHook is a test implementation of FrameworkHook
type mockHook struct {
	onComponentMount   func(id, name string)
	onRenderComplete   func(id string, duration time.Duration)
	onComponentUnmount func(id string)
}

func (m *mockHook) OnComponentMount(id, name string) {
	if m.onComponentMount != nil {
		m.onComponentMount(id, name)
	}
}

func (m *mockHook) OnComponentUpdate(id string, msg interface{}) {}

func (m *mockHook) OnComponentUnmount(id string) {
	if m.onComponentUnmount != nil {
		m.onComponentUnmount(id)
	}
}

func (m *mockHook) OnRefChange(id string, oldValue, newValue interface{}) {}

func (m *mockHook) OnEvent(componentID, eventName string, data interface{}) {}

func (m *mockHook) OnRenderComplete(componentID string, duration time.Duration) {
	if m.onRenderComplete != nil {
		m.onRenderComplete(componentID, duration)
	}
}

func (m *mockHook) OnComputedChange(id string, oldValue, newValue interface{}) {}

func (m *mockHook) OnWatchCallback(watcherID string, newValue, oldValue interface{}) {}

func (m *mockHook) OnEffectRun(effectID string) {}

func (m *mockHook) OnChildAdded(parentID, childID string) {}

func (m *mockHook) OnChildRemoved(parentID, childID string) {}

func (m *mockHook) OnRefExposed(componentID, refID, refName string) {}

// Ensure mockHook implements bubbly.FrameworkHook
var _ bubbly.FrameworkHook = (*mockHook)(nil)
