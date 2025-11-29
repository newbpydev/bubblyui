package composables

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseTimer_InitialRemainingEqualsDuration verifies that initial remaining equals the duration.
func TestUseTimer_InitialRemainingEqualsDuration(t *testing.T) {
	ctx := createTestContext()
	duration := 5 * time.Second

	timer := UseTimer(ctx, duration)

	assert.Equal(t, duration, timer.Remaining.GetTyped(), "initial remaining should equal duration")
	assert.False(t, timer.IsRunning.GetTyped(), "timer should not be running initially")
	assert.False(t, timer.IsExpired.Get().(bool), "timer should not be expired initially")
	assert.Equal(t, 0.0, timer.Progress.Get().(float64), "progress should be 0.0 initially")
}

// TestUseTimer_StartBeginsCountdown verifies that Start() begins the countdown.
func TestUseTimer_StartBeginsCountdown(t *testing.T) {
	ctx := createTestContext()
	duration := 200 * time.Millisecond

	timer := UseTimer(ctx, duration, WithTickInterval(50*time.Millisecond))

	timer.Start()
	assert.True(t, timer.IsRunning.GetTyped(), "timer should be running after Start()")

	// Wait for a few ticks
	time.Sleep(120 * time.Millisecond)

	remaining := timer.Remaining.GetTyped()
	assert.Less(t, remaining, duration, "remaining should decrease after Start()")
	assert.Greater(t, remaining, time.Duration(0), "remaining should still be positive")

	// Cleanup
	timer.Stop()
}

// TestUseTimer_StopPausesCountdown verifies that Stop() pauses the countdown.
func TestUseTimer_StopPausesCountdown(t *testing.T) {
	ctx := createTestContext()
	duration := 500 * time.Millisecond

	timer := UseTimer(ctx, duration, WithTickInterval(50*time.Millisecond))

	timer.Start()
	time.Sleep(100 * time.Millisecond)

	timer.Stop()
	remainingAtStop := timer.Remaining.GetTyped()
	assert.False(t, timer.IsRunning.GetTyped(), "timer should not be running after Stop()")

	// Wait more and verify remaining hasn't changed
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, remainingAtStop, timer.Remaining.GetTyped(), "remaining should not change after Stop()")
}

// TestUseTimer_ResetRestartsFromFullDuration verifies that Reset() restarts from full duration.
func TestUseTimer_ResetRestartsFromFullDuration(t *testing.T) {
	ctx := createTestContext()
	duration := 300 * time.Millisecond

	timer := UseTimer(ctx, duration, WithTickInterval(50*time.Millisecond))

	timer.Start()
	time.Sleep(100 * time.Millisecond)

	// Remaining should have decreased
	assert.Less(t, timer.Remaining.GetTyped(), duration)

	timer.Reset()

	assert.Equal(t, duration, timer.Remaining.GetTyped(), "remaining should be reset to full duration")
	assert.False(t, timer.IsRunning.GetTyped(), "timer should not be running after Reset()")
	assert.False(t, timer.IsExpired.Get().(bool), "timer should not be expired after Reset()")
	assert.Equal(t, 0.0, timer.Progress.Get().(float64), "progress should be 0.0 after Reset()")
}

// TestUseTimer_IsExpiredTrueWhenRemainingZero verifies that IsExpired is true when remaining <= 0.
func TestUseTimer_IsExpiredTrueWhenRemainingZero(t *testing.T) {
	ctx := createTestContext()
	duration := 80 * time.Millisecond

	timer := UseTimer(ctx, duration, WithTickInterval(20*time.Millisecond))

	assert.False(t, timer.IsExpired.Get().(bool), "should not be expired initially")

	timer.Start()
	time.Sleep(150 * time.Millisecond) // Wait for timer to expire

	assert.True(t, timer.IsExpired.Get().(bool), "should be expired after duration")
	assert.LessOrEqual(t, timer.Remaining.GetTyped(), time.Duration(0), "remaining should be <= 0")
	assert.False(t, timer.IsRunning.GetTyped(), "timer should stop when expired")
}

// TestUseTimer_ProgressCalculatedCorrectly verifies that progress is calculated correctly.
func TestUseTimer_ProgressCalculatedCorrectly(t *testing.T) {
	tests := []struct {
		name            string
		duration        time.Duration
		waitTime        time.Duration
		tickInterval    time.Duration
		expectedMinProg float64
		expectedMaxProg float64
	}{
		{
			name:            "0% progress at start",
			duration:        200 * time.Millisecond,
			waitTime:        0,
			tickInterval:    20 * time.Millisecond,
			expectedMinProg: 0.0,
			expectedMaxProg: 0.0,
		},
		{
			name:            "~50% progress at half duration",
			duration:        200 * time.Millisecond,
			waitTime:        100 * time.Millisecond,
			tickInterval:    20 * time.Millisecond,
			expectedMinProg: 0.4,
			expectedMaxProg: 0.7,
		},
		{
			name:            "100% progress at expiry",
			duration:        80 * time.Millisecond,
			waitTime:        150 * time.Millisecond,
			tickInterval:    20 * time.Millisecond,
			expectedMinProg: 1.0,
			expectedMaxProg: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			timer := UseTimer(ctx, tt.duration, WithTickInterval(tt.tickInterval))

			if tt.waitTime > 0 {
				timer.Start()
				time.Sleep(tt.waitTime)
				timer.Stop()
			}

			progress := timer.Progress.Get().(float64)
			assert.GreaterOrEqual(t, progress, tt.expectedMinProg, "progress should be >= expected min")
			assert.LessOrEqual(t, progress, tt.expectedMaxProg, "progress should be <= expected max")
		})
	}
}

// TestUseTimer_OnExpireCallbackExecuted verifies that OnExpire callback is executed.
func TestUseTimer_OnExpireCallbackExecuted(t *testing.T) {
	ctx := createTestContext()
	var called int32

	timer := UseTimer(ctx, 60*time.Millisecond,
		WithTickInterval(20*time.Millisecond),
		WithOnExpire(func() {
			atomic.AddInt32(&called, 1)
		}),
	)

	timer.Start()
	time.Sleep(120 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called), "OnExpire callback should be called once")
	assert.True(t, timer.IsExpired.Get().(bool), "timer should be expired")
}

// TestUseTimer_TickIntervalConfigurable verifies that tick interval is configurable.
func TestUseTimer_TickIntervalConfigurable(t *testing.T) {
	ctx := createTestContext()
	duration := 200 * time.Millisecond

	// With fast tick interval
	timerFast := UseTimer(ctx, duration, WithTickInterval(10*time.Millisecond))
	timerFast.Start()
	time.Sleep(50 * time.Millisecond)
	timerFast.Stop()
	remainingFast := timerFast.Remaining.GetTyped()

	// With slow tick interval
	ctx2 := createTestContext()
	timerSlow := UseTimer(ctx2, duration, WithTickInterval(100*time.Millisecond))
	timerSlow.Start()
	time.Sleep(50 * time.Millisecond)
	timerSlow.Stop()
	remainingSlow := timerSlow.Remaining.GetTyped()

	// Fast ticker should have more precise remaining (closer to actual elapsed)
	// Slow ticker might not have ticked yet
	assert.Less(t, remainingFast, remainingSlow, "fast ticker should show more elapsed time")
}

// TestUseTimer_NegativeDuration_Panics verifies that negative duration causes panic.
func TestUseTimer_NegativeDuration_Panics(t *testing.T) {
	ctx := createTestContext()

	assert.Panics(t, func() {
		UseTimer(ctx, -1*time.Second)
	}, "UseTimer should panic with negative duration")
}

// TestUseTimer_ZeroDuration_Panics verifies that zero duration causes panic.
func TestUseTimer_ZeroDuration_Panics(t *testing.T) {
	ctx := createTestContext()

	assert.Panics(t, func() {
		UseTimer(ctx, 0)
	}, "UseTimer should panic with zero duration")
}

// TestUseTimer_CleanupOnUnmount verifies that the timer is stopped on component unmount.
func TestUseTimer_CleanupOnUnmount(t *testing.T) {
	var timer *TimerReturn
	var called int32

	// Create a component with proper lifecycle
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			timer = UseTimer(ctx, 100*time.Millisecond,
				WithTickInterval(20*time.Millisecond),
				WithOnExpire(func() {
					atomic.AddInt32(&called, 1)
				}),
			)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return ""
		}).
		Build()
	require.NoError(t, err)

	// Initialize the component (triggers Setup)
	component.Init()

	// Start the timer
	timer.Start()
	assert.True(t, timer.IsRunning.GetTyped())

	// Unmount the component (should trigger OnUnmounted and stop timer)
	time.Sleep(30 * time.Millisecond)
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Wait past original duration
	time.Sleep(150 * time.Millisecond)

	assert.Equal(t, int32(0), atomic.LoadInt32(&called), "OnExpire should not be called after unmount")
	assert.False(t, timer.IsRunning.GetTyped(), "timer should not be running after unmount")
}

// TestUseTimer_ConcurrentAccess verifies thread safety with concurrent access.
func TestUseTimer_ConcurrentAccess(t *testing.T) {
	ctx := createTestContext()
	timer := UseTimer(ctx, 500*time.Millisecond, WithTickInterval(20*time.Millisecond))

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			switch n % 4 {
			case 0:
				timer.Start()
			case 1:
				timer.Stop()
			case 2:
				timer.Reset()
			case 3:
				_ = timer.Remaining.GetTyped()
				_ = timer.IsRunning.GetTyped()
				_ = timer.IsExpired.Get()
				_ = timer.Progress.Get()
			}
		}(i)
	}

	wg.Wait()

	// Should not panic and should be in a consistent state
	_ = timer.IsRunning.GetTyped()
	_ = timer.Remaining.GetTyped()

	// Cleanup
	timer.Stop()
}

// TestUseTimer_WorksWithCreateShared verifies integration with CreateShared pattern.
func TestUseTimer_WorksWithCreateShared(t *testing.T) {
	var sharedTimer *TimerReturn
	var called int32

	UseSharedTimer := CreateShared(func(ctx *bubbly.Context) *TimerReturn {
		return UseTimer(ctx, 60*time.Millisecond,
			WithTickInterval(20*time.Millisecond),
			WithOnExpire(func() {
				atomic.AddInt32(&called, 1)
			}),
		)
	})

	// First call creates the timer
	ctx1 := createTestContext()
	sharedTimer = UseSharedTimer(ctx1)
	require.NotNil(t, sharedTimer)

	// Second call returns the same instance
	ctx2 := createTestContext()
	sameTimer := UseSharedTimer(ctx2)
	assert.Same(t, sharedTimer, sameTimer, "CreateShared should return same instance")

	// Verify it works
	sharedTimer.Start()
	time.Sleep(120 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called), "shared timer OnExpire should have been called")
	assert.True(t, sharedTimer.IsExpired.Get().(bool))
}

// TestUseTimer_NilContext verifies that nil context is handled gracefully.
func TestUseTimer_NilContext(t *testing.T) {
	var called int32

	// Should not panic with nil context
	timer := UseTimer(nil, 60*time.Millisecond,
		WithTickInterval(20*time.Millisecond),
		WithOnExpire(func() {
			atomic.AddInt32(&called, 1)
		}),
	)

	require.NotNil(t, timer)
	assert.False(t, timer.IsRunning.GetTyped())

	timer.Start()
	time.Sleep(120 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	assert.True(t, timer.IsExpired.Get().(bool))
}

// TestUseTimer_MultipleStartsNoOp verifies that multiple Start() calls are no-op.
func TestUseTimer_MultipleStartsNoOp(t *testing.T) {
	ctx := createTestContext()
	timer := UseTimer(ctx, 200*time.Millisecond, WithTickInterval(50*time.Millisecond))

	// Multiple starts should not create multiple tickers
	timer.Start()
	timer.Start()
	timer.Start()

	assert.True(t, timer.IsRunning.GetTyped())

	time.Sleep(80 * time.Millisecond)
	remaining := timer.Remaining.GetTyped()

	// Should have decreased by approximately 50-100ms worth
	assert.Less(t, remaining, 200*time.Millisecond)
	assert.Greater(t, remaining, 100*time.Millisecond)

	timer.Stop()
}

// TestUseTimer_MultipleStopsNoOp verifies that multiple Stop() calls are no-op.
func TestUseTimer_MultipleStopsNoOp(t *testing.T) {
	ctx := createTestContext()
	timer := UseTimer(ctx, 200*time.Millisecond, WithTickInterval(50*time.Millisecond))

	// Multiple stops should not panic
	assert.NotPanics(t, func() {
		timer.Stop()
		timer.Stop()
		timer.Stop()
	}, "multiple Stop() calls should not panic")

	assert.False(t, timer.IsRunning.GetTyped())
}

// TestUseTimer_IsRunningReactive verifies that IsRunning ref is reactive.
func TestUseTimer_IsRunningReactive(t *testing.T) {
	ctx := createTestContext()
	timer := UseTimer(ctx, 100*time.Millisecond, WithTickInterval(20*time.Millisecond))

	// Track changes with Watch
	var changes []bool
	var mu sync.Mutex

	bubbly.Watch(timer.IsRunning, func(newVal, oldVal bool) {
		mu.Lock()
		changes = append(changes, newVal)
		mu.Unlock()
	})

	timer.Start()
	time.Sleep(150 * time.Millisecond) // Wait for expiry

	// Give Watch time to process
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []bool{true, false}, changes, "IsRunning should be reactive (true on start, false on expiry)")
}

// TestUseTimer_RemainingReactive verifies that Remaining ref is reactive.
func TestUseTimer_RemainingReactive(t *testing.T) {
	ctx := createTestContext()
	timer := UseTimer(ctx, 100*time.Millisecond, WithTickInterval(20*time.Millisecond))

	// Track changes with Watch
	var changeCount int32

	bubbly.Watch(timer.Remaining, func(newVal, oldVal time.Duration) {
		atomic.AddInt32(&changeCount, 1)
	})

	timer.Start()
	time.Sleep(80 * time.Millisecond)
	timer.Stop()

	// Give Watch time to process
	time.Sleep(20 * time.Millisecond)

	count := atomic.LoadInt32(&changeCount)
	assert.GreaterOrEqual(t, count, int32(2), "Remaining should have changed multiple times during countdown")
}

// TestUseTimer_ResumeAfterStop verifies that Start() after Stop() resumes from current remaining.
func TestUseTimer_ResumeAfterStop(t *testing.T) {
	ctx := createTestContext()
	timer := UseTimer(ctx, 200*time.Millisecond, WithTickInterval(20*time.Millisecond))

	timer.Start()
	time.Sleep(60 * time.Millisecond)
	timer.Stop()

	remainingAtStop := timer.Remaining.GetTyped()
	assert.Less(t, remainingAtStop, 200*time.Millisecond, "remaining should have decreased")

	// Resume
	timer.Start()
	time.Sleep(60 * time.Millisecond)
	timer.Stop()

	remainingAfterResume := timer.Remaining.GetTyped()
	assert.Less(t, remainingAfterResume, remainingAtStop, "remaining should continue decreasing after resume")
}

// TestUseTimer_DefaultTickInterval verifies the default tick interval is 100ms.
func TestUseTimer_DefaultTickInterval(t *testing.T) {
	ctx := createTestContext()
	timer := UseTimer(ctx, 500*time.Millisecond) // No WithTickInterval

	timer.Start()
	time.Sleep(50 * time.Millisecond)

	// With 100ms default tick, no tick should have happened yet
	remaining := timer.Remaining.GetTyped()
	assert.Equal(t, 500*time.Millisecond, remaining, "remaining should not have changed before first tick")

	time.Sleep(80 * time.Millisecond) // Total ~130ms, one tick should have happened

	remaining = timer.Remaining.GetTyped()
	assert.Less(t, remaining, 500*time.Millisecond, "remaining should decrease after first tick")

	timer.Stop()
}
