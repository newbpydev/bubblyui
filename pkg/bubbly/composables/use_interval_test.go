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

// TestUseInterval_StartsInStoppedState verifies that UseInterval starts in stopped state.
func TestUseInterval_StartsInStoppedState(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	interval := UseInterval(ctx, callback, 100*time.Millisecond)

	assert.False(t, interval.IsRunning.GetTyped(), "interval should start in stopped state")
}

// TestUseInterval_Start_BeginsInterval verifies that Start() begins the interval.
func TestUseInterval_Start_BeginsInterval(t *testing.T) {
	ctx := createTestContext()
	var counter int32

	interval := UseInterval(ctx, func() {
		atomic.AddInt32(&counter, 1)
	}, 10*time.Millisecond)

	interval.Start()
	assert.True(t, interval.IsRunning.GetTyped(), "interval should be running after Start()")

	// Wait for a few ticks
	time.Sleep(50 * time.Millisecond)

	interval.Stop()
	assert.Greater(t, atomic.LoadInt32(&counter), int32(0), "callback should have been called")
}

// TestUseInterval_Stop_PausesInterval verifies that Stop() pauses the interval.
func TestUseInterval_Stop_PausesInterval(t *testing.T) {
	ctx := createTestContext()
	var counter int32

	interval := UseInterval(ctx, func() {
		atomic.AddInt32(&counter, 1)
	}, 10*time.Millisecond)

	interval.Start()
	time.Sleep(30 * time.Millisecond)
	interval.Stop()

	assert.False(t, interval.IsRunning.GetTyped(), "interval should be stopped after Stop()")

	// Wait a bit for any in-flight callback to complete
	time.Sleep(15 * time.Millisecond)

	// Record counter value after stop has fully taken effect
	countAfterStop := atomic.LoadInt32(&counter)

	// Wait and verify no more ticks
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, countAfterStop, atomic.LoadInt32(&counter), "callback should not be called after Stop()")
}

// TestUseInterval_Toggle_FlipsState verifies that Toggle() flips the running state.
func TestUseInterval_Toggle_FlipsState(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	interval := UseInterval(ctx, callback, 100*time.Millisecond)

	// Initially stopped
	assert.False(t, interval.IsRunning.GetTyped())

	// Toggle to running
	interval.Toggle()
	assert.True(t, interval.IsRunning.GetTyped())

	// Toggle back to stopped
	interval.Toggle()
	assert.False(t, interval.IsRunning.GetTyped())

	// Toggle to running again
	interval.Toggle()
	assert.True(t, interval.IsRunning.GetTyped())

	// Cleanup
	interval.Stop()
}

// TestUseInterval_Reset_Restarts verifies that Reset() stops and restarts the interval.
func TestUseInterval_Reset_Restarts(t *testing.T) {
	ctx := createTestContext()
	var counter int32

	interval := UseInterval(ctx, func() {
		atomic.AddInt32(&counter, 1)
	}, 10*time.Millisecond)

	interval.Start()
	time.Sleep(30 * time.Millisecond)

	// Reset should restart
	interval.Reset()
	assert.True(t, interval.IsRunning.GetTyped(), "interval should be running after Reset()")

	// Verify callback continues
	countBeforeWait := atomic.LoadInt32(&counter)
	time.Sleep(30 * time.Millisecond)
	assert.Greater(t, atomic.LoadInt32(&counter), countBeforeWait, "callback should continue after Reset()")

	interval.Stop()
}

// TestUseInterval_CallbackExecutedOnTick verifies that the callback is executed on each tick.
func TestUseInterval_CallbackExecutedOnTick(t *testing.T) {
	ctx := createTestContext()
	var counter int32

	interval := UseInterval(ctx, func() {
		atomic.AddInt32(&counter, 1)
	}, 10*time.Millisecond)

	interval.Start()
	time.Sleep(55 * time.Millisecond)
	interval.Stop()

	// Should have been called approximately 5 times (50ms / 10ms)
	count := atomic.LoadInt32(&counter)
	assert.GreaterOrEqual(t, count, int32(3), "callback should have been called multiple times")
	assert.LessOrEqual(t, count, int32(7), "callback count should be reasonable")
}

// TestUseInterval_CleanupOnUnmount verifies that the interval is stopped on component unmount.
func TestUseInterval_CleanupOnUnmount(t *testing.T) {
	var counter int32
	var interval *IntervalReturn

	// Create a component with proper lifecycle
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			interval = UseInterval(ctx, func() {
				atomic.AddInt32(&counter, 1)
			}, 10*time.Millisecond)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return ""
		}).
		Build()
	require.NoError(t, err)

	// Initialize the component (triggers Setup)
	component.Init()

	// Start the interval
	interval.Start()
	time.Sleep(30 * time.Millisecond)

	// Verify it's running
	countBeforeUnmount := atomic.LoadInt32(&counter)
	assert.Greater(t, countBeforeUnmount, int32(0), "callback should have been called")

	// Unmount the component (should trigger OnUnmounted and stop interval)
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	countAfterUnmount := atomic.LoadInt32(&counter)
	time.Sleep(30 * time.Millisecond)

	assert.Equal(t, countAfterUnmount, atomic.LoadInt32(&counter), "callback should not be called after unmount")
	assert.False(t, interval.IsRunning.GetTyped(), "interval should be stopped after unmount")
}

// TestUseInterval_NegativeDuration_Panics verifies that negative duration causes panic.
func TestUseInterval_NegativeDuration_Panics(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	assert.Panics(t, func() {
		UseInterval(ctx, callback, -1*time.Millisecond)
	}, "UseInterval should panic with negative duration")
}

// TestUseInterval_ZeroDuration_Panics verifies that zero duration causes panic.
func TestUseInterval_ZeroDuration_Panics(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	assert.Panics(t, func() {
		UseInterval(ctx, callback, 0)
	}, "UseInterval should panic with zero duration")
}

// TestUseInterval_MultipleStarts_NoOp verifies that multiple Start() calls are no-op.
func TestUseInterval_MultipleStarts_NoOp(t *testing.T) {
	ctx := createTestContext()
	var counter int32

	interval := UseInterval(ctx, func() {
		atomic.AddInt32(&counter, 1)
	}, 10*time.Millisecond)

	// Multiple starts should not create multiple goroutines
	interval.Start()
	interval.Start()
	interval.Start()

	assert.True(t, interval.IsRunning.GetTyped())

	time.Sleep(35 * time.Millisecond)
	interval.Stop()

	// Count should be reasonable (not 3x what it should be)
	count := atomic.LoadInt32(&counter)
	assert.GreaterOrEqual(t, count, int32(1))
	assert.LessOrEqual(t, count, int32(5), "should not have multiple goroutines running")
}

// TestUseInterval_MultipleStops_NoOp verifies that multiple Stop() calls are no-op.
func TestUseInterval_MultipleStops_NoOp(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	interval := UseInterval(ctx, callback, 100*time.Millisecond)

	// Multiple stops should not panic
	assert.NotPanics(t, func() {
		interval.Stop()
		interval.Stop()
		interval.Stop()
	}, "multiple Stop() calls should not panic")

	assert.False(t, interval.IsRunning.GetTyped())
}

// TestUseInterval_ConcurrentStartStop verifies thread safety with concurrent Start/Stop calls.
func TestUseInterval_ConcurrentStartStop(t *testing.T) {
	ctx := createTestContext()
	var counter int32

	interval := UseInterval(ctx, func() {
		atomic.AddInt32(&counter, 1)
	}, 5*time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				interval.Start()
			} else {
				interval.Stop()
			}
		}(i)
	}

	wg.Wait()

	// Should not panic and should be in a consistent state
	// Final state depends on timing, but should be valid
	_ = interval.IsRunning.GetTyped()

	// Cleanup
	interval.Stop()
}

// TestUseInterval_WorksWithCreateShared verifies integration with CreateShared pattern.
func TestUseInterval_WorksWithCreateShared(t *testing.T) {
	var sharedInterval *IntervalReturn
	var counter int32

	UseSharedInterval := CreateShared(func(ctx *bubbly.Context) *IntervalReturn {
		return UseInterval(ctx, func() {
			atomic.AddInt32(&counter, 1)
		}, 10*time.Millisecond)
	})

	// First call creates the interval
	ctx1 := createTestContext()
	sharedInterval = UseSharedInterval(ctx1)
	require.NotNil(t, sharedInterval)

	// Second call returns the same instance
	ctx2 := createTestContext()
	sameInterval := UseSharedInterval(ctx2)
	assert.Same(t, sharedInterval, sameInterval, "CreateShared should return same instance")

	// Verify it works
	sharedInterval.Start()
	time.Sleep(30 * time.Millisecond)
	sharedInterval.Stop()

	assert.Greater(t, atomic.LoadInt32(&counter), int32(0), "shared interval callback should have been called")
}

// TestUseInterval_NilContext verifies that nil context is handled gracefully.
func TestUseInterval_NilContext(t *testing.T) {
	var counter int32

	// Should not panic with nil context
	interval := UseInterval(nil, func() {
		atomic.AddInt32(&counter, 1)
	}, 10*time.Millisecond)

	require.NotNil(t, interval)
	assert.False(t, interval.IsRunning.GetTyped())

	interval.Start()
	time.Sleep(30 * time.Millisecond)
	interval.Stop()

	assert.Greater(t, atomic.LoadInt32(&counter), int32(0))
}

// TestUseInterval_StartStopStartSequence verifies start-stop-start sequence works correctly.
func TestUseInterval_StartStopStartSequence(t *testing.T) {
	ctx := createTestContext()
	var counter int32

	interval := UseInterval(ctx, func() {
		atomic.AddInt32(&counter, 1)
	}, 10*time.Millisecond)

	// Start
	interval.Start()
	time.Sleep(25 * time.Millisecond)
	countAfterFirstRun := atomic.LoadInt32(&counter)

	// Stop
	interval.Stop()
	time.Sleep(25 * time.Millisecond)
	countAfterStop := atomic.LoadInt32(&counter)
	assert.Equal(t, countAfterFirstRun, countAfterStop, "counter should not change while stopped")

	// Start again
	interval.Start()
	time.Sleep(25 * time.Millisecond)
	interval.Stop()

	assert.Greater(t, atomic.LoadInt32(&counter), countAfterStop, "counter should increase after restart")
}

// TestUseInterval_IsRunningReactive verifies that IsRunning ref is reactive.
func TestUseInterval_IsRunningReactive(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	interval := UseInterval(ctx, callback, 100*time.Millisecond)

	// Track changes with Watch
	var changes []bool
	var mu sync.Mutex

	bubbly.Watch(interval.IsRunning, func(newVal, oldVal bool) {
		mu.Lock()
		changes = append(changes, newVal)
		mu.Unlock()
	})

	interval.Start()
	interval.Stop()
	interval.Start()
	interval.Stop()

	// Give Watch time to process
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []bool{true, false, true, false}, changes, "IsRunning should be reactive")
}
