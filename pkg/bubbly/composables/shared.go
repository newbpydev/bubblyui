package composables

import (
	"sync"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateShared wraps a composable factory to return a singleton instance.
// Inspired by VueUse's createSharedComposable, this enables sharing state
// and logic across multiple components without prop drilling or global variables.
//
// The factory function is called exactly once (thread-safe via sync.Once),
// and all subsequent calls return the same instance.
//
// Type-safe with Go generics.
//
// Example:
//
//	var UseSharedCounter = CreateShared(
//	  func(ctx *bubbly.Context) *CounterComposable {
//	    return UseCounter(ctx, 0)
//	  },
//	)
//
//	// In any component - same instance across all
//	counter := UseSharedCounter(ctx)
//
// Note: The context passed to the factory is from the first component that
// calls the shared composable. Ensure the factory doesn't rely on
// component-specific context state.
func CreateShared[T any](factory func(*bubbly.Context) T) func(*bubbly.Context) T {
	var instance T
	var once sync.Once

	return func(ctx *bubbly.Context) T {
		once.Do(func() {
			instance = factory(ctx)
		})
		return instance
	}
}
