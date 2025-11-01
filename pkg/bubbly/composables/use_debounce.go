package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseDebounce creates a debounced version of a reactive value.
// It delays updating the debounced ref until the specified delay has passed
// without any new changes to the source value.
//
// This is useful for:
//   - Search input fields (wait for user to stop typing)
//   - Window resize handlers (wait for resize to complete)
//   - Form validation (wait for user to finish editing)
//   - API calls (batch rapid changes into single request)
//
// The debounced ref starts with the same value as the source.
// When the source changes, a timer is started. If the source changes again
// before the timer expires, the timer is reset. Only when the timer expires
// without interruption does the debounced ref update to the latest value.
//
// Timer cleanup is automatically registered with the component's lifecycle,
// ensuring no goroutine leaks when the component unmounts.
//
// Type parameter T must match the type of the source ref, ensuring compile-time
// type safety for both the source and debounced values.
//
// Parameters:
//   - ctx: The component context (required for lifecycle management)
//   - value: The source reactive value to debounce
//   - delay: How long to wait after the last change before updating
//
// Returns:
//   - *Ref[T]: A new reactive reference that updates after the delay
//
// Example - Search Input:
//
//	Setup(func(ctx *Context) {
//	    searchTerm := ctx.Ref("")
//	    debouncedSearch := UseDebounce(ctx, searchTerm, 300*time.Millisecond)
//
//	    // Watch debounced value for API calls
//	    ctx.Watch(debouncedSearch, func(newVal, _ string) {
//	        if newVal != "" {
//	            performSearch(newVal)
//	        }
//	    })
//
//	    ctx.Expose("searchTerm", searchTerm)
//	})
//
// Example - Window Resize:
//
//	Setup(func(ctx *Context) {
//	    windowWidth := ctx.Ref(800)
//	    debouncedWidth := UseDebounce(ctx, windowWidth, 150*time.Millisecond)
//
//	    // Only recalculate layout after resize stops
//	    ctx.Watch(debouncedWidth, func(newVal, _ int) {
//	        recalculateLayout(newVal)
//	    })
//	})
//
// Example - Form Validation:
//
//	Setup(func(ctx *Context) {
//	    email := ctx.Ref("")
//	    debouncedEmail := UseDebounce(ctx, email, 500*time.Millisecond)
//
//	    // Validate after user stops typing
//	    ctx.Watch(debouncedEmail, func(newVal, _ string) {
//	        if !isValidEmail(newVal) {
//	            showError("Invalid email")
//	        }
//	    })
//	})
//
// Performance:
//
// UseDebounce creates one Ref and one Watch. The overhead is minimal (< 200ns).
// Timer operations are efficient and don't block the main goroutine.
//
// Thread Safety:
//
// UseDebounce is thread-safe. Multiple concurrent changes to the source value
// are handled correctly with proper mutex synchronization.
func UseDebounce[T any](ctx *bubbly.Context, value *bubbly.Ref[T], delay time.Duration) *bubbly.Ref[T] {
	// Create debounced ref with initial value from source
	debounced := bubbly.NewRef(value.GetTyped())

	// Timer for debouncing (protected by mutex for thread safety)
	var timer *time.Timer
	var timerMu sync.Mutex

	// Watch source value for changes
	cleanup := bubbly.Watch(value, func(newVal, _ T) {
		// Stop existing timer if any
		timerMu.Lock()
		if timer != nil {
			timer.Stop()
		}

		// Create new timer that updates debounced value after delay
		timer = time.AfterFunc(delay, func() {
			debounced.Set(newVal)
		})
		timerMu.Unlock()
	})

	// Register cleanup to stop timer on unmount (if context is available)
	if ctx != nil {
		ctx.OnUnmounted(func() {
			// Stop the watcher
			cleanup()

			// Stop any pending timer
			timerMu.Lock()
			if timer != nil {
				timer.Stop()
				timer = nil
			}
			timerMu.Unlock()
		})
	}

	return debounced
}
