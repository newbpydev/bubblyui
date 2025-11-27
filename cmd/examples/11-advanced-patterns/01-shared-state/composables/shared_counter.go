package composables

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// UseSharedCounter creates a SHARED counter instance across all components.
// This demonstrates the CreateShared pattern inspired by VueUse's createSharedComposable.
//
// Key difference from UseCounter:
// - UseCounter: Each component gets its own independent counter instance
// - UseSharedCounter: All components share the SAME counter instance
//
// This enables state synchronization across components without prop drilling or global variables.
//
// Example:
//
//	// Component A
//	counter := composables.UseSharedCounter(ctx)
//	counter.Increment() // Count becomes 1
//
//	// Component B (different component, same instance!)
//	counter := composables.UseSharedCounter(ctx)
//	// counter.Count.Get() is 1 (sees Component A's change)
var UseSharedCounter = composables.CreateShared(
	func(ctx *bubbly.Context) *CounterComposable {
		return UseCounter(ctx, 0) // Initial value: 0
	},
)
