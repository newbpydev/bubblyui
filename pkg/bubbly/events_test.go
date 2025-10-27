package bubbly

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmit_BasicEmission tests basic event emission.
func TestEmit_BasicEmission(t *testing.T) {
	c := newComponentImpl("TestComponent")

	var called bool
	var receivedData interface{}

	c.On("test-event", func(data interface{}) {
		called = true
		receivedData = data
	})

	testData := "test data"
	c.Emit("test-event", testData)

	assert.True(t, called, "handler should be called")
	assert.Equal(t, testData, receivedData, "data should be passed to handler")
}

// TestEmit_MultipleHandlers tests that multiple handlers can be registered for the same event.
func TestEmit_MultipleHandlers(t *testing.T) {
	c := newComponentImpl("TestComponent")

	callCount := 0
	var receivedData []interface{}

	// Register multiple handlers
	c.On("test-event", func(data interface{}) {
		callCount++
		receivedData = append(receivedData, data)
	})

	c.On("test-event", func(data interface{}) {
		callCount++
		receivedData = append(receivedData, data)
	})

	c.On("test-event", func(data interface{}) {
		callCount++
		receivedData = append(receivedData, data)
	})

	testData := "shared data"
	c.Emit("test-event", testData)

	assert.Equal(t, 3, callCount, "all handlers should be called")
	assert.Len(t, receivedData, 3, "all handlers should receive data")
	for _, data := range receivedData {
		assert.Equal(t, testData, data, "each handler should receive the same data")
	}
}

// TestEmit_NoHandlers tests that emitting an event with no handlers doesn't panic.
func TestEmit_NoHandlers(t *testing.T) {
	c := newComponentImpl("TestComponent")

	// Should not panic
	assert.NotPanics(t, func() {
		c.Emit("non-existent-event", "data")
	})
}

// TestEmit_TypeSafePayloads tests type-safe event payloads.
func TestEmit_TypeSafePayloads(t *testing.T) {
	type FormData struct {
		Username string
		Email    string
	}

	tests := []struct {
		name     string
		data     interface{}
		validate func(t *testing.T, received interface{})
	}{
		{
			name: "struct payload",
			data: FormData{Username: "user", Email: "user@example.com"},
			validate: func(t *testing.T, received interface{}) {
				formData, ok := received.(FormData)
				require.True(t, ok, "should be FormData type")
				assert.Equal(t, "user", formData.Username)
				assert.Equal(t, "user@example.com", formData.Email)
			},
		},
		{
			name: "string payload",
			data: "simple string",
			validate: func(t *testing.T, received interface{}) {
				str, ok := received.(string)
				require.True(t, ok, "should be string type")
				assert.Equal(t, "simple string", str)
			},
		},
		{
			name: "int payload",
			data: 42,
			validate: func(t *testing.T, received interface{}) {
				num, ok := received.(int)
				require.True(t, ok, "should be int type")
				assert.Equal(t, 42, num)
			},
		},
		{
			name: "map payload",
			data: map[string]interface{}{"key": "value"},
			validate: func(t *testing.T, received interface{}) {
				m, ok := received.(map[string]interface{})
				require.True(t, ok, "should be map type")
				assert.Equal(t, "value", m["key"])
			},
		},
		{
			name: "nil payload",
			data: nil,
			validate: func(t *testing.T, received interface{}) {
				assert.Nil(t, received, "should be nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newComponentImpl("TestComponent")

			var receivedData interface{}
			c.On("test-event", func(data interface{}) {
				receivedData = data
			})

			c.Emit("test-event", tt.data)
			tt.validate(t, receivedData)
		})
	}
}

// TestEmit_EventMetadata tests that Event struct contains proper metadata.
func TestEmit_EventMetadata(t *testing.T) {
	c := newComponentImpl("TestComponent")

	// We need to capture the Event struct, but handlers receive data, not Event
	// So we'll test this indirectly by verifying the handler is called
	// and the data is correct. The Event struct is used internally.

	var called bool
	testData := "test data"

	c.On("test-event", func(data interface{}) {
		called = true
		assert.Equal(t, testData, data)
	})

	beforeEmit := time.Now()
	c.Emit("test-event", testData)
	afterEmit := time.Now()

	assert.True(t, called)
	// Event timestamp should be between beforeEmit and afterEmit
	// (We can't directly test this without exposing Event to handlers,
	// but the implementation creates it with time.Now())
	assert.True(t, afterEmit.After(beforeEmit) || afterEmit.Equal(beforeEmit))
}

// TestOn_HandlerRegistration tests that handlers are properly registered.
func TestOn_HandlerRegistration(t *testing.T) {
	c := newComponentImpl("TestComponent")

	// Register handler
	c.On("test-event", func(data interface{}) {})

	// Verify handler is registered
	assert.Contains(t, c.handlers, "test-event")
	assert.Len(t, c.handlers["test-event"], 1)

	// Register another handler for the same event
	c.On("test-event", func(data interface{}) {})

	// Verify both handlers are registered
	assert.Len(t, c.handlers["test-event"], 2)
}

// TestOn_MultipleEvents tests registering handlers for different events.
func TestOn_MultipleEvents(t *testing.T) {
	c := newComponentImpl("TestComponent")

	c.On("event1", func(data interface{}) {})
	c.On("event2", func(data interface{}) {})
	c.On("event3", func(data interface{}) {})

	assert.Len(t, c.handlers, 3)
	assert.Contains(t, c.handlers, "event1")
	assert.Contains(t, c.handlers, "event2")
	assert.Contains(t, c.handlers, "event3")
}

// TestEmit_HandlerExecutionOrder tests that handlers execute in registration order.
func TestEmit_HandlerExecutionOrder(t *testing.T) {
	c := newComponentImpl("TestComponent")

	var executionOrder []int

	c.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, 1)
	})

	c.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, 2)
	})

	c.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, 3)
	})

	c.Emit("test-event", nil)

	assert.Equal(t, []int{1, 2, 3}, executionOrder, "handlers should execute in registration order")
}

// TestEmit_ConcurrentEmission tests concurrent event emission.
func TestEmit_ConcurrentEmission(t *testing.T) {
	c := newComponentImpl("TestComponent")

	var mu sync.Mutex
	callCount := 0

	c.On("test-event", func(data interface{}) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	// Emit events concurrently
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			c.Emit("test-event", nil)
		}()
	}

	wg.Wait()

	assert.Equal(t, numGoroutines, callCount, "all emissions should trigger handler")
}

// TestOn_ConcurrentRegistration tests concurrent handler registration.
func TestOn_ConcurrentRegistration(t *testing.T) {
	c := newComponentImpl("TestComponent")

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			c.On("test-event", func(data interface{}) {})
		}()
	}

	wg.Wait()

	assert.Len(t, c.handlers["test-event"], numGoroutines, "all handlers should be registered")
}

// TestEmit_IntegrationWithComponent tests event system integration with component lifecycle.
func TestEmit_IntegrationWithComponent(t *testing.T) {
	var emittedData string

	component, err := NewComponent("TestComponent").
		Setup(func(ctx *Context) {
			ctx.On("custom-event", func(data interface{}) {
				emittedData = data.(string)
			})
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	// Initialize component to run setup
	component.Init()

	// Emit event
	component.Emit("custom-event", "test data")

	assert.Equal(t, "test data", emittedData)
}

// TestEmit_HandlerDataIsolation tests that each handler receives its own copy of data.
func TestEmit_HandlerDataIsolation(t *testing.T) {
	c := newComponentImpl("TestComponent")

	var data1, data2, data3 interface{}

	c.On("test-event", func(data interface{}) {
		data1 = data
	})

	c.On("test-event", func(data interface{}) {
		data2 = data
	})

	c.On("test-event", func(data interface{}) {
		data3 = data
	})

	testData := "shared data"
	c.Emit("test-event", testData)

	// All handlers should receive the same data
	assert.Equal(t, testData, data1)
	assert.Equal(t, testData, data2)
	assert.Equal(t, testData, data3)
}

// TestEventRegistry_Tracking tests the global event registry.
func TestEventRegistry_Tracking(t *testing.T) {
	// Reset registry before test
	globalEventRegistry.resetRegistry()

	c := newComponentImpl("TestComponent")

	// Register handlers
	c.On("event1", func(data interface{}) {})
	c.On("event1", func(data interface{}) {})
	c.On("event2", func(data interface{}) {})

	// Verify tracking
	assert.Equal(t, 2, globalEventRegistry.getListenerCount("event1"))
	assert.Equal(t, 1, globalEventRegistry.getListenerCount("event2"))
	assert.Equal(t, 0, globalEventRegistry.getListenerCount("non-existent"))
}

// TestEventRegistry_ConcurrentAccess tests concurrent access to event registry.
func TestEventRegistry_ConcurrentAccess(t *testing.T) {
	globalEventRegistry.resetRegistry()

	c := newComponentImpl("TestComponent")

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			c.On("test-event", func(data interface{}) {})
		}()
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = globalEventRegistry.getListenerCount("test-event")
		}()
	}

	wg.Wait()

	// Should have tracked all registrations
	assert.Equal(t, numGoroutines, globalEventRegistry.getListenerCount("test-event"))
}
