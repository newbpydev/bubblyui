package composables

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// CounterComposable encapsulates counter logic (Vue-like composable)
type CounterComposable struct {
	Count     *bubbly.Ref[int]
	Increment func()
	Decrement func()
	Reset     func()
	IsEven    *bubbly.Computed[interface{}]
}

// UseCounter creates a reusable counter with reactive state
// This demonstrates the composable pattern - reusable logic that can be shared
// across components, similar to Vue's Composition API
func UseCounter(ctx *bubbly.Context, initial int) *CounterComposable {
	// Create reactive state
	count := bubbly.NewRef(initial)

	// Create computed value (automatically updates when count changes)
	isEven := ctx.Computed(func() interface{} {
		return count.Get().(int)%2 == 0
	})

	// Define methods that operate on state
	increment := func() {
		current := count.Get().(int)
		count.Set(current + 1)
	}

	decrement := func() {
		current := count.Get().(int)
		count.Set(current - 1)
	}

	reset := func() {
		count.Set(initial)
	}

	// Return composable with state and methods
	return &CounterComposable{
		Count:     count,
		Increment: increment,
		Decrement: decrement,
		Reset:     reset,
		IsEven:    isEven,
	}
}
