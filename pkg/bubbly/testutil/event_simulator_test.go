package testutil

import (
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestEmit_Basic tests basic event emission functionality.
func TestEmit_Basic(t *testing.T) {
	tests := []struct {
		name    string
		event   string
		payload interface{}
	}{
		{
			name:    "emit with nil payload",
			event:   "click",
			payload: nil,
		},
		{
			name:    "emit with string payload",
			event:   "submit",
			payload: "test data",
		},
		{
			name:    "emit with int payload",
			event:   "count",
			payload: 42,
		},
		{
			name:    "emit with struct payload",
			event:   "data",
			payload: struct{ Value string }{Value: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			harness := NewHarness(t)

			// Create a simple component that tracks events
			eventReceived := false
			var receivedPayload interface{}

			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					ctx.On(tt.event, func(data interface{}) {
						eventReceived = true
						receivedPayload = data
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			assert.NoError(t, err)

			ct := harness.Mount(component)

			// Emit event
			ct.Emit(tt.event, tt.payload)

			// Verify event was received
			assert.True(t, eventReceived, "event should have been received")
			assert.Equal(t, tt.payload, receivedPayload, "payload should match")
		})
	}
}

// TestEmit_EventTracking tests that emitted events are tracked.
func TestEmit_EventTracking(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("test-event", func(data interface{}) {
				// Handler exists
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// Emit event
	ct.Emit("test-event", "payload")

	// Verify event was tracked
	assert.True(t, ct.events.tracker.WasFired("test-event"))
	events := ct.events.tracker.GetEvents("test-event")
	assert.Len(t, events, 1)
	assert.Equal(t, "payload", events[0].Payload)
}

// TestEmit_MultipleEvents tests emitting multiple different events.
func TestEmit_MultipleEvents(t *testing.T) {
	harness := NewHarness(t)

	events := make(map[string]int)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("event1", func(data interface{}) {
				events["event1"]++
			})
			ctx.On("event2", func(data interface{}) {
				events["event2"]++
			})
			ctx.On("event3", func(data interface{}) {
				events["event3"]++
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// Emit multiple events
	ct.Emit("event1", nil)
	ct.Emit("event2", nil)
	ct.Emit("event3", nil)
	ct.Emit("event1", nil) // Emit event1 again

	// Verify all events received
	assert.Equal(t, 2, events["event1"])
	assert.Equal(t, 1, events["event2"])
	assert.Equal(t, 1, events["event3"])
}

// TestEmitAndWait_Success tests successful wait for event processing.
func TestEmitAndWait_Success(t *testing.T) {
	harness := NewHarness(t)

	processed := false

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("process", func(data interface{}) {
				// Simulate some processing time
				time.Sleep(50 * time.Millisecond)
				processed = true
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// Emit and wait
	ct.EmitAndWait("process", nil, 200*time.Millisecond)

	// Should have processed by now
	assert.True(t, processed, "event should have been processed")
}

// TestEmitAndWait_Timeout tests timeout behavior.
func TestEmitAndWait_Timeout(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("slow", func(data interface{}) {
				// Simulate slow processing
				time.Sleep(200 * time.Millisecond)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// This should timeout (50ms timeout, but handler takes 200ms)
	// EmitAndWait should not panic, just return after timeout
	ct.EmitAndWait("slow", nil, 50*time.Millisecond)

	// Test passes if we get here without hanging
}

// TestEmitMultiple_Order tests that multiple events are emitted in order.
func TestEmitMultiple_Order(t *testing.T) {
	harness := NewHarness(t)

	var order []string

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("event1", func(data interface{}) {
				order = append(order, "event1")
			})
			ctx.On("event2", func(data interface{}) {
				order = append(order, "event2")
			})
			ctx.On("event3", func(data interface{}) {
				order = append(order, "event3")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// Emit multiple events
	events := []Event{
		{Name: "event1", Payload: nil},
		{Name: "event2", Payload: nil},
		{Name: "event3", Payload: nil},
	}
	ct.EmitMultiple(events)

	// Verify order
	assert.Equal(t, []string{"event1", "event2", "event3"}, order)
}

// TestEmitMultiple_WithPayloads tests multiple events with different payloads.
func TestEmitMultiple_WithPayloads(t *testing.T) {
	harness := NewHarness(t)

	var payloads []interface{}

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("data", func(data interface{}) {
				payloads = append(payloads, data)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// Emit multiple events with payloads
	events := []Event{
		{Name: "data", Payload: "first"},
		{Name: "data", Payload: 42},
		{Name: "data", Payload: true},
	}
	ct.EmitMultiple(events)

	// Verify payloads
	assert.Len(t, payloads, 3)
	assert.Equal(t, "first", payloads[0])
	assert.Equal(t, 42, payloads[1])
	assert.Equal(t, true, payloads[2])
}

// TestEmitMultiple_Empty tests emitting empty event list.
func TestEmitMultiple_Empty(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// Should not panic with empty list
	ct.EmitMultiple([]Event{})
	ct.EmitMultiple(nil)
}

// TestEmit_StateUpdates tests that state updates after event emission.
func TestEmit_StateUpdates(t *testing.T) {
	harness := NewHarness(t)

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(data interface{}) {
				current := count.Get().(int)
				count.Set(current + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// Initial state
	ct.AssertRefEquals("count", 0)

	// Emit event
	ct.Emit("increment", nil)

	// State should be updated
	ct.AssertRefEquals("count", 1)

	// Emit again
	ct.Emit("increment", nil)
	ct.AssertRefEquals("count", 2)
}

// TestEmit_HandlerExecution tests that event handlers execute correctly.
func TestEmit_HandlerExecution(t *testing.T) {
	harness := NewHarness(t)

	handlerCalled := false
	var receivedData interface{}

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("test", func(data interface{}) {
				handlerCalled = true
				receivedData = data
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	ct := harness.Mount(component)

	// Emit event
	testData := map[string]string{"key": "value"}
	ct.Emit("test", testData)

	// Verify handler executed
	assert.True(t, handlerCalled)
	assert.Equal(t, testData, receivedData)
}
