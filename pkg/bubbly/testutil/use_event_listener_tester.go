package testutil

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseEventListenerTester provides utilities for testing event listeners.
// It integrates with the UseEventListener composable to test event subscription,
// handler execution, and cleanup in a deterministic way.
//
// This tester is specifically designed for testing components that use the UseEventListener
// composable. It allows you to:
//   - Emit events to trigger handlers
//   - Trigger manual cleanup
//   - Track handler call counts
//   - Verify event handling behavior
//
// The tester automatically extracts the cleanup function from the component,
// making it easy to test event listener lifecycle.
//
// Example:
//
//	comp := createEventListenerComponent() // Component using UseEventListener
//	tester := NewUseEventListenerTester(comp)
//
//	// Emit event
//	tester.EmitEvent("click", nil)
//	assert.Equal(t, 1, tester.GetCallCount("clickCount"))
//
//	// Trigger cleanup
//	tester.TriggerCleanup()
//
//	// Event should no longer trigger handler
//	tester.EmitEvent("click", nil)
//	assert.Equal(t, 1, tester.GetCallCount("clickCount"))
//
// Thread Safety:
//
// UseEventListenerTester is not thread-safe. It should only be used from a single test goroutine.
type UseEventListenerTester struct {
	component bubbly.Component
	cleanup   func()
}

// NewUseEventListenerTester creates a new UseEventListenerTester for testing event listeners.
//
// The component must expose "cleanup" in its Setup function.
// This corresponds to the cleanup function returned by UseEventListener composable.
//
// Parameters:
//   - comp: The component to test (must expose cleanup function)
//
// Returns:
//   - *UseEventListenerTester: A new tester instance
//
// Panics:
//   - If the component doesn't expose the cleanup function
//
// Example:
//
//	comp, err := bubbly.NewComponent("TestEventListener").
//	    Setup(func(ctx *bubbly.Context) {
//	        cleanup := composables.UseEventListener(ctx, "click", handleClick)
//	        ctx.Expose("cleanup", cleanup)
//	    }).
//	    Build()
//	comp.Init()
//
//	tester := NewUseEventListenerTester(comp)
func NewUseEventListenerTester(comp bubbly.Component) *UseEventListenerTester {
	// Extract cleanup function
	cleanupValue := extractExposedValue(comp, "cleanup")
	if cleanupValue == nil {
		panic("component must expose 'cleanup' function")
	}

	cleanup, ok := cleanupValue.(func())
	if !ok {
		panic("'cleanup' must be a function with signature func()")
	}

	return &UseEventListenerTester{
		component: comp,
		cleanup:   cleanup,
	}
}

// EmitEvent emits an event to trigger the event listener.
//
// Parameters:
//   - event: The event name
//   - data: The event data (can be nil)
//
// Example:
//
//	tester.EmitEvent("click", nil)
//	tester.EmitEvent("submit", map[string]interface{}{"value": "test"})
func (uelt *UseEventListenerTester) EmitEvent(event string, data interface{}) {
	uelt.component.Emit(event, data)
}

// TriggerCleanup triggers the cleanup function to unregister the event listener.
// After cleanup, emitted events should no longer trigger the handler.
//
// Example:
//
//	tester.TriggerCleanup()
//	tester.EmitEvent("click", nil) // Handler should not execute
func (uelt *UseEventListenerTester) TriggerCleanup() {
	uelt.cleanup()
}

// GetCallCount returns the call count for an event handler.
// This requires the component to expose a counter variable.
//
// Parameters:
//   - counterName: The name of the exposed counter variable
//
// Returns:
//   - int: The current call count
//
// Example:
//
//	count := tester.GetCallCount("clickCount")
//	assert.Equal(t, 2, count)
func (uelt *UseEventListenerTester) GetCallCount(counterName string) int {
	// Extract the counter value
	counterValue := extractExposedValue(uelt.component, counterName)
	if counterValue == nil {
		return 0
	}

	// The counter is a pointer to int
	if ptr, ok := counterValue.(*int); ok {
		return *ptr
	}

	return 0
}
