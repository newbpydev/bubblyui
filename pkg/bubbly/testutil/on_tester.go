package testutil

import (
	"reflect"
	"sync"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// OnTester provides utilities for testing On directive event handler binding.
// It helps verify that:
//   - Event handlers are registered correctly
//   - Handlers are called when events are triggered
//   - Payloads are passed correctly to handlers
//   - Multiple handlers per event work correctly
//   - Handler cleanup works on unmount
//
// This tester is specifically designed for testing components that use the On
// directive for event handling. It tracks handler registrations and provides
// methods to test event handling behavior.
//
// Example:
//
//	comp := NewMockComponent("test")
//	tester := NewOnTester(comp)
//
//	// Register a handler
//	tester.RegisterHandler("click", func(data interface{}) {
//	    // Handle click
//	})
//
//	// Trigger the event
//	tester.TriggerEvent("click", "payload")
//
//	// Assert handler was called
//	tester.AssertHandlerCalled(t, "click", 1)
//	tester.AssertPayload(t, "click", "payload")
//
// Thread Safety:
//
// OnTester is thread-safe for concurrent operations using sync.RWMutex.
type OnTester struct {
	component    bubbly.Component
	handlers     map[string][]func(interface{})
	callCounts   map[string]int
	lastPayloads map[string]interface{}
	mu           sync.RWMutex
}

// NewOnTester creates a new OnTester for testing event handler binding.
//
// The component parameter should be a Component that supports event handling.
// The tester will register tracking handlers on this component to monitor
// event handling behavior.
//
// Parameters:
//   - comp: A Component that supports event handling
//
// Returns:
//   - *OnTester: A new tester instance
//
// Example:
//
//	comp := NewMockComponent("test")
//	tester := NewOnTester(comp)
func NewOnTester(comp bubbly.Component) *OnTester {
	return &OnTester{
		component:    comp,
		handlers:     make(map[string][]func(interface{})),
		callCounts:   make(map[string]int),
		lastPayloads: make(map[string]interface{}),
	}
}

// RegisterHandler registers a test handler for the specified event.
//
// This method stores the handler and registers a tracking wrapper on the
// component that increments call counts and stores payloads. When the event
// is triggered, both the tracking wrapper and the user's handler will be called.
//
// Parameters:
//   - event: The event name to register the handler for
//   - handler: The handler function to call when the event is triggered
//
// Example:
//
//	tester.RegisterHandler("click", func(data interface{}) {
//	    fmt.Println("Clicked:", data)
//	})
func (ot *OnTester) RegisterHandler(event string, handler func(interface{})) {
	ot.mu.Lock()
	defer ot.mu.Unlock()

	// Store the user's handler
	ot.handlers[event] = append(ot.handlers[event], handler)

	// Register a tracking wrapper on the component
	if ot.component != nil {
		ot.component.On(event, func(data interface{}) {
			// Track the call
			ot.mu.Lock()
			ot.callCounts[event]++
			ot.lastPayloads[event] = data
			ot.mu.Unlock()

			// Call the user's handler
			if handler != nil {
				handler(data)
			}
		})
	}
}

// TriggerEvent triggers the specified event with the given payload.
//
// This method calls component.Emit() to trigger all registered handlers
// for the event. The handlers will receive the payload as their argument.
//
// Parameters:
//   - name: The event name to trigger
//   - payload: The data to pass to the handlers
//
// Example:
//
//	tester.TriggerEvent("click", "button1")
//	tester.TriggerEvent("submit", map[string]string{"field": "value"})
func (ot *OnTester) TriggerEvent(name string, payload interface{}) {
	if ot.component != nil {
		ot.component.Emit(name, payload)
	}
}

// AssertHandlerCalled asserts that the handler for the specified event
// was called the expected number of times.
//
// This method checks the call count for the event and reports an error
// if it doesn't match the expected value.
//
// Parameters:
//   - t: The testing.T instance for assertions
//   - event: The event name to check
//   - times: The expected number of times the handler should have been called
//
// Example:
//
//	tester.TriggerEvent("click", nil)
//	tester.TriggerEvent("click", nil)
//	tester.AssertHandlerCalled(t, "click", 2)
func (ot *OnTester) AssertHandlerCalled(t testingT, event string, times int) {
	t.Helper()

	ot.mu.RLock()
	actual := ot.callCounts[event]
	ot.mu.RUnlock()

	if actual != times {
		t.Errorf("Expected handler for event %q to be called %d times, but was called %d times",
			event, times, actual)
	}
}

// AssertPayload asserts that the last payload for the specified event
// matches the expected value.
//
// This method uses reflect.DeepEqual to compare the payloads, which works
// correctly for all Go types including slices, maps, and structs.
//
// Parameters:
//   - t: The testing.T instance for assertions
//   - event: The event name to check
//   - expected: The expected payload value
//
// Example:
//
//	tester.TriggerEvent("click", "test data")
//	tester.AssertPayload(t, "click", "test data")
func (ot *OnTester) AssertPayload(t testingT, event string, expected interface{}) {
	t.Helper()

	ot.mu.RLock()
	actual := ot.lastPayloads[event]
	ot.mu.RUnlock()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected payload for event %q to be %v, but got %v",
			event, expected, actual)
	}
}

// GetCallCount returns the number of times the handler for the specified
// event has been called.
//
// This is a helper method for custom assertions or debugging. It returns 0
// if the event has never been triggered.
//
// Parameters:
//   - event: The event name to check
//
// Returns:
//   - int: The number of times the handler was called
//
// Example:
//
//	count := tester.GetCallCount("click")
//	if count > 5 {
//	    t.Error("Too many clicks")
//	}
func (ot *OnTester) GetCallCount(event string) int {
	ot.mu.RLock()
	defer ot.mu.RUnlock()

	return ot.callCounts[event]
}

// GetLastPayload returns the last payload received for the specified event.
//
// This is a helper method for custom assertions or debugging. It returns nil
// if the event has never been triggered.
//
// Parameters:
//   - event: The event name to check
//
// Returns:
//   - interface{}: The last payload, or nil if never triggered
//
// Example:
//
//	payload := tester.GetLastPayload("click")
//	if data, ok := payload.(string); ok {
//	    fmt.Println("Last click data:", data)
//	}
func (ot *OnTester) GetLastPayload(event string) interface{} {
	ot.mu.RLock()
	defer ot.mu.RUnlock()

	return ot.lastPayloads[event]
}
