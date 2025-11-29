package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// TimeoutReturn is the return value of UseTimeout.
// It provides delayed execution management with start/cancel/reset controls.
//
// The timeout uses an internal goroutine with time.AfterFunc for timing.
// The goroutine is properly cleaned up when Cancel() is called or when the
// component unmounts (via OnUnmounted hook).
//
// Thread Safety:
// All methods are thread-safe and can be called concurrently.
// The IsPending and IsExpired refs are updated atomically with the internal state.
type TimeoutReturn struct {
	// IsPending indicates if the timeout is waiting to fire.
	// This is a reactive ref that can be watched for changes.
	IsPending *bubbly.Ref[bool]

	// IsExpired indicates if the timeout has fired.
	// This is a reactive ref that can be watched for changes.
	IsExpired *bubbly.Ref[bool]

	// callback is the function to execute when timeout expires
	callback func()

	// duration is the delay before callback execution
	duration time.Duration

	// mu protects internal state
	mu sync.Mutex

	// timer is the internal timer
	timer *time.Timer

	// pending tracks if timer is active (internal, synced with IsPending ref)
	pending bool
}

// Start begins the timeout.
// If the timeout is already pending, this is a no-op.
// The callback will be executed after the duration elapses.
// If the timeout has already expired, Start() will reset IsExpired and start a new timeout.
//
// Example:
//
//	timeout := UseTimeout(ctx, func() {
//	    fmt.Println("timeout fired!")
//	}, 5*time.Second)
//	timeout.Start() // Will print "timeout fired!" after 5 seconds
func (t *TimeoutReturn) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.pending {
		return // Already pending
	}

	t.pending = true
	t.IsPending.Set(true)
	t.IsExpired.Set(false)

	// Capture callback to avoid race on t.callback
	callback := t.callback

	t.timer = time.AfterFunc(t.duration, func() {
		t.mu.Lock()
		// Check if still pending (might have been cancelled)
		if !t.pending {
			t.mu.Unlock()
			return
		}
		t.pending = false
		t.mu.Unlock()

		// Update refs outside lock to avoid deadlock with Watch
		t.IsPending.Set(false)
		t.IsExpired.Set(true)

		// Execute callback
		callback()
	})
}

// Cancel cancels the pending timeout.
// If the timeout is not pending, this is a no-op.
// The callback will not be executed after Cancel() is called.
// IsExpired remains unchanged (if it was already expired, it stays expired).
//
// Example:
//
//	timeout.Start()
//	// ... some time later ...
//	timeout.Cancel() // Cancels the timeout, callback won't fire
func (t *TimeoutReturn) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.pending {
		return // Not pending, nothing to cancel
	}

	t.pending = false
	t.IsPending.Set(false)

	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}

// Reset cancels any pending timeout and starts a new one.
// This is useful for implementing debounce-like behavior or
// restarting a timeout from the beginning.
//
// Example:
//
//	timeout.Start()
//	// ... user does something ...
//	timeout.Reset() // Restarts the timeout from the beginning
func (t *TimeoutReturn) Reset() {
	t.Cancel()
	t.Start()
}

// UseTimeout creates a delayed execution composable.
// It executes the callback function once after the specified duration.
//
// The timeout starts in stopped state. Call Start() to begin the countdown.
// The callback is executed in a separate goroutine, so it should be thread-safe.
//
// This composable is useful for:
//   - Delayed actions (e.g., auto-save after inactivity)
//   - Debouncing user input
//   - Timeout notifications
//   - Delayed UI transitions
//
// Parameters:
//   - ctx: The component context (required for lifecycle management)
//   - callback: The function to execute when timeout expires
//   - duration: The delay before execution (must be positive)
//
// Returns:
//   - *TimeoutReturn: A struct with IsPending/IsExpired refs and control methods
//
// Panics:
//   - If duration is zero or negative
//
// Example - Auto-save:
//
//	Setup(func(ctx *bubbly.Context) {
//	    autoSave := composables.UseTimeout(ctx, func() {
//	        saveDocument()
//	    }, 3*time.Second)
//	    ctx.Expose("autoSave", autoSave)
//
//	    // Start auto-save timer when content changes
//	    ctx.On("contentChanged", func(_ interface{}) {
//	        autoSave.Reset() // Reset timer on each change
//	    })
//	})
//
// Example - Notification dismiss:
//
//	Setup(func(ctx *bubbly.Context) {
//	    showNotification := composables.UseState(ctx, false)
//	    ctx.Expose("showNotification", showNotification)
//
//	    dismissTimer := composables.UseTimeout(ctx, func() {
//	        showNotification.Set(false)
//	    }, 5*time.Second)
//
//	    ctx.On("notify", func(_ interface{}) {
//	        showNotification.Set(true)
//	        dismissTimer.Reset()
//	    })
//	})
//
// Integration with CreateShared:
//
//	var UseSharedTimeout = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.TimeoutReturn {
//	        return composables.UseTimeout(ctx, func() {
//	            handleGlobalTimeout()
//	        }, 30*time.Second)
//	    },
//	)
//
// Thread Safety:
//
// UseTimeout is thread-safe. The callback is executed in a separate goroutine,
// so ensure the callback itself is thread-safe if it accesses shared state.
//
// Cleanup:
//
// The timeout is automatically cancelled when the component unmounts.
// You can also manually cancel it by calling Cancel().
func UseTimeout(ctx *bubbly.Context, callback func(), duration time.Duration) *TimeoutReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseTimeout", time.Since(start))
	}()

	// Validate duration
	if duration <= 0 {
		panic("UseTimeout: duration must be positive")
	}

	// Create return struct
	timeout := &TimeoutReturn{
		IsPending: bubbly.NewRef(false),
		IsExpired: bubbly.NewRef(false),
		callback:  callback,
		duration:  duration,
	}

	// Register cleanup on unmount
	if ctx != nil {
		ctx.OnUnmounted(func() {
			timeout.Cancel()
		})
	}

	return timeout
}
