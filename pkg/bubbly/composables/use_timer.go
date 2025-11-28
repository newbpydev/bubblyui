package composables

import (
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// defaultTickInterval is the default interval for timer updates.
const defaultTickInterval = 100 * time.Millisecond

// timerConfig holds configuration options for UseTimer.
type timerConfig struct {
	onExpire     func()
	tickInterval time.Duration
}

// TimerOption configures UseTimer.
type TimerOption func(*timerConfig)

// WithOnExpire sets a callback function to be executed when the timer expires.
// The callback is executed once when the timer reaches zero.
//
// Example:
//
//	timer := UseTimer(ctx, 60*time.Second,
//	    WithOnExpire(func() {
//	        fmt.Println("Timer expired!")
//	    }),
//	)
func WithOnExpire(fn func()) TimerOption {
	return func(c *timerConfig) {
		c.onExpire = fn
	}
}

// WithTickInterval sets the update frequency for the timer.
// The default tick interval is 100ms.
// Smaller intervals provide smoother progress updates but use more CPU.
//
// Example:
//
//	timer := UseTimer(ctx, 60*time.Second,
//	    WithTickInterval(50*time.Millisecond), // Update every 50ms
//	)
func WithTickInterval(d time.Duration) TimerOption {
	return func(c *timerConfig) {
		if d > 0 {
			c.tickInterval = d
		}
	}
}

// TimerReturn is the return value of UseTimer.
// It provides countdown timer functionality with progress tracking.
//
// The timer uses an internal goroutine with time.Ticker for timing.
// The goroutine is properly cleaned up when Stop() is called or when the
// component unmounts (via OnUnmounted hook).
//
// Thread Safety:
// All methods are thread-safe and can be called concurrently.
// The Remaining and IsRunning refs are updated atomically with the internal state.
type TimerReturn struct {
	// Remaining is the remaining time until expiry.
	// This is a reactive ref that can be watched for changes.
	Remaining *bubbly.Ref[time.Duration]

	// IsRunning indicates if the timer is actively counting down.
	// This is a reactive ref that can be watched for changes.
	IsRunning *bubbly.Ref[bool]

	// IsExpired indicates if the timer has reached zero.
	// This is a computed value that reacts to Remaining changes.
	IsExpired *bubbly.Computed[bool]

	// Progress is the completion percentage (0.0 to 1.0).
	// This is a computed value that reacts to Remaining changes.
	Progress *bubbly.Computed[float64]

	// initialDuration is the original duration for Reset()
	initialDuration time.Duration

	// tickInterval is the update frequency
	tickInterval time.Duration

	// onExpire is the callback to execute when timer expires
	onExpire func()

	// mu protects internal state
	mu sync.Mutex

	// ticker is the internal time.Ticker
	ticker *time.Ticker

	// stopChan signals the goroutine to stop
	stopChan chan struct{}

	// running tracks if goroutine is active (internal, synced with IsRunning ref)
	running bool

	// expired tracks if timer has expired (to prevent multiple onExpire calls)
	expired bool
}

// Start begins the countdown.
// If the timer is already running, this is a no-op.
// The timer will count down from the current Remaining value.
//
// Example:
//
//	timer := UseTimer(ctx, 60*time.Second)
//	timer.Start() // Starts counting down from 60 seconds
func (t *TimerReturn) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return // Already running
	}

	t.running = true
	t.IsRunning.Set(true)

	ticker := time.NewTicker(t.tickInterval)
	stopChan := make(chan struct{})

	t.ticker = ticker
	t.stopChan = stopChan

	// Capture callback to avoid race on t.onExpire
	onExpire := t.onExpire

	go func() {
		for {
			select {
			case <-ticker.C:
				t.mu.Lock()
				if !t.running {
					t.mu.Unlock()
					return
				}

				// Decrease remaining
				remaining := t.Remaining.GetTyped()
				newRemaining := remaining - t.tickInterval
				if newRemaining < 0 {
					newRemaining = 0
				}
				t.Remaining.Set(newRemaining)

				// Check if expired
				if newRemaining <= 0 && !t.expired {
					t.expired = true
					t.running = false
					t.mu.Unlock()

					// Update refs outside lock
					t.IsRunning.Set(false)

					// Stop ticker
					ticker.Stop()

					// Execute callback
					if onExpire != nil {
						onExpire()
					}
					return
				}
				t.mu.Unlock()

			case <-stopChan:
				return
			}
		}
	}()
}

// Stop pauses the countdown.
// If the timer is already stopped, this is a no-op.
// The current Remaining value is preserved, allowing resume with Start().
//
// Example:
//
//	timer.Start()
//	// ... some time later ...
//	timer.Stop() // Pauses the countdown
//	// ... later ...
//	timer.Start() // Resumes from where it stopped
func (t *TimerReturn) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return // Already stopped
	}

	t.running = false
	t.IsRunning.Set(false)

	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}

	if t.stopChan != nil {
		close(t.stopChan)
		t.stopChan = nil
	}
}

// Reset stops the timer and restarts from the full initial duration.
// This resets the Remaining value to the original duration and clears the expired state.
// The timer is left in stopped state after Reset().
//
// Example:
//
//	timer := UseTimer(ctx, 60*time.Second)
//	timer.Start()
//	// ... some time later ...
//	timer.Reset() // Resets to 60 seconds, timer is stopped
//	timer.Start() // Start again from 60 seconds
func (t *TimerReturn) Reset() {
	t.Stop()

	t.mu.Lock()
	t.expired = false
	t.mu.Unlock()

	t.Remaining.Set(t.initialDuration)
}

// UseTimer creates a countdown timer composable with progress tracking.
// It counts down from the specified duration, updating Remaining on each tick.
//
// The timer starts in stopped state. Call Start() to begin the countdown.
// The timer automatically stops when it reaches zero and calls the OnExpire callback.
//
// This composable is useful for:
//   - Countdown timers (e.g., game timers, session timeouts)
//   - Progress indicators (e.g., loading bars, time-limited actions)
//   - Timed events (e.g., auto-dismiss notifications)
//   - Animation timing
//
// Parameters:
//   - ctx: The component context (required for lifecycle management)
//   - duration: The initial countdown duration (must be positive)
//   - opts: Optional configuration (WithOnExpire, WithTickInterval)
//
// Returns:
//   - *TimerReturn: A struct with Remaining, IsRunning, IsExpired, Progress and control methods
//
// Panics:
//   - If duration is zero or negative
//
// Example - Countdown timer:
//
//	Setup(func(ctx *bubbly.Context) {
//	    timer := composables.UseTimer(ctx, 60*time.Second,
//	        composables.WithOnExpire(func() {
//	            ctx.Emit("timerExpired", nil)
//	        }),
//	    )
//	    ctx.Expose("timer", timer)
//
//	    ctx.On("startTimer", func(_ interface{}) {
//	        timer.Start()
//	    })
//	})
//
// Example - Progress bar:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    timer := ctx.Get("timer").(*composables.TimerReturn)
//	    progress := timer.Progress.Get().(float64)
//	    remaining := timer.Remaining.GetTyped()
//
//	    return fmt.Sprintf("Progress: %.0f%% | Remaining: %v", progress*100, remaining)
//	})
//
// Integration with CreateShared:
//
//	var UseSharedTimer = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.TimerReturn {
//	        return composables.UseTimer(ctx, 5*time.Minute,
//	            composables.WithOnExpire(func() {
//	                handleGlobalTimeout()
//	            }),
//	        )
//	    },
//	)
//
// Thread Safety:
//
// UseTimer is thread-safe. The countdown runs in a separate goroutine,
// and all state updates are synchronized.
//
// Cleanup:
//
// The timer is automatically stopped when the component unmounts.
// You can also manually stop it by calling Stop().
func UseTimer(ctx *bubbly.Context, duration time.Duration, opts ...TimerOption) *TimerReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseTimer", time.Since(start))
	}()

	// Validate duration
	if duration <= 0 {
		panic("UseTimer: duration must be positive")
	}

	// Apply options
	config := &timerConfig{
		tickInterval: defaultTickInterval,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Create Remaining ref
	remainingRef := bubbly.NewRef(duration)

	// Create timer struct
	timer := &TimerReturn{
		Remaining:       remainingRef,
		IsRunning:       bubbly.NewRef(false),
		initialDuration: duration,
		tickInterval:    config.tickInterval,
		onExpire:        config.onExpire,
	}

	// Create IsExpired computed
	timer.IsExpired = bubbly.NewComputed(func() bool {
		return remainingRef.GetTyped() <= 0
	})

	// Create Progress computed
	timer.Progress = bubbly.NewComputed(func() float64 {
		remaining := remainingRef.GetTyped()
		if duration <= 0 {
			return 1.0 // Avoid division by zero
		}
		progress := 1.0 - float64(remaining)/float64(duration)
		if progress < 0 {
			return 0.0
		}
		if progress > 1.0 {
			return 1.0
		}
		return progress
	})

	// Register cleanup on unmount
	if ctx != nil {
		ctx.OnUnmounted(func() {
			timer.Stop()
		})
	}

	return timer
}
