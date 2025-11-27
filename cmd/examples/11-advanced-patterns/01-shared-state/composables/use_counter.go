package composables

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// CounterComposable provides counter state and operations
type CounterComposable struct {
	Count     *bubbly.Ref[interface{}]
	Doubled   *bubbly.Computed[interface{}]
	IsEven    *bubbly.Computed[interface{}]
	History   *bubbly.Ref[interface{}]
	Increment func()
	Decrement func()
	Reset     func()
}

// UseCounter creates a counter composable with state and operations
func UseCounter(ctx *bubbly.Context, initial int) *CounterComposable {
	count := ctx.Ref(initial)
	history := ctx.Ref([]int{initial})

	doubled := ctx.Computed(func() interface{} {
		return count.Get().(int) * 2
	})

	isEven := ctx.Computed(func() interface{} {
		return count.Get().(int)%2 == 0
	})

	addToHistory := func(newVal int) {
		hist := history.Get().([]int)
		if len(hist) >= 5 {
			hist = hist[1:]
		}
		history.Set(append(hist, newVal))
	}

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
		count.Set(initial)
		history.Set([]int{initial})
	}

	return &CounterComposable{
		Count:     count,
		Doubled:   doubled,
		IsEven:    isEven,
		History:   history,
		Increment: increment,
		Decrement: decrement,
		Reset:     reset,
	}
}
