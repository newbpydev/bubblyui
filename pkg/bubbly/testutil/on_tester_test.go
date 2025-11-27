package testutil

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestNewOnTester verifies OnTester creation.
func TestNewOnTester(t *testing.T) {
	tests := []struct {
		name      string
		component bubbly.Component
		wantNil   bool
	}{
		{
			name:      "with valid component",
			component: NewMockComponent("test"),
			wantNil:   false,
		},
		{
			name:      "with nil component",
			component: nil,
			wantNil:   false, // Tester created, but component is nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewOnTester(tt.component)

			if tt.wantNil {
				assert.Nil(t, tester)
			} else {
				assert.NotNil(t, tester)
				assert.Equal(t, tt.component, tester.component)
			}
		})
	}
}

// TestOnTester_RegisterHandler verifies handler registration.
func TestOnTester_RegisterHandler(t *testing.T) {
	tests := []struct {
		name      string
		event     string
		wantPanic bool
	}{
		{
			name:      "register click handler",
			event:     "click",
			wantPanic: false,
		},
		{
			name:      "register submit handler",
			event:     "submit",
			wantPanic: false,
		},
		{
			name:      "register empty event name",
			event:     "",
			wantPanic: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewMockComponent("test")
			tester := NewOnTester(comp)

			handler := func(data interface{}) {
				// Handler implementation
			}

			if tt.wantPanic {
				assert.Panics(t, func() {
					tester.RegisterHandler(tt.event, handler)
				})
			} else {
				assert.NotPanics(t, func() {
					tester.RegisterHandler(tt.event, handler)
				})

				// Verify handler was stored
				assert.Contains(t, tester.handlers, tt.event)
			}
		})
	}
}

// TestOnTester_TriggerEvent verifies event triggering.
func TestOnTester_TriggerEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   string
		payload interface{}
	}{
		{
			name:    "trigger with string payload",
			event:   "click",
			payload: "test data",
		},
		{
			name:    "trigger with int payload",
			event:   "count",
			payload: 42,
		},
		{
			name:    "trigger with nil payload",
			event:   "empty",
			payload: nil,
		},
		{
			name:    "trigger with struct payload",
			event:   "data",
			payload: struct{ Value string }{Value: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewMockComponent("test")
			tester := NewOnTester(comp)

			var receivedPayload interface{}
			handler := func(data interface{}) {
				receivedPayload = data
			}

			tester.RegisterHandler(tt.event, handler)
			tester.TriggerEvent(tt.event, tt.payload)

			// Verify handler was called with correct payload
			assert.Equal(t, tt.payload, receivedPayload)
		})
	}
}

// TestOnTester_AssertHandlerCalled verifies handler call assertions.
func TestOnTester_AssertHandlerCalled(t *testing.T) {
	tests := []struct {
		name          string
		event         string
		triggerCount  int
		expectedTimes int
		shouldFail    bool
	}{
		{
			name:          "handler called once",
			event:         "click",
			triggerCount:  1,
			expectedTimes: 1,
			shouldFail:    false,
		},
		{
			name:          "handler called multiple times",
			event:         "click",
			triggerCount:  5,
			expectedTimes: 5,
			shouldFail:    false,
		},
		{
			name:          "handler not called",
			event:         "click",
			triggerCount:  0,
			expectedTimes: 0,
			shouldFail:    false,
		},
		{
			name:          "handler called wrong number of times",
			event:         "click",
			triggerCount:  3,
			expectedTimes: 5,
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewMockComponent("test")
			tester := NewOnTester(comp)

			handler := func(data interface{}) {}
			tester.RegisterHandler(tt.event, handler)

			// Trigger event the specified number of times
			for i := 0; i < tt.triggerCount; i++ {
				tester.TriggerEvent(tt.event, nil)
			}

			// Use mock testing.T to capture assertion failures
			mockT := &mockTestingT{}
			tester.AssertHandlerCalled(mockT, tt.event, tt.expectedTimes)

			if tt.shouldFail {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			} else {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			}
		})
	}
}

// TestOnTester_AssertPayload verifies payload assertions.
func TestOnTester_AssertPayload(t *testing.T) {
	tests := []struct {
		name       string
		event      string
		payload    interface{}
		expected   interface{}
		shouldFail bool
	}{
		{
			name:       "string payload matches",
			event:      "click",
			payload:    "test",
			expected:   "test",
			shouldFail: false,
		},
		{
			name:       "int payload matches",
			event:      "count",
			payload:    42,
			expected:   42,
			shouldFail: false,
		},
		{
			name:       "nil payload matches",
			event:      "empty",
			payload:    nil,
			expected:   nil,
			shouldFail: false,
		},
		{
			name:       "payload mismatch",
			event:      "click",
			payload:    "actual",
			expected:   "expected",
			shouldFail: true,
		},
		{
			name:       "struct payload matches",
			event:      "data",
			payload:    struct{ Value string }{Value: "test"},
			expected:   struct{ Value string }{Value: "test"},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewMockComponent("test")
			tester := NewOnTester(comp)

			handler := func(data interface{}) {}
			tester.RegisterHandler(tt.event, handler)
			tester.TriggerEvent(tt.event, tt.payload)

			// Use mock testing.T to capture assertion failures
			mockT := &mockTestingT{}
			tester.AssertPayload(mockT, tt.event, tt.expected)

			if tt.shouldFail {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			} else {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			}
		})
	}
}

// TestOnTester_MultipleHandlers verifies multiple handlers per event.
func TestOnTester_MultipleHandlers(t *testing.T) {
	comp := NewMockComponent("test")
	tester := NewOnTester(comp)

	// Register multiple handlers for same event
	handler1Called := false
	handler2Called := false
	handler3Called := false

	tester.RegisterHandler("click", func(data interface{}) {
		handler1Called = true
	})
	tester.RegisterHandler("click", func(data interface{}) {
		handler2Called = true
	})
	tester.RegisterHandler("click", func(data interface{}) {
		handler3Called = true
	})

	// Trigger event once
	tester.TriggerEvent("click", nil)

	// All handlers should be called
	assert.True(t, handler1Called, "Handler 1 should be called")
	assert.True(t, handler2Called, "Handler 2 should be called")
	assert.True(t, handler3Called, "Handler 3 should be called")

	// Call count should be 3 (once per handler)
	assert.Equal(t, 3, tester.GetCallCount("click"))
}

// TestOnTester_GetCallCount verifies call count retrieval.
func TestOnTester_GetCallCount(t *testing.T) {
	tests := []struct {
		name         string
		event        string
		triggerCount int
		want         int
	}{
		{
			name:         "no calls",
			event:        "click",
			triggerCount: 0,
			want:         0,
		},
		{
			name:         "single call",
			event:        "click",
			triggerCount: 1,
			want:         1,
		},
		{
			name:         "multiple calls",
			event:        "click",
			triggerCount: 10,
			want:         10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewMockComponent("test")
			tester := NewOnTester(comp)

			handler := func(data interface{}) {}
			tester.RegisterHandler(tt.event, handler)

			for i := 0; i < tt.triggerCount; i++ {
				tester.TriggerEvent(tt.event, nil)
			}

			got := tester.GetCallCount(tt.event)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestOnTester_GetLastPayload verifies last payload retrieval.
func TestOnTester_GetLastPayload(t *testing.T) {
	tests := []struct {
		name     string
		event    string
		payloads []interface{}
		want     interface{}
	}{
		{
			name:     "single payload",
			event:    "click",
			payloads: []interface{}{"first"},
			want:     "first",
		},
		{
			name:     "multiple payloads - last one wins",
			event:    "click",
			payloads: []interface{}{"first", "second", "third"},
			want:     "third",
		},
		{
			name:     "nil payload",
			event:    "empty",
			payloads: []interface{}{nil},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewMockComponent("test")
			tester := NewOnTester(comp)

			handler := func(data interface{}) {}
			tester.RegisterHandler(tt.event, handler)

			for _, payload := range tt.payloads {
				tester.TriggerEvent(tt.event, payload)
			}

			got := tester.GetLastPayload(tt.event)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestOnTester_ThreadSafety verifies concurrent access is safe.
func TestOnTester_ThreadSafety(t *testing.T) {
	comp := NewMockComponent("test")
	tester := NewOnTester(comp)

	// Register handler
	handler := func(data interface{}) {}
	tester.RegisterHandler("click", handler)

	// Concurrent triggers
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			tester.TriggerEvent("click", val)
		}(i)
	}

	wg.Wait()

	// Should have 100 calls
	assert.Equal(t, 100, tester.GetCallCount("click"))
}

// TestOnTester_NilComponent verifies safe handling of nil component.
func TestOnTester_NilComponent(t *testing.T) {
	tester := NewOnTester(nil)

	// Should not panic
	assert.NotPanics(t, func() {
		tester.RegisterHandler("click", func(data interface{}) {})
		tester.TriggerEvent("click", "test")
		tester.GetCallCount("click")
		tester.GetLastPayload("click")
	})
}

// TestOnTester_UnregisteredEvent verifies behavior with unregistered events.
func TestOnTester_UnregisteredEvent(t *testing.T) {
	comp := NewMockComponent("test")
	tester := NewOnTester(comp)

	// Trigger event without registering handler
	assert.NotPanics(t, func() {
		tester.TriggerEvent("unregistered", "data")
	})

	// Call count should be 0
	assert.Equal(t, 0, tester.GetCallCount("unregistered"))

	// Last payload should be nil
	assert.Nil(t, tester.GetLastPayload("unregistered"))
}
