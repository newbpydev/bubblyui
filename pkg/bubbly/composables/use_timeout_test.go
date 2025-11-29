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

// TestUseTimeout_StartsNotPending verifies that UseTimeout starts in not pending state.
func TestUseTimeout_StartsNotPending(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	timeout := UseTimeout(ctx, callback, 100*time.Millisecond)

	assert.False(t, timeout.IsPending.GetTyped(), "timeout should start not pending")
	assert.False(t, timeout.IsExpired.GetTyped(), "timeout should start not expired")
}

// TestUseTimeout_Start_BeginsTimeout verifies that Start() begins the timeout.
func TestUseTimeout_Start_BeginsTimeout(t *testing.T) {
	ctx := createTestContext()
	var called int32

	timeout := UseTimeout(ctx, func() {
		atomic.AddInt32(&called, 1)
	}, 20*time.Millisecond)

	timeout.Start()
	assert.True(t, timeout.IsPending.GetTyped(), "timeout should be pending after Start()")
	assert.False(t, timeout.IsExpired.GetTyped(), "timeout should not be expired yet")

	// Wait for timeout to fire
	time.Sleep(50 * time.Millisecond)

	assert.False(t, timeout.IsPending.GetTyped(), "timeout should not be pending after expiry")
	assert.True(t, timeout.IsExpired.GetTyped(), "timeout should be expired after duration")
	assert.Equal(t, int32(1), atomic.LoadInt32(&called), "callback should have been called once")
}

// TestUseTimeout_Cancel_StopsPendingTimeout verifies that Cancel() stops a pending timeout.
func TestUseTimeout_Cancel_StopsPendingTimeout(t *testing.T) {
	ctx := createTestContext()
	var called int32

	timeout := UseTimeout(ctx, func() {
		atomic.AddInt32(&called, 1)
	}, 30*time.Millisecond)

	timeout.Start()
	assert.True(t, timeout.IsPending.GetTyped())

	// Cancel before it fires
	time.Sleep(10 * time.Millisecond)
	timeout.Cancel()

	assert.False(t, timeout.IsPending.GetTyped(), "timeout should not be pending after Cancel()")
	assert.False(t, timeout.IsExpired.GetTyped(), "timeout should not be expired after Cancel()")

	// Wait past original duration
	time.Sleep(40 * time.Millisecond)

	assert.Equal(t, int32(0), atomic.LoadInt32(&called), "callback should not have been called after Cancel()")
}

// TestUseTimeout_Reset_CancelsAndRestarts verifies that Reset() cancels and restarts the timeout.
func TestUseTimeout_Reset_CancelsAndRestarts(t *testing.T) {
	ctx := createTestContext()
	var called int32

	timeout := UseTimeout(ctx, func() {
		atomic.AddInt32(&called, 1)
	}, 30*time.Millisecond)

	timeout.Start()
	time.Sleep(15 * time.Millisecond)

	// Reset should restart the timer
	timeout.Reset()
	assert.True(t, timeout.IsPending.GetTyped(), "timeout should be pending after Reset()")
	assert.False(t, timeout.IsExpired.GetTyped(), "timeout should not be expired after Reset()")

	// Wait for original duration (should not fire yet due to reset)
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(&called), "callback should not have fired yet (reset extended)")

	// Wait for reset duration to complete
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&called), "callback should have been called after reset duration")
	assert.True(t, timeout.IsExpired.GetTyped(), "timeout should be expired")
}

// TestUseTimeout_CallbackExecutedOnExpiry verifies that the callback is executed when timeout expires.
func TestUseTimeout_CallbackExecutedOnExpiry(t *testing.T) {
	ctx := createTestContext()
	var executionTime time.Time
	var mu sync.Mutex

	startTime := time.Now()
	timeout := UseTimeout(ctx, func() {
		mu.Lock()
		executionTime = time.Now()
		mu.Unlock()
	}, 25*time.Millisecond)

	timeout.Start()
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	elapsed := executionTime.Sub(startTime)
	mu.Unlock()

	// Should have executed around 25ms mark (with some tolerance)
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(20), "callback should execute after duration")
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(50), "callback should not execute too late")
}

// TestUseTimeout_IsExpiredSetAfterExecution verifies that IsExpired is set after callback execution.
func TestUseTimeout_IsExpiredSetAfterExecution(t *testing.T) {
	ctx := createTestContext()
	var callbackDone int32

	timeout := UseTimeout(ctx, func() {
		atomic.StoreInt32(&callbackDone, 1)
	}, 15*time.Millisecond)

	assert.False(t, timeout.IsExpired.GetTyped(), "should not be expired initially")

	timeout.Start()
	time.Sleep(30 * time.Millisecond)

	assert.True(t, timeout.IsExpired.GetTyped(), "should be expired after callback")
	assert.Equal(t, int32(1), atomic.LoadInt32(&callbackDone), "callback should have executed")
}

// TestUseTimeout_CleanupOnUnmount verifies that the timeout is cancelled on component unmount.
func TestUseTimeout_CleanupOnUnmount(t *testing.T) {
	var called int32
	var timeout *TimeoutReturn

	// Create a component with proper lifecycle
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			timeout = UseTimeout(ctx, func() {
				atomic.AddInt32(&called, 1)
			}, 30*time.Millisecond)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return ""
		}).
		Build()
	require.NoError(t, err)

	// Initialize the component (triggers Setup)
	component.Init()

	// Start the timeout
	timeout.Start()
	assert.True(t, timeout.IsPending.GetTyped())

	// Unmount the component (should trigger OnUnmounted and cancel timeout)
	time.Sleep(10 * time.Millisecond)
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Wait past original duration
	time.Sleep(40 * time.Millisecond)

	assert.Equal(t, int32(0), atomic.LoadInt32(&called), "callback should not be called after unmount")
	assert.False(t, timeout.IsPending.GetTyped(), "timeout should not be pending after unmount")
}

// TestUseTimeout_NegativeDuration_Panics verifies that negative duration causes panic.
func TestUseTimeout_NegativeDuration_Panics(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	assert.Panics(t, func() {
		UseTimeout(ctx, callback, -1*time.Millisecond)
	}, "UseTimeout should panic with negative duration")
}

// TestUseTimeout_ZeroDuration_Panics verifies that zero duration causes panic.
func TestUseTimeout_ZeroDuration_Panics(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	assert.Panics(t, func() {
		UseTimeout(ctx, callback, 0)
	}, "UseTimeout should panic with zero duration")
}

// TestUseTimeout_MultipleStarts_NoOp verifies that multiple Start() calls are no-op.
func TestUseTimeout_MultipleStarts_NoOp(t *testing.T) {
	ctx := createTestContext()
	var called int32

	timeout := UseTimeout(ctx, func() {
		atomic.AddInt32(&called, 1)
	}, 20*time.Millisecond)

	// Multiple starts should not create multiple timers
	timeout.Start()
	timeout.Start()
	timeout.Start()

	assert.True(t, timeout.IsPending.GetTyped())

	time.Sleep(50 * time.Millisecond)

	// Callback should only be called once
	assert.Equal(t, int32(1), atomic.LoadInt32(&called), "callback should only be called once")
}

// TestUseTimeout_MultipleCancels_NoOp verifies that multiple Cancel() calls are no-op.
func TestUseTimeout_MultipleCancels_NoOp(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	timeout := UseTimeout(ctx, callback, 100*time.Millisecond)

	// Multiple cancels should not panic
	assert.NotPanics(t, func() {
		timeout.Cancel()
		timeout.Cancel()
		timeout.Cancel()
	}, "multiple Cancel() calls should not panic")

	assert.False(t, timeout.IsPending.GetTyped())
}

// TestUseTimeout_ConcurrentStartCancel verifies thread safety with concurrent Start/Cancel calls.
func TestUseTimeout_ConcurrentStartCancel(t *testing.T) {
	ctx := createTestContext()
	var called int32

	timeout := UseTimeout(ctx, func() {
		atomic.AddInt32(&called, 1)
	}, 10*time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				timeout.Start()
			} else {
				timeout.Cancel()
			}
		}(i)
	}

	wg.Wait()

	// Should not panic and should be in a consistent state
	_ = timeout.IsPending.GetTyped()
	_ = timeout.IsExpired.GetTyped()

	// Cleanup
	timeout.Cancel()
}

// TestUseTimeout_WorksWithCreateShared verifies integration with CreateShared pattern.
func TestUseTimeout_WorksWithCreateShared(t *testing.T) {
	var sharedTimeout *TimeoutReturn
	var called int32

	UseSharedTimeout := CreateShared(func(ctx *bubbly.Context) *TimeoutReturn {
		return UseTimeout(ctx, func() {
			atomic.AddInt32(&called, 1)
		}, 20*time.Millisecond)
	})

	// First call creates the timeout
	ctx1 := createTestContext()
	sharedTimeout = UseSharedTimeout(ctx1)
	require.NotNil(t, sharedTimeout)

	// Second call returns the same instance
	ctx2 := createTestContext()
	sameTimeout := UseSharedTimeout(ctx2)
	assert.Same(t, sharedTimeout, sameTimeout, "CreateShared should return same instance")

	// Verify it works
	sharedTimeout.Start()
	time.Sleep(40 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called), "shared timeout callback should have been called")
	assert.True(t, sharedTimeout.IsExpired.GetTyped())
}

// TestUseTimeout_NilContext verifies that nil context is handled gracefully.
func TestUseTimeout_NilContext(t *testing.T) {
	var called int32

	// Should not panic with nil context
	timeout := UseTimeout(nil, func() {
		atomic.AddInt32(&called, 1)
	}, 15*time.Millisecond)

	require.NotNil(t, timeout)
	assert.False(t, timeout.IsPending.GetTyped())

	timeout.Start()
	time.Sleep(30 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	assert.True(t, timeout.IsExpired.GetTyped())
}

// TestUseTimeout_StartAfterExpiry verifies that Start() after expiry restarts the timeout.
func TestUseTimeout_StartAfterExpiry(t *testing.T) {
	ctx := createTestContext()
	var called int32

	timeout := UseTimeout(ctx, func() {
		atomic.AddInt32(&called, 1)
	}, 15*time.Millisecond)

	// First run
	timeout.Start()
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&called))
	assert.True(t, timeout.IsExpired.GetTyped())

	// Start again after expiry
	timeout.Start()
	assert.True(t, timeout.IsPending.GetTyped(), "should be pending after restart")
	assert.False(t, timeout.IsExpired.GetTyped(), "should reset expired state on restart")

	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, int32(2), atomic.LoadInt32(&called), "callback should be called again")
}

// TestUseTimeout_IsPendingReactive verifies that IsPending ref is reactive.
func TestUseTimeout_IsPendingReactive(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	timeout := UseTimeout(ctx, callback, 20*time.Millisecond)

	// Track changes with Watch
	var changes []bool
	var mu sync.Mutex

	bubbly.Watch(timeout.IsPending, func(newVal, oldVal bool) {
		mu.Lock()
		changes = append(changes, newVal)
		mu.Unlock()
	})

	timeout.Start()
	time.Sleep(40 * time.Millisecond) // Wait for expiry

	// Give Watch time to process
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []bool{true, false}, changes, "IsPending should be reactive (true on start, false on expiry)")
}

// TestUseTimeout_IsExpiredReactive verifies that IsExpired ref is reactive.
func TestUseTimeout_IsExpiredReactive(t *testing.T) {
	ctx := createTestContext()
	callback := func() {}

	timeout := UseTimeout(ctx, callback, 20*time.Millisecond)

	// Track changes with Watch
	var changes []bool
	var mu sync.Mutex

	bubbly.Watch(timeout.IsExpired, func(newVal, oldVal bool) {
		mu.Lock()
		changes = append(changes, newVal)
		mu.Unlock()
	})

	timeout.Start()
	time.Sleep(40 * time.Millisecond) // Wait for expiry

	// Give Watch time to process
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Contains(t, changes, true, "IsExpired should become true after expiry")
}

// TestUseTimeout_CancelAfterExpiry_NoOp verifies that Cancel() after expiry is a no-op.
func TestUseTimeout_CancelAfterExpiry_NoOp(t *testing.T) {
	ctx := createTestContext()
	var called int32

	timeout := UseTimeout(ctx, func() {
		atomic.AddInt32(&called, 1)
	}, 15*time.Millisecond)

	timeout.Start()
	time.Sleep(30 * time.Millisecond)

	assert.True(t, timeout.IsExpired.GetTyped())
	assert.Equal(t, int32(1), atomic.LoadInt32(&called))

	// Cancel after expiry should not panic or change state
	assert.NotPanics(t, func() {
		timeout.Cancel()
	})

	assert.True(t, timeout.IsExpired.GetTyped(), "IsExpired should remain true after Cancel()")
}
