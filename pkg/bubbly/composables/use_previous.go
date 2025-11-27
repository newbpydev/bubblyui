package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// PreviousReturn is the return value of UsePrevious.
// It provides access to the previous value of a watched ref, enabling
// comparison between current and previous states.
//
// The Value field is a Ref[*T] where:
//   - nil means no previous value exists yet (before any changes)
//   - non-nil pointer points to the actual previous value
//
// This design allows distinguishing between "no previous" and "previous is zero value".
type PreviousReturn[T any] struct {
	// Value is the previous value (nil if no previous).
	// This is a reactive ref that updates whenever the source ref changes.
	Value *bubbly.Ref[*T]
}

// Get returns the previous value (nil if none).
// This is a convenience method that unwraps the Value ref.
//
// Returns:
//   - nil if no previous value exists (before any changes to the source ref)
//   - pointer to the previous value after at least one change
//
// Example:
//
//	ref := bubbly.NewRef(10)
//	previous := UsePrevious(ctx, ref)
//
//	previous.Get()  // nil (no changes yet)
//
//	ref.Set(20)
//	previous.Get()  // *10 (previous value)
//
//	ref.Set(30)
//	previous.Get()  // *20 (previous value)
func (p *PreviousReturn[T]) Get() *T {
	return p.Value.GetTyped()
}

// UsePrevious tracks the previous value of a ref.
// It uses Watch internally to observe changes to the source ref and stores
// the old value whenever a change occurs.
//
// This composable is useful for:
//   - Comparing current and previous values in templates
//   - Detecting direction of change (increase/decrease)
//   - Implementing undo functionality (single step)
//   - Animation transitions based on value changes
//   - Form validation comparing old and new values
//
// The previous value is stored as a pointer (*T) to distinguish between:
//   - nil: No previous value exists (before any changes)
//   - *T: The actual previous value (even if it's the zero value)
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - ref: The source ref to track previous values for
//
// Returns:
//   - *PreviousReturn[T]: A struct containing the reactive previous value
//
// Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    count := bubbly.NewRef(0)
//	    previous := composables.UsePrevious(ctx, count)
//	    ctx.Expose("count", count)
//	    ctx.Expose("previous", previous)
//
//	    ctx.On("increment", func(_ interface{}) {
//	        count.Set(count.GetTyped() + 1)
//	    })
//	}).
//	WithKeyBinding("+", "increment", "Increment")
//
// Template usage:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    count := ctx.Get("count").(*bubbly.Ref[int])
//	    previous := ctx.Get("previous").(*composables.PreviousReturn[int])
//
//	    current := count.GetTyped()
//	    prev := previous.Get()
//
//	    if prev == nil {
//	        return fmt.Sprintf("Count: %d (no previous)", current)
//	    }
//
//	    direction := "→"
//	    if current > *prev {
//	        direction = "↑"
//	    } else if current < *prev {
//	        direction = "↓"
//	    }
//
//	    return fmt.Sprintf("Count: %d %s (was %d)", current, direction, *prev)
//	})
//
// Integration with CreateShared:
//
//	var sharedRef = bubbly.NewRef(0)
//	var UseSharedPrevious = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.PreviousReturn[int] {
//	        return composables.UsePrevious(ctx, sharedRef)
//	    },
//	)
func UsePrevious[T any](ctx *bubbly.Context, ref *bubbly.Ref[T]) *PreviousReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UsePrevious", time.Since(start))
	}()

	// Create ref to store previous value (initially nil - no previous yet)
	previousRef := bubbly.NewRef[*T](nil)

	// Watch the source ref for changes
	// When the source changes, store the old value as the previous
	bubbly.Watch(ref, func(newVal, oldVal T) {
		// Make a copy of oldVal to avoid aliasing issues
		// This is important because oldVal might be reused by the caller
		oldCopy := oldVal
		previousRef.Set(&oldCopy)
	})

	return &PreviousReturn[T]{
		Value: previousRef,
	}
}
