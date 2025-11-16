package testutil

import (
	"reflect"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseEffectTester provides utilities for testing side effects and cleanup functions.
// It integrates with the UseEffect composable to test effect execution, cleanup timing,
// and dependency tracking in a deterministic way.
//
// This tester is specifically designed for testing components that use the UseEffect
// composable. It allows you to:
//   - Trigger lifecycle events (mount, update, unmount)
//   - Verify effect execution
//   - Verify cleanup execution
//   - Test dependency tracking
//   - Track call counts
//
// The tester provides methods to manually trigger lifecycle events that would
// normally be triggered by the component lifecycle.
//
// Example:
//
//	comp := createEffectComponent() // Component using UseEffect
//	tester := NewUseEffectTester(comp)
//
//	// Verify effect called on mount
//	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"))
//
//	// Trigger update
//	tester.TriggerUpdate()
//
//	// Verify cleanup and re-execution
//	assert.Equal(t, 1, tester.GetCleanupCallCount("cleanupCalled"))
//	assert.Equal(t, 2, tester.GetEffectCallCount("effectCalled"))
//
//	// Trigger unmount
//	tester.TriggerUnmount()
//
//	// Verify final cleanup
//	assert.Equal(t, 2, tester.GetCleanupCallCount("cleanupCalled"))
//
// Thread Safety:
//
// UseEffectTester is not thread-safe. It should only be used from a single test goroutine.
type UseEffectTester struct {
	component bubbly.Component
}

// NewUseEffectTester creates a new UseEffectTester for testing side effects.
//
// The component should use UseEffect composable in its Setup function.
// The tester doesn't require specific exposed values, but you should expose
// counters or flags to verify effect and cleanup execution.
//
// Parameters:
//   - comp: The component to test (should use UseEffect)
//
// Returns:
//   - *UseEffectTester: A new tester instance
//
// Example:
//
//	comp, err := bubbly.NewComponent("TestEffect").
//	    Setup(func(ctx *bubbly.Context) {
//	        effectCalled := 0
//	        composables.UseEffect(ctx, func() composables.UseEffectCleanup {
//	            effectCalled++
//	            return func() {
//	                // cleanup
//	            }
//	        })
//	        ctx.Expose("effectCalled", &effectCalled)
//	    }).
//	    Build()
//	comp.Init()
//
//	tester := NewUseEffectTester(comp)
func NewUseEffectTester(comp bubbly.Component) *UseEffectTester {
	return &UseEffectTester{
		component: comp,
	}
}

// TriggerUpdate triggers the component's update lifecycle.
// This will cause UseEffect to re-run effects (either all effects if no dependencies,
// or only effects whose dependencies have changed).
//
// Before re-running an effect, its cleanup function (if any) will be called.
//
// Example:
//
//	tester.TriggerUpdate()
//	assert.Equal(t, 2, tester.GetEffectCallCount("effectCalled"))
func (uet *UseEffectTester) TriggerUpdate() {
	// Trigger the component's update by calling Update with a nil message
	// This will invoke the OnUpdated lifecycle hooks
	uet.component.Update(nil)
}

// TriggerUnmount triggers the component's unmount lifecycle.
// This will cause UseEffect to call all cleanup functions for all effects.
//
// Example:
//
//	tester.TriggerUnmount()
//	assert.Equal(t, 1, tester.GetCleanupCallCount("cleanupCalled"))
func (uet *UseEffectTester) TriggerUnmount() {
	// Use type assertion to access the Unmount method
	// Component.Unmount() is now public, so we can call it directly
	type unmounter interface {
		Unmount()
	}
	if u, ok := uet.component.(unmounter); ok {
		u.Unmount()
	}
}

// SetRefValue sets a ref value to trigger dependency changes.
// This is useful for testing effects with dependencies.
//
// Parameters:
//   - refName: The name of the exposed ref
//   - value: The new value to set
//
// Example:
//
//	tester.SetRefValue("count", 5)
//	tester.TriggerUpdate()
//	// Effect should re-run if it depends on "count"
func (uet *UseEffectTester) SetRefValue(refName string, value interface{}) {
	// Extract the ref and set its value
	refValue := extractExposedValue(uet.component, refName)
	if refValue == nil {
		return
	}

	// Use reflection to call Set() on the ref
	setRefValue(refValue, value)
}

// setRefValue is a helper to set a ref value using reflection
func setRefValue(refValue interface{}, value interface{}) {
	// Use reflection to call Set() method on the ref
	v := reflect.ValueOf(refValue)
	if !v.IsValid() || v.IsNil() {
		return
	}

	// Call Set() method
	setMethod := v.MethodByName("Set")
	if !setMethod.IsValid() {
		return
	}

	// Call Set with the value
	setMethod.Call([]reflect.Value{reflect.ValueOf(value)})
}

// GetEffectCallCount returns the call count for an effect.
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
//	count := tester.GetEffectCallCount("effectCalled")
//	assert.Equal(t, 2, count)
func (uet *UseEffectTester) GetEffectCallCount(counterName string) int {
	// Extract the counter value
	counterValue := extractExposedValue(uet.component, counterName)
	if counterValue == nil {
		return 0
	}

	// The counter is a pointer to int
	if ptr, ok := counterValue.(*int); ok {
		return *ptr
	}

	return 0
}

// GetCleanupCallCount returns the call count for a cleanup function.
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
//	count := tester.GetCleanupCallCount("cleanupCalled")
//	assert.Equal(t, 1, count)
func (uet *UseEffectTester) GetCleanupCallCount(counterName string) int {
	// Same implementation as GetEffectCallCount
	return uet.GetEffectCallCount(counterName)
}
