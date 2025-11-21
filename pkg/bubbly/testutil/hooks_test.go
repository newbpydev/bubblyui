package testutil

import (
	"sync"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTestHooks tests creating a new TestHooks instance
func TestNewTestHooks(t *testing.T) {
	hooks := NewTestHooks()

	require.NotNil(t, hooks)
	assert.Nil(t, hooks.onStateChange)
	assert.Nil(t, hooks.onEvent)
	assert.Nil(t, hooks.onUpdate)
}

// TestTestHooks_SetOnStateChange tests setting state change callback
func TestTestHooks_SetOnStateChange(t *testing.T) {
	hooks := NewTestHooks()

	called := false
	var capturedName string
	var capturedValue interface{}

	hooks.SetOnStateChange(func(name string, value interface{}) {
		called = true
		capturedName = name
		capturedValue = value
	})

	// Trigger the callback
	hooks.TriggerStateChange("count", 42)

	assert.True(t, called, "callback should be called")
	assert.Equal(t, "count", capturedName)
	assert.Equal(t, 42, capturedValue)
}

// TestTestHooks_SetOnEvent tests setting event callback
func TestTestHooks_SetOnEvent(t *testing.T) {
	hooks := NewTestHooks()

	called := false
	var capturedEvent string
	var capturedPayload interface{}

	hooks.SetOnEvent(func(eventName string, payload interface{}) {
		called = true
		capturedEvent = eventName
		capturedPayload = payload
	})

	// Trigger the callback
	hooks.TriggerEvent("click", map[string]int{"x": 10, "y": 20})

	assert.True(t, called, "callback should be called")
	assert.Equal(t, "click", capturedEvent)
	assert.Equal(t, map[string]int{"x": 10, "y": 20}, capturedPayload)
}

// TestTestHooks_SetOnUpdate tests setting update callback
func TestTestHooks_SetOnUpdate(t *testing.T) {
	hooks := NewTestHooks()

	callCount := 0

	hooks.SetOnUpdate(func() {
		callCount++
	})

	// Trigger multiple times
	hooks.TriggerUpdate()
	hooks.TriggerUpdate()
	hooks.TriggerUpdate()

	assert.Equal(t, 3, callCount, "callback should be called 3 times")
}

// TestTestHooks_TriggerWithoutCallback tests triggering without callbacks set
func TestTestHooks_TriggerWithoutCallback(t *testing.T) {
	hooks := NewTestHooks()

	// Should not panic when callbacks are nil
	assert.NotPanics(t, func() {
		hooks.TriggerStateChange("count", 42)
		hooks.TriggerEvent("click", nil)
		hooks.TriggerUpdate()
	})
}

// TestTestHooks_Clear tests clearing all callbacks
func TestTestHooks_Clear(t *testing.T) {
	hooks := NewTestHooks()

	// Set all callbacks
	stateChangeCalled := false
	eventCalled := false
	updateCalled := false

	hooks.SetOnStateChange(func(string, interface{}) {
		stateChangeCalled = true
	})
	hooks.SetOnEvent(func(string, interface{}) {
		eventCalled = true
	})
	hooks.SetOnUpdate(func() {
		updateCalled = true
	})

	// Clear all callbacks
	hooks.Clear()

	// Trigger should not call anything
	hooks.TriggerStateChange("count", 42)
	hooks.TriggerEvent("click", nil)
	hooks.TriggerUpdate()

	assert.False(t, stateChangeCalled, "state change callback should not be called after clear")
	assert.False(t, eventCalled, "event callback should not be called after clear")
	assert.False(t, updateCalled, "update callback should not be called after clear")
}

// TestTestHooks_ThreadSafety tests concurrent access to hooks
func TestTestHooks_ThreadSafety(t *testing.T) {
	hooks := NewTestHooks()

	var wg sync.WaitGroup
	callCount := 0
	var mu sync.Mutex

	// Set callback that increments counter
	hooks.SetOnUpdate(func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	// Trigger from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				hooks.TriggerUpdate()
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1000, callCount, "all triggers should be counted")
}

// TestTestHooks_MultipleCallbacks tests setting callbacks multiple times
func TestTestHooks_MultipleCallbacks(t *testing.T) {
	hooks := NewTestHooks()

	// Set first callback
	firstCalled := false
	hooks.SetOnUpdate(func() {
		firstCalled = true
	})

	// Set second callback (should replace first)
	secondCalled := false
	hooks.SetOnUpdate(func() {
		secondCalled = true
	})

	hooks.TriggerUpdate()

	assert.False(t, firstCalled, "first callback should be replaced")
	assert.True(t, secondCalled, "second callback should be called")
}

// TestTestHarness_InstallHooks tests installing hooks into component
func TestTestHarness_InstallHooks(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	// Install hooks
	harness.installHooks(component)

	// Verify hooks are created
	assert.NotNil(t, harness.hooks, "hooks should be created")
	assert.Equal(t, component, harness.component, "component should be stored")
}

// TestTestHarness_InstallHooksIdempotent tests installing hooks multiple times
func TestTestHarness_InstallHooksIdempotent(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	// Install hooks multiple times
	harness.installHooks(component)
	firstHooks := harness.hooks

	harness.installHooks(component)
	secondHooks := harness.hooks

	// Should reuse same hooks instance
	assert.Equal(t, firstHooks, secondHooks, "should reuse hooks instance")
}

// TestTestHarness_RemoveHooks tests removing hooks
func TestTestHarness_RemoveHooks(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	// Install hooks
	harness.installHooks(component)
	require.NotNil(t, harness.hooks)

	// Set a callback to verify it gets cleared
	harness.hooks.SetOnUpdate(func() {
		// This callback should be cleared when hooks are removed
	})

	// Remove hooks
	harness.removeHooks()

	// Verify hooks are cleared
	assert.Nil(t, harness.hooks, "hooks should be nil after removal")
}

// TestTestHarness_RemoveHooksIdempotent tests removing hooks multiple times
func TestTestHarness_RemoveHooksIdempotent(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	// Install hooks
	harness.installHooks(component)

	// Remove hooks multiple times (should not panic)
	assert.NotPanics(t, func() {
		harness.removeHooks()
		harness.removeHooks()
		harness.removeHooks()
	})

	assert.Nil(t, harness.hooks)
}

// TestTestHarness_HooksIntegration tests full hooks workflow
func TestTestHarness_HooksIntegration(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	// Install hooks
	harness.installHooks(component)
	require.NotNil(t, harness.hooks)

	// Track state changes
	stateChanges := []string{}
	harness.hooks.SetOnStateChange(func(name string, value interface{}) {
		stateChanges = append(stateChanges, name)
	})

	// Track events
	eventsFired := []string{}
	harness.hooks.SetOnEvent(func(eventName string, payload interface{}) {
		eventsFired = append(eventsFired, eventName)
	})

	// Track updates
	updateCount := 0
	harness.hooks.SetOnUpdate(func() {
		updateCount++
	})

	// Simulate state change
	harness.hooks.TriggerStateChange("count", 1)
	harness.hooks.TriggerStateChange("count", 2)

	// Simulate events
	harness.hooks.TriggerEvent("increment", nil)
	harness.hooks.TriggerEvent("decrement", nil)

	// Simulate updates
	harness.hooks.TriggerUpdate()
	harness.hooks.TriggerUpdate()
	harness.hooks.TriggerUpdate()

	// Verify tracking
	assert.Equal(t, []string{"count", "count"}, stateChanges)
	assert.Equal(t, []string{"increment", "decrement"}, eventsFired)
	assert.Equal(t, 3, updateCount)

	// Remove hooks
	harness.removeHooks()
	assert.Nil(t, harness.hooks)
}

// TestTestHooks_StateChangeTracking tests tracking multiple state changes
func TestTestHooks_StateChangeTracking(t *testing.T) {
	tests := []struct {
		name    string
		changes []struct {
			name  string
			value interface{}
		}
		expected []string
	}{
		{
			name: "single state change",
			changes: []struct {
				name  string
				value interface{}
			}{
				{"count", 42},
			},
			expected: []string{"count"},
		},
		{
			name: "multiple different states",
			changes: []struct {
				name  string
				value interface{}
			}{
				{"count", 1},
				{"name", "test"},
				{"enabled", true},
			},
			expected: []string{"count", "name", "enabled"},
		},
		{
			name: "same state multiple times",
			changes: []struct {
				name  string
				value interface{}
			}{
				{"count", 1},
				{"count", 2},
				{"count", 3},
			},
			expected: []string{"count", "count", "count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := NewTestHooks()

			stateChanges := []string{}
			hooks.SetOnStateChange(func(name string, value interface{}) {
				stateChanges = append(stateChanges, name)
			})

			for _, change := range tt.changes {
				hooks.TriggerStateChange(change.name, change.value)
			}

			assert.Equal(t, tt.expected, stateChanges)
		})
	}
}

// TestTestHooks_EventTracking tests tracking multiple events
func TestTestHooks_EventTracking(t *testing.T) {
	tests := []struct {
		name   string
		events []struct {
			name    string
			payload interface{}
		}
		expected []string
	}{
		{
			name: "single event",
			events: []struct {
				name    string
				payload interface{}
			}{
				{"click", nil},
			},
			expected: []string{"click"},
		},
		{
			name: "multiple different events",
			events: []struct {
				name    string
				payload interface{}
			}{
				{"click", nil},
				{"submit", map[string]string{"form": "test"}},
				{"cancel", nil},
			},
			expected: []string{"click", "submit", "cancel"},
		},
		{
			name: "same event multiple times",
			events: []struct {
				name    string
				payload interface{}
			}{
				{"click", nil},
				{"click", nil},
				{"click", nil},
			},
			expected: []string{"click", "click", "click"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := NewTestHooks()

			eventsFired := []string{}
			hooks.SetOnEvent(func(eventName string, payload interface{}) {
				eventsFired = append(eventsFired, eventName)
			})

			for _, event := range tt.events {
				hooks.TriggerEvent(event.name, event.payload)
			}

			assert.Equal(t, tt.expected, eventsFired)
		})
	}
}

// TestTestHooks_UpdateTracking tests tracking update calls
func TestTestHooks_UpdateTracking(t *testing.T) {
	tests := []struct {
		name          string
		triggerCount  int
		expectedCount int
	}{
		{
			name:          "no updates",
			triggerCount:  0,
			expectedCount: 0,
		},
		{
			name:          "single update",
			triggerCount:  1,
			expectedCount: 1,
		},
		{
			name:          "multiple updates",
			triggerCount:  10,
			expectedCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := NewTestHooks()

			updateCount := 0
			hooks.SetOnUpdate(func() {
				updateCount++
			})

			for i := 0; i < tt.triggerCount; i++ {
				hooks.TriggerUpdate()
			}

			assert.Equal(t, tt.expectedCount, updateCount)
		})
	}
}
