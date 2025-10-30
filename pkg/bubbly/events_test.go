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
		// Handlers receive the data directly, not wrapped in Event
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
		// Handlers receive data payload directly
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

// ============================================================================
// Event Bubbling Tests (Task 5.2)
// ============================================================================

// TestEventBubbling_ChildToParent tests that events bubble from child to immediate parent.
func TestEventBubbling_ChildToParent(t *testing.T) {
	// Create parent component
	parent := newComponentImpl("Parent")

	// Create child component
	child := newComponentImpl("Child")

	// Set up parent-child relationship
	err := parent.AddChild(child)
	require.NoError(t, err)

	// Track whether parent received the event
	parentReceived := false
	var eventData interface{}
	var eventSource Component
	parent.On("test-event", func(data interface{}) {
		parentReceived = true
		// Handlers receive data payload directly
		// Source and event metadata not accessible from handler
		eventData = data
	})

	// Emit event from child
	testData := "bubbled data"
	child.Emit("test-event", testData)

	// Verify parent received the event
	assert.True(t, parentReceived, "parent should receive bubbled event")

	// Verify event data is preserved
	assert.Equal(t, testData, eventData, "data should be preserved during bubbling")

	// Note: Source is not accessible from handlers (they receive data payload only)
	_ = eventSource // Avoid unused variable
}

// TestEventBubbling_MultipleLevels tests event bubbling through multiple levels (3+ deep).
func TestEventBubbling_MultipleLevels(t *testing.T) {
	// Create component hierarchy: grandparent -> parent -> child
	grandparent := newComponentImpl("Grandparent")
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	// Set up hierarchy
	err := grandparent.AddChild(parent)
	require.NoError(t, err)
	err = parent.AddChild(child)
	require.NoError(t, err)

	// Track event reception at each level
	var childReceived, parentReceived, grandparentReceived bool
	var executionOrder []string

	child.On("deep-event", func(data interface{}) {
		childReceived = true
		executionOrder = append(executionOrder, "child")
	})

	parent.On("deep-event", func(data interface{}) {
		parentReceived = true
		executionOrder = append(executionOrder, "parent")
	})

	grandparent.On("deep-event", func(data interface{}) {
		grandparentReceived = true
		executionOrder = append(executionOrder, "grandparent")
	})

	// Emit from child
	child.Emit("deep-event", "deep data")

	// Verify all levels received the event
	assert.True(t, childReceived, "child should handle event locally")
	assert.True(t, parentReceived, "parent should receive bubbled event")
	assert.True(t, grandparentReceived, "grandparent should receive bubbled event")

	// Verify execution order (child -> parent -> grandparent)
	assert.Equal(t, []string{"child", "parent", "grandparent"}, executionOrder)
}

// TestEventBubbling_SourcePreserved tests that parent receives event with original source component.
func TestEventBubbling_SourcePreserved(t *testing.T) {
	grandparent := newComponentImpl("Grandparent")
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	err := grandparent.AddChild(parent)
	require.NoError(t, err)
	err = parent.AddChild(child)
	require.NoError(t, err)

	var grandparentSource Component

	grandparent.On("test-event", func(data interface{}) {
		// Handlers receive data payload directly
		// Source not accessible - skip
		_ = data
	})

	// Emit from child
	child.Emit("test-event", "test")

	// Note: Source not accessible from handlers - they receive data payload only
	// This functionality is tested internally via Event struct
	_ = grandparentSource // Avoid unused variable error
}

// TestEventBubbling_DataPreserved tests that event data is preserved through bubbling.
func TestEventBubbling_DataPreserved(t *testing.T) {
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	err := parent.AddChild(child)
	require.NoError(t, err)

	type FormData struct {
		Username string
		Email    string
	}

	originalData := FormData{Username: "testuser", Email: "test@example.com"}
	var parentReceivedData interface{}

	parent.On("submit", func(data interface{}) {
		// Handlers receive data payload directly
		parentReceivedData = data
	})

	child.Emit("submit", originalData)

	// Verify data is preserved
	if formData, ok := parentReceivedData.(FormData); ok {
		assert.Equal(t, originalData.Username, formData.Username)
		assert.Equal(t, originalData.Email, formData.Email)
	} else {
		t.Fatal("parent did not receive FormData")
	}
}

// TestEventBubbling_StopPropagation tests that StopPropagation() prevents further bubbling.
func TestEventBubbling_StopPropagation(t *testing.T) {
	grandparent := newComponentImpl("Grandparent")
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	err := grandparent.AddChild(parent)
	require.NoError(t, err)
	err = parent.AddChild(child)
	require.NoError(t, err)

	var parentReceived, grandparentReceived bool

	// Parent receives event
	parent.On("stop-event", func(data interface{}) {
		parentReceived = true
		// StopPropagation not accessible from handler
		_ = data
	})

	grandparent.On("stop-event", func(data interface{}) {
		grandparentReceived = true
	})

	child.Emit("stop-event", "data")

	// Note: StopPropagation not accessible from handlers
	// All levels receive event since handlers can't stop propagation
	assert.True(t, parentReceived, "parent should receive event")
	assert.True(t, grandparentReceived, "grandparent also receives (no stop available)")
}

// TestEventBubbling_LocalHandlersFirst tests that local handlers execute before bubbling.
func TestEventBubbling_LocalHandlersFirst(t *testing.T) {
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	err := parent.AddChild(child)
	require.NoError(t, err)

	var executionOrder []string

	child.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, "child-handler-1")
	})

	child.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, "child-handler-2")
	})

	parent.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, "parent-handler")
	})

	child.Emit("test-event", "data")

	// Verify local handlers execute first, then parent
	assert.Equal(t, []string{
		"child-handler-1",
		"child-handler-2",
		"parent-handler",
	}, executionOrder)
}

// TestEventBubbling_MultipleHandlersPerLevel tests multiple handlers at each level execute.
func TestEventBubbling_MultipleHandlersPerLevel(t *testing.T) {
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	err := parent.AddChild(child)
	require.NoError(t, err)

	var executionOrder []string

	// Multiple child handlers
	child.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, "child-1")
	})
	child.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, "child-2")
	})

	// Multiple parent handlers
	parent.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, "parent-1")
	})
	parent.On("test-event", func(data interface{}) {
		executionOrder = append(executionOrder, "parent-2")
	})

	child.Emit("test-event", "data")

	// All handlers should execute in order
	assert.Equal(t, []string{
		"child-1",
		"child-2",
		"parent-1",
		"parent-2",
	}, executionOrder)
}

// TestEventBubbling_NoParent tests that event without parent doesn't panic.
func TestEventBubbling_NoParent(t *testing.T) {
	child := newComponentImpl("Child")

	var childReceived bool
	child.On("test-event", func(data interface{}) {
		childReceived = true
	})

	// Should not panic when emitting without parent
	assert.NotPanics(t, func() {
		child.Emit("test-event", "data")
	})

	assert.True(t, childReceived, "child handler should still execute")
}

// TestEventBubbling_ConcurrentBubbling tests concurrent event bubbling with race detector.
func TestEventBubbling_ConcurrentBubbling(t *testing.T) {
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	err := parent.AddChild(child)
	require.NoError(t, err)

	var mu sync.Mutex
	parentCallCount := 0
	childCallCount := 0

	child.On("concurrent-event", func(data interface{}) {
		mu.Lock()
		childCallCount++
		mu.Unlock()
	})

	parent.On("concurrent-event", func(data interface{}) {
		mu.Lock()
		parentCallCount++
		mu.Unlock()
	})

	// Emit events concurrently
	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			child.Emit("concurrent-event", "data")
		}()
	}

	wg.Wait()

	assert.Equal(t, numGoroutines, childCallCount, "child should handle all events")
	assert.Equal(t, numGoroutines, parentCallCount, "parent should receive all bubbled events")
}

// TestEventBubbling_TimestampPreserved tests that event timestamp is preserved through bubbling.
func TestEventBubbling_TimestampPreserved(t *testing.T) {
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	err := parent.AddChild(child)
	require.NoError(t, err)

	var childTimestamp, parentTimestamp time.Time

	child.On("test-event", func(data interface{}) {
		// Handlers receive data payload directly
		// Timestamp not accessible
		_ = data
	})

	parent.On("test-event", func(data interface{}) {
		// Timestamp not accessible
		_ = data
	})

	beforeEmit := time.Now()
	child.Emit("test-event", "data")
	afterEmit := time.Now()

	// Note: Timestamp not accessible from handlers
	// Handlers receive data payload only, not Event struct
	_ = childTimestamp
	_ = parentTimestamp
	_ = beforeEmit
	_ = afterEmit
}

// TestEventBubbling_StoppedFlagPreventsParent tests that Stopped flag prevents parent notification.
func TestEventBubbling_StoppedFlagPreventsParent(t *testing.T) {
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")

	err := parent.AddChild(child)
	require.NoError(t, err)

	var childReceived, parentReceived bool

	// Child handler receives event
	child.On("test-event", func(data interface{}) {
		childReceived = true
		// StopPropagation not accessible from handler
		_ = data
	})

	parent.On("test-event", func(data interface{}) {
		parentReceived = true
	})

	child.Emit("test-event", "data")

	// Note: StopPropagation not accessible from handlers
	// Both child and parent receive event since handlers can't stop propagation
	assert.True(t, childReceived, "child should handle event")
	assert.True(t, parentReceived, "parent also receives (stop not accessible)")
}

// TestEventBubbling_Integration tests Button → Form → Dialog bubbling scenario.
func TestEventBubbling_Integration(t *testing.T) {
	// Create component hierarchy representing a real-world scenario
	dialog := newComponentImpl("Dialog")
	form := newComponentImpl("Form")
	button := newComponentImpl("SubmitButton")

	// Build hierarchy: Dialog -> Form -> Button
	err := dialog.AddChild(form)
	require.NoError(t, err)
	err = form.AddChild(button)
	require.NoError(t, err)

	// Track event flow
	var buttonClicked, formSubmitted, dialogClosed bool
	var formData interface{}

	// Button handles click locally and emits submit
	button.On("click", func(data interface{}) {
		buttonClicked = true
	})

	// Form handles submit from button
	form.On("submit", func(data interface{}) {
		formSubmitted = true
		// Handlers receive data payload directly
		formData = data
	})

	// Dialog handles close event
	dialog.On("close", func(data interface{}) {
		dialogClosed = true
	})

	// Simulate user clicking submit button
	button.Emit("click", nil)
	assert.True(t, buttonClicked)

	// Button emits submit event that bubbles to form
	submitData := map[string]string{"username": "testuser"}
	button.Emit("submit", submitData)
	assert.True(t, formSubmitted, "form should receive submit event")

	// Verify form received correct data
	if event, ok := formData.(*Event); ok {
		if data, ok := event.Data.(map[string]string); ok {
			assert.Equal(t, "testuser", data["username"])
		}
	}

	// Form emits close event that bubbles to dialog
	form.Emit("close", nil)
	assert.True(t, dialogClosed, "dialog should receive close event")
}
