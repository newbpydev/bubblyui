package main

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// CounterComposable encapsulates counter logic and reactive state
type CounterComposable struct {
	Count     *bubbly.Ref[int]
	Increment func()
	Decrement func()
	Reset     func()
	IsEven    *bubbly.Computed[interface{}]
}

// UseCounter creates a counter composable with reactive state
// This follows Vue's Composition API pattern for reusable logic
func UseCounter(ctx *bubbly.Context, initial int) *CounterComposable {
	// Create reactive state
	count := bubbly.NewRef(initial)

	// Create computed value (automatically updates when count changes)
	isEven := ctx.Computed(func() interface{} {
		return count.Get().(int)%2 == 0
	})

	// Define methods
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

	return &CounterComposable{
		Count:     count,
		Increment: increment,
		Decrement: decrement,
		Reset:     reset,
		IsEven:    isEven,
	}
}
