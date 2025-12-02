package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// IntervalReturn is the return value of UseInterval.
// It provides periodic execution management with start/stop/toggle/reset controls.
//
// The interval uses an internal goroutine with time.Ticker for timing.
// The goroutine is properly cleaned up when Stop() is called or when the
// component unmounts (via OnUnmounted hook).
//
// Thread Safety:
// All methods are thread-safe and can be called concurrently.
// The IsRunning ref is updated atomically with the internal state.
type IntervalReturn struct {
	// IsRunning indicates if the interval is active.
	// This is a reactive ref that can be watched for changes.
	IsRunning *bubbly.Ref[bool]

	// callback is the function to execute on each tick
	callback func()

	// duration is the interval between ticks
	duration time.Duration

	// mu protects internal state
	mu sync.Mutex

	// ticker is the internal time.Ticker
	ticker *time.Ticker

	// stopChan signals the goroutine to stop
	stopChan chan struct{}

	// running tracks if goroutine is active (internal, not same as IsRunning ref)
	running bool
}

// Start begins the interval.
// If the interval is already running, this is a no-op.
// The callback will be executed after each duration interval.
//
// Example:
//
//	interval := UseInterval(ctx, func() {
//	    fmt.Println("tick")
//	}, time.Second)
//	interval.Start() // Starts printing "tick" every second
func (i *IntervalReturn) Start() {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.running {
		return // Already running
	}

	i.running = true
	i.IsRunning.Set(true)

	ticker := time.NewTicker(i.duration)
	stopChan := make(chan struct{})

	i.ticker = ticker
	i.stopChan = stopChan

	// Capture callback to avoid race on i.callback
	callback := i.callback

	go func() {
		for {
			select {
			case <-ticker.C:
				// Check if we're still running before executing callback
				// This prevents race condition where Stop() is called but
				// callback is already scheduled to execute
				i.mu.Lock()
				stillRunning := i.running
				i.mu.Unlock()

				if stillRunning {
					callback()
				}
			case <-stopChan:
				return
			}
		}
	}()
}

// Stop pauses the interval.
// If the interval is already stopped, this is a no-op.
// The callback will no longer be executed after Stop() returns.
//
// Example:
//
//	interval.Start()
//	// ... some time later ...
//	interval.Stop() // Stops the interval
func (i *IntervalReturn) Stop() {
	i.mu.Lock()
	defer i.mu.Unlock()

	if !i.running {
		return // Already stopped
	}

	// Set running to false first to prevent new callbacks
	i.running = false
	i.IsRunning.Set(false)

	// Close stopChan to signal goroutine to exit
	// Do this before stopping ticker to ensure goroutine exits cleanly
	if i.stopChan != nil {
		close(i.stopChan)
		i.stopChan = nil
	}

	// Now stop the ticker
	if i.ticker != nil {
		i.ticker.Stop()
		i.ticker = nil
	}
}

// Toggle starts if stopped, stops if running.
// This is a convenience method for toggling the interval state.
//
// Example:
//
//	interval.Toggle() // Starts if stopped
//	interval.Toggle() // Stops if running
func (i *IntervalReturn) Toggle() {
	if i.IsRunning.GetTyped() {
		i.Stop()
	} else {
		i.Start()
	}
}

// Reset stops and restarts the interval.
// This is useful for resetting the timing cycle.
// After Reset(), the next callback will execute after a full duration.
//
// Example:
//
//	interval.Start()
//	// ... some time later ...
//	interval.Reset() // Restarts the interval timing
func (i *IntervalReturn) Reset() {
	i.Stop()
	i.Start()
}

// UseInterval creates a periodic execution composable.
// It executes the callback function at regular intervals specified by duration.
//
// The interval starts in stopped state. Call Start() to begin execution.
// The callback is executed in a separate goroutine, so it should be thread-safe.
//
// This composable is useful for:
//   - Auto-refresh functionality (e.g., refresh data every 5 seconds)
//   - Polling for updates
//   - Periodic UI updates (e.g., clock display)
//   - Animation timing
//
// Parameters:
//   - ctx: The component context (required for lifecycle management)
//   - callback: The function to execute on each tick
//   - duration: The interval between ticks (must be positive)
//
// Returns:
//   - *IntervalReturn: A struct with IsRunning ref and control methods
//
// Panics:
//   - If duration is zero or negative
//
// Example - Auto-refresh:
//
//	Setup(func(ctx *bubbly.Context) {
//	    autoRefresh := composables.UseInterval(ctx, func() {
//	        fetchData()
//	    }, 5*time.Second)
//	    ctx.Expose("autoRefresh", autoRefresh)
//
//	    // Start auto-refresh
//	    autoRefresh.Start()
//	}).
//	WithKeyBinding("r", "toggleRefresh", "Toggle auto-refresh")
//
// Example - Clock update:
//
//	Setup(func(ctx *bubbly.Context) {
//	    currentTime := composables.UseState(ctx, time.Now())
//	    ctx.Expose("currentTime", currentTime)
//
//	    clock := composables.UseInterval(ctx, func() {
//	        currentTime.Set(time.Now())
//	    }, time.Second)
//
//	    clock.Start()
//	})
//
// Integration with CreateShared:
//
//	var UseSharedAutoRefresh = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.IntervalReturn {
//	        return composables.UseInterval(ctx, func() {
//	            refreshGlobalData()
//	        }, 10*time.Second)
//	    },
//	)
//
// Thread Safety:
//
// UseInterval is thread-safe. The callback is executed in a separate goroutine,
// so ensure the callback itself is thread-safe if it accesses shared state.
//
// Cleanup:
//
// The interval is automatically stopped when the component unmounts.
// You can also manually stop it by calling Stop().
func UseInterval(ctx *bubbly.Context, callback func(), duration time.Duration) *IntervalReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseInterval", time.Since(start))
	}()

	// Validate duration
	if duration <= 0 {
		panic("UseInterval: duration must be positive")
	}

	// Create return struct
	interval := &IntervalReturn{
		IsRunning: bubbly.NewRef(false),
		callback:  callback,
		duration:  duration,
	}

	// Register cleanup on unmount
	if ctx != nil {
		ctx.OnUnmounted(func() {
			interval.Stop()
		})
	}

	return interval
}
