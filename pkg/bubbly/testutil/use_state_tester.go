package testutil

import (
	"reflect"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseStateTester provides utilities for testing simple state management.
// It integrates with the UseState composable to test state get/set operations
// in a deterministic way.
//
// This tester is specifically designed for testing components that use the UseState
// composable. It allows you to:
//   - Set state values
//   - Get current state values
//   - Verify state updates
//
// The tester automatically extracts the state refs from the component,
// making it easy to assert on state behavior at any point in the test.
//
// Example:
//
//	comp := createStateComponent() // Component using UseState
//	tester := NewUseStateTester[string](comp)
//
//	// Get initial value
//	assert.Equal(t, "initial", tester.GetValue())
//
//	// Set new value
//	tester.SetValue("updated")
//
//	// Verify update
//	assert.Equal(t, "updated", tester.GetValue())
//
// Thread Safety:
//
// UseStateTester is not thread-safe. It should only be used from a single test goroutine.
type UseStateTester[T any] struct {
	component bubbly.Component
	valueRef  interface{} // *Ref[T]
	set       func(T)
	get       func() T
}

// NewUseStateTester creates a new UseStateTester for testing state operations.
//
// The component must expose "value", "set", and "get" in its Setup function.
// These correspond to the fields returned by UseState composable.
//
// Parameters:
//   - comp: The component to test (must expose state value and methods)
//
// Returns:
//   - *UseStateTester[T]: A new tester instance
//
// Panics:
//   - If the component doesn't expose required refs or functions
//
// Example:
//
//	comp, err := bubbly.NewComponent("TestState").
//	    Setup(func(ctx *bubbly.Context) {
//	        state := composables.UseState(ctx, "initial")
//	        ctx.Expose("value", state.Value)
//	        ctx.Expose("set", state.Set)
//	        ctx.Expose("get", state.Get)
//	    }).
//	    Build()
//	comp.Init()
//
//	tester := NewUseStateTester[string](comp)
func NewUseStateTester[T any](comp bubbly.Component) *UseStateTester[T] {
	// Extract exposed values from component using reflection

	// Get value ref
	valueRef := extractExposedValue(comp, "value")
	if valueRef == nil {
		panic("component must expose 'value' ref")
	}

	// Extract set function
	setValue := extractExposedValue(comp, "set")
	if setValue == nil {
		panic("component must expose 'set' function")
	}
	set, ok := setValue.(func(T))
	if !ok {
		panic("'set' must be a function with signature func(T)")
	}

	// Extract get function
	getValue := extractExposedValue(comp, "get")
	if getValue == nil {
		panic("component must expose 'get' function")
	}
	get, ok := getValue.(func() T)
	if !ok {
		panic("'get' must be a function with signature func() T")
	}

	return &UseStateTester[T]{
		component: comp,
		valueRef:  valueRef,
		set:       set,
		get:       get,
	}
}

// SetValue sets a new state value.
//
// Parameters:
//   - value: The value to set
//
// Example:
//
//	tester.SetValue("new value")
//	assert.Equal(t, "new value", tester.GetValue())
func (ust *UseStateTester[T]) SetValue(value T) {
	ust.set(value)
}

// GetValue returns the current state value using the get function.
//
// Returns:
//   - T: The current value
//
// Example:
//
//	value := tester.GetValue()
//	assert.Equal(t, "expected", value)
func (ust *UseStateTester[T]) GetValue() T {
	return ust.get()
}

// GetValueFromRef returns the current value by reading directly from the ref.
// This is an alternative to GetValue() that uses reflection to access the ref.
//
// Returns:
//   - T: The current value
//
// Example:
//
//	value := tester.GetValueFromRef()
//	assert.Equal(t, "expected", value)
func (ust *UseStateTester[T]) GetValueFromRef() T {
	// Use reflection to call Get() on the typed ref
	v := reflect.ValueOf(ust.valueRef)
	if !v.IsValid() || v.IsNil() {
		var zero T
		return zero
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		var zero T
		return zero
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		var zero T
		return zero
	}

	// Return the typed value
	return result[0].Interface().(T)
}
