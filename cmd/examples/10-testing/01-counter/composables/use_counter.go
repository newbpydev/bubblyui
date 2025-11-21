package composables

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// CounterComposable provides counter state and operations
type CounterComposable struct {
	Count     *bubbly.Ref[interface{}]
	History   *bubbly.Ref[interface{}]
	Doubled   *bubbly.Computed[interface{}]
	IsEven    *bubbly.Computed[interface{}]
	Increment func()
	Decrement func()
	Reset     func()
	SetValue  func(int)
}

// UseCounter creates a counter composable with state and operations
func UseCounter(ctx *bubbly.Context, initialCount int) *CounterComposable {
	// Create reactive state
	count := ctx.Ref(initialCount)
	history := ctx.Ref([]int{initialCount})

	// Computed values - reactive derived state
	doubled := ctx.Computed(func() interface{} {
		return count.Get().(int) * 2
	})

	isEven := ctx.Computed(func() interface{} {
		return count.Get().(int)%2 == 0
	})

	// Helper to add to history
	addToHistory := func(newVal int) {
		hist := history.Get().([]int)
		// Keep last 5 values
		if len(hist) >= 5 {
			hist = hist[1:]
		}
		history.Set(append(hist, newVal))
	}

	// Operations
	increment := func() {
		newVal := count.Get().(int) + 1
		count.Set(newVal)
		addToHistory(newVal)
	}

	decrement := func() {
		newVal := count.Get().(int) - 1
		count.Set(newVal)
		addToHistory(newVal)
	}

	reset := func() {
		count.Set(initialCount)
		history.Set([]int{initialCount})
	}

	setValue := func(val int) {
		count.Set(val)
		addToHistory(val)
	}

	return &CounterComposable{
		Count:     count,
		History:   history,
		Doubled:   doubled,
		IsEven:    isEven,
		Increment: increment,
		Decrement: decrement,
		Reset:     reset,
		SetValue:  setValue,
	}
}
