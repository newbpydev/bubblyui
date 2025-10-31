package composables

import (
	"sync"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseEventListener registers an event handler with automatic cleanup.
// It provides a convenient way to listen to component events with proper
// lifecycle management, ensuring handlers are cleaned up when the component
// unmounts or when manually cleaned up.
//
// Unlike ctx.On() which requires EventHandler (func(interface{})), UseEventListener
// accepts a simpler handler signature (func()) for cases where the event data
// is not needed. This makes it more ergonomic for common use cases.
//
// The handler is automatically cleaned up when:
//   - The component unmounts (automatic cleanup via lifecycle)
//   - The returned cleanup function is called (manual cleanup)
//
// Cleanup is idempotent - calling the cleanup function multiple times is safe
// and will not cause panics or errors.
//
// Thread Safety:
//
// UseEventListener is thread-safe. The cleanup flag is protected by a mutex,
// allowing safe concurrent access from event emission, manual cleanup, and
// unmount cleanup.
//
// Parameters:
//   - ctx: The component context (required for lifecycle management)
//   - event: The event name to listen for
//   - handler: The function to execute when the event is emitted
//
// Returns:
//   - func(): A cleanup function that removes the event listener
//
// Example - Basic Usage:
//
//	Setup(func(ctx *Context) {
//	    handleClick := func() {
//	        fmt.Println("Button clicked!")
//	    }
//
//	    cleanup := UseEventListener(ctx, "click", handleClick)
//
//	    // Listener automatically cleaned up on unmount
//	    // Or manually: cleanup()
//	})
//
// Example - With Manual Cleanup:
//
//	Setup(func(ctx *Context) {
//	    count := ctx.Ref(0)
//
//	    cleanup := UseEventListener(ctx, "increment", func() {
//	        count.Set(count.GetTyped() + 1)
//	    })
//
//	    // Cleanup after 10 increments
//	    ctx.Watch(count, func(newVal, _ interface{}) {
//	        if newVal.(int) >= 10 {
//	            cleanup()
//	        }
//	    })
//	})
//
// Example - Multiple Listeners:
//
//	Setup(func(ctx *Context) {
//	    UseEventListener(ctx, "save", func() {
//	        saveData()
//	    })
//
//	    UseEventListener(ctx, "cancel", func() {
//	        resetForm()
//	    })
//
//	    UseEventListener(ctx, "delete", func() {
//	        deleteData()
//	    })
//	})
//
// Example - Conditional Listener:
//
//	Setup(func(ctx *Context) {
//	    isEnabled := ctx.Ref(true)
//	    var cleanup func()
//
//	    ctx.Watch(isEnabled, func(newVal, _ interface{}) {
//	        if newVal.(bool) {
//	            // Enable listener
//	            cleanup = UseEventListener(ctx, "action", func() {
//	                performAction()
//	            })
//	        } else {
//	            // Disable listener
//	            if cleanup != nil {
//	                cleanup()
//	            }
//	        }
//	    })
//	})
//
// Performance:
//
// UseEventListener has minimal overhead. It creates a closure with a mutex and
// boolean flag. The wrapper function adds negligible overhead to event handling.
//
// Integration with Component System:
//
// UseEventListener integrates seamlessly with the component event system. It
// uses ctx.On() internally and registers cleanup via ctx.OnUnmounted() to ensure
// proper lifecycle management.
func UseEventListener(ctx *bubbly.Context, event string, handler func()) func() {
	// Track whether the listener has been cleaned up
	var cleanedUp bool
	var mu sync.Mutex

	// Wrap the handler to check cleanup flag before executing
	wrappedHandler := func(data interface{}) {
		mu.Lock()
		defer mu.Unlock()

		// Don't execute if cleaned up
		if cleanedUp {
			return
		}

		// Execute the user's handler
		handler()
	}

	// Register the wrapped handler with the component
	if ctx != nil {
		ctx.On(event, wrappedHandler)
	}

	// Create cleanup function
	cleanup := func() {
		mu.Lock()
		defer mu.Unlock()
		cleanedUp = true
	}

	// Register automatic cleanup on unmount (if context is available)
	if ctx != nil {
		ctx.OnUnmounted(cleanup)
	}

	// Return manual cleanup function
	return cleanup
}
