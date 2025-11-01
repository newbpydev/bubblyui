package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseThrottle creates a throttled version of a function that limits how often it can execute.
// Unlike debouncing which waits for a quiet period, throttling ensures the function executes
// immediately on the first call, then prevents subsequent calls until the delay period has passed.
//
// This is useful for:
//   - Scroll event handlers (limit scroll processing rate)
//   - Button click prevention (prevent double-clicks)
//   - API rate limiting (enforce maximum request rate)
//   - Window resize handlers (limit layout recalculations)
//   - Mouse move tracking (reduce event processing overhead)
//
// The throttled function executes immediately on the first call. Subsequent calls within
// the delay period are ignored. Once the delay period passes, the next call will execute
// immediately and start a new throttle period.
//
// Cleanup is automatically registered with the component's lifecycle, ensuring no goroutine
// leaks when the component unmounts. The timer is stopped and any pending throttle state
// is cleared on unmount.
//
// Thread Safety:
//
// UseThrottle is thread-safe. Multiple concurrent calls to the throttled function are
// handled correctly with proper mutex synchronization. Only one execution will occur
// per throttle period, even under high concurrency.
//
// Parameters:
//   - ctx: The component context (required for lifecycle management)
//   - fn: The function to throttle
//   - delay: Minimum time between function executions
//
// Returns:
//   - func(): A throttled version of the input function
//
// Example - Scroll Handler:
//
//	Setup(func(ctx *Context) {
//	    handleScroll := func() {
//	        // Process scroll event
//	        updateScrollPosition()
//	    }
//
//	    throttledScroll := UseThrottle(ctx, handleScroll, 100*time.Millisecond)
//
//	    ctx.On("scroll", func(_ interface{}) {
//	        throttledScroll()  // Executes at most once per 100ms
//	    })
//	})
//
// Example - Button Click Prevention:
//
//	Setup(func(ctx *Context) {
//	    handleSubmit := func() {
//	        // Submit form
//	        submitForm()
//	    }
//
//	    throttledSubmit := UseThrottle(ctx, handleSubmit, 1*time.Second)
//
//	    ctx.On("submit", func(_ interface{}) {
//	        throttledSubmit()  // Prevents double-submission
//	    })
//	})
//
// Example - API Rate Limiting:
//
//	Setup(func(ctx *Context) {
//	    searchTerm := ctx.Ref("")
//
//	    performSearch := func() {
//	        if searchTerm.GetTyped() != "" {
//	            callSearchAPI(searchTerm.GetTyped())
//	        }
//	    }
//
//	    throttledSearch := UseThrottle(ctx, performSearch, 500*time.Millisecond)
//
//	    // Watch for changes, but throttle API calls
//	    ctx.Watch(searchTerm, func(newVal, _ string) {
//	        throttledSearch()  // Max 2 requests per second
//	    })
//	})
//
// Example - Mouse Move Tracking:
//
//	Setup(func(ctx *Context) {
//	    mouseX := ctx.Ref(0)
//	    mouseY := ctx.Ref(0)
//
//	    updatePosition := func() {
//	        // Update UI with current position
//	        renderCursor(mouseX.GetTyped(), mouseY.GetTyped())
//	    }
//
//	    throttledUpdate := UseThrottle(ctx, updatePosition, 16*time.Millisecond) // ~60fps
//
//	    ctx.On("mousemove", func(event interface{}) {
//	        // Update refs
//	        mouseX.Set(event.X)
//	        mouseY.Set(event.Y)
//	        throttledUpdate()  // Smooth rendering at 60fps
//	    })
//	})
//
// Performance:
//
// UseThrottle has minimal overhead (< 100ns). It creates a closure with a mutex and
// boolean flag. Timer operations are efficient and don't block the calling goroutine.
//
// Throttle vs Debounce:
//
//   - Throttle: Executes immediately, then limits rate (good for continuous events)
//   - Debounce: Waits for quiet period, executes once (good for sporadic events)
//
// Use throttle when you want regular execution during continuous activity (scrolling, resizing).
// Use debounce when you want to wait for activity to stop (search input, form validation).
func UseThrottle(ctx *bubbly.Context, fn func(), delay time.Duration) func() {
	// Track whether we're currently in a throttled state
	var isThrottled bool
	var mu sync.Mutex
	var timer *time.Timer

	// Return throttled function
	throttled := func() {
		mu.Lock()
		defer mu.Unlock()

		// If already throttled, ignore this call
		if isThrottled {
			return
		}

		// Execute the function immediately
		fn()

		// Set throttled state
		isThrottled = true

		// Start timer to reset throttle state after delay
		if delay > 0 {
			timer = time.AfterFunc(delay, func() {
				mu.Lock()
				isThrottled = false
				mu.Unlock()
			})
		} else {
			// Zero delay means no throttling
			isThrottled = false
		}
	}

	// Register cleanup to stop timer on unmount (if context is available)
	if ctx != nil {
		ctx.OnUnmounted(func() {
			mu.Lock()
			if timer != nil {
				timer.Stop()
				timer = nil
			}
			isThrottled = false
			mu.Unlock()
		})
	}

	return throttled
}
