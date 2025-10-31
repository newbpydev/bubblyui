package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// UseStateReturn is the return type for the UseState composable.
// It provides a reactive state value with convenient getter and setter methods.
//
// Fields:
//   - Value: The underlying reactive reference
//   - Set: Function to update the state value
//   - Get: Function to retrieve the current state value
//
// Example:
//
//	state := UseState(ctx, 0)
//	state.Set(42)
//	value := state.GetTyped()  // Returns 42
type UseStateReturn[T any] struct {
	// Value is the underlying reactive reference.
	// Can be used directly for advanced operations like watching.
	Value *bubbly.Ref[T]

	// Set updates the state value and triggers reactivity.
	Set func(T)

	// Get retrieves the current state value.
	Get func() T
}

// UseState creates a simple reactive state with getter and setter methods.
// This is the most basic composable for state management, wrapping a Ref
// with a convenient API.
//
// UseState is type-safe using Go generics, ensuring compile-time type checking.
// Multiple calls to UseState create independent state instances.
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - initial: The initial value for the state
//
// Returns:
//   - UseStateReturn[T]: A struct containing the reactive value and methods
//
// Example:
//
//	Setup(func(ctx *Context) {
//	    count := UseState(ctx, 0)
//
//	    ctx.On("increment", func(_ interface{}) {
//	        count.Set(count.GetTyped() + 1)
//	    })
//
//	    ctx.Expose("count", count.Value)
//	})
//
// Type Safety Example:
//
//	// Different types work independently
//	name := UseState(ctx, "Alice")    // UseStateReturn[string]
//	age := UseState(ctx, 30)          // UseStateReturn[int]
//	active := UseState(ctx, true)     // UseStateReturn[bool]
//
// Multiple Instances:
//
//	// Each instance is independent
//	counter1 := UseState(ctx, 0)
//	counter2 := UseState(ctx, 0)
//	counter1.Set(10)  // counter2 remains 0
func UseState[T any](ctx *bubbly.Context, initial T) UseStateReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseState", time.Since(start))
	}()

	// Create the underlying reactive reference
	value := bubbly.NewRef(initial)

	// Return the composable interface with convenient methods
	return UseStateReturn[T]{
		Value: value,
		Set: func(v T) {
			value.Set(v)
		},
		Get: func() T {
			return value.GetTyped()
		},
	}
}
