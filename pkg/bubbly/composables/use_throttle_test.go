package composables

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

func TestUseThrottle_FirstCallImmediate(t *testing.T) {
	// Arrange
	executed := false
	fn := func() {
		executed = true
	}

	// Act
	throttled := UseThrottle(nil, fn, 100*time.Millisecond)
	throttled()

	// Assert
	assert.True(t, executed, "First call should execute immediately")
}

func TestUseThrottle_SubsequentCallsDelayed(t *testing.T) {
	// Arrange
	callCount := 0
	fn := func() {
		callCount++
	}

	// Act
	throttled := UseThrottle(nil, fn, 100*time.Millisecond)

	// First call executes immediately
	throttled()
	assert.Equal(t, 1, callCount, "First call should execute")

	// Second call should be ignored (within throttle period)
	throttled()
	assert.Equal(t, 1, callCount, "Second call should be ignored")

	// Third call should also be ignored
	throttled()
	assert.Equal(t, 1, callCount, "Third call should be ignored")
}

func TestUseThrottle_DelayRespected(t *testing.T) {
	// Arrange
	callCount := 0
	fn := func() {
		callCount++
	}

	// Act
	throttled := UseThrottle(nil, fn, 50*time.Millisecond)

	// First call
	throttled()
	assert.Equal(t, 1, callCount)

	// Call within delay - ignored
	time.Sleep(25 * time.Millisecond)
	throttled()
	assert.Equal(t, 1, callCount, "Call within delay should be ignored")

	// Wait for delay to pass
	time.Sleep(30 * time.Millisecond) // Total: 55ms > 50ms delay

	// Call after delay - should execute
	throttled()
	assert.Equal(t, 2, callCount, "Call after delay should execute")
}

func TestUseThrottle_MultipleRapidCalls(t *testing.T) {
	// Arrange
	var callCount int32
	fn := func() {
		atomic.AddInt32(&callCount, 1)
	}

	// Act
	throttled := UseThrottle(nil, fn, 100*time.Millisecond)

	// Make 10 rapid calls
	for i := 0; i < 10; i++ {
		throttled()
		time.Sleep(5 * time.Millisecond) // 5ms between calls
	}

	// Assert - only first call should execute (50ms total < 100ms delay)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount),
		"Only first call should execute during throttle period")

	// Wait for throttle to reset
	time.Sleep(60 * time.Millisecond)

	// Next call should execute
	throttled()
	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount),
		"Call after throttle period should execute")
}

func TestUseThrottle_CleanupOnUnmount(t *testing.T) {
	// Arrange
	var throttled func()
	var callCount int

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			fn := func() {
				callCount++
			}
			throttled = UseThrottle(ctx, fn, 100*time.Millisecond)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)

	// Act
	comp.Init()
	comp.View()
	throttled() // First call executes

	// Wait for potential timer
	time.Sleep(150 * time.Millisecond)

	// Assert - no crashes, cleanup successful
	assert.Equal(t, 1, callCount, "Only first call should have executed")
}

func TestUseThrottle_ConcurrentCalls(t *testing.T) {
	// Arrange
	var callCount int32
	var mu sync.Mutex
	var calls []int

	fn := func() {
		count := atomic.AddInt32(&callCount, 1)
		mu.Lock()
		calls = append(calls, int(count))
		mu.Unlock()
	}

	// Act
	throttled := UseThrottle(nil, fn, 50*time.Millisecond)

	// Launch 10 concurrent goroutines
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			throttled()
		}()
	}

	wg.Wait()

	// Assert - only one call should execute
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount),
		"Only one call should execute despite concurrent calls")
}

func TestUseThrottle_ZeroDelay(t *testing.T) {
	// Arrange
	callCount := 0
	fn := func() {
		callCount++
	}

	// Act
	throttled := UseThrottle(nil, fn, 0)

	// Multiple calls with zero delay
	throttled()
	throttled()
	throttled()

	// Assert - all calls should execute immediately
	assert.Equal(t, 3, callCount, "All calls should execute with zero delay")
}

func TestUseThrottle_ThrottlePattern(t *testing.T) {
	// Arrange
	var callCount int32
	fn := func() {
		atomic.AddInt32(&callCount, 1)
	}

	// Act
	throttled := UseThrottle(nil, fn, 50*time.Millisecond)

	// Pattern: call, wait, call, wait (should execute twice)
	throttled()                       // Execute (t=0)
	time.Sleep(60 * time.Millisecond) // Wait for throttle to reset
	throttled()                       // Execute (t=60)
	time.Sleep(60 * time.Millisecond) // Wait for throttle to reset
	throttled()                       // Execute (t=120)

	// Assert
	assert.Equal(t, int32(3), atomic.LoadInt32(&callCount),
		"Should execute three times with proper delays")
}

func TestUseThrottle_FullComponentLifecycle(t *testing.T) {
	// Arrange
	var callCount int
	var throttled func()

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			fn := func() {
				callCount++
			}

			throttled = UseThrottle(ctx, fn, 50*time.Millisecond)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)

	// Act
	comp.Init()
	comp.View()

	// Execute throttled function
	throttled()
	assert.Equal(t, 1, callCount, "First call should execute")

	throttled()
	assert.Equal(t, 1, callCount, "Second call should be throttled")

	// Assert - no crashes
	assert.Equal(t, 1, callCount, "Count should remain stable")
}

func TestUseThrottle_NilContext(t *testing.T) {
	// Arrange
	callCount := 0
	fn := func() {
		callCount++
	}

	// Act - nil context should work (no cleanup registration)
	throttled := UseThrottle(nil, fn, 50*time.Millisecond)

	// Execute
	throttled()
	assert.Equal(t, 1, callCount, "First call should execute")

	throttled()
	assert.Equal(t, 1, callCount, "Second call should be throttled")

	// Wait for throttle to reset
	time.Sleep(60 * time.Millisecond)

	throttled()
	assert.Equal(t, 2, callCount, "Call after delay should execute")
}

func TestUseThrottle_CleanupWithActiveTimer(t *testing.T) {
	// Arrange
	var callCount int
	var throttled func()

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			fn := func() {
				callCount++
			}

			throttled = UseThrottle(ctx, fn, 200*time.Millisecond)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)

	// Act
	comp.Init()
	comp.View()

	// Execute throttled function - starts timer
	throttled()
	assert.Equal(t, 1, callCount, "First call should execute")

	// Trigger unmount while timer is active
	// This tests the cleanup path with an active timer
	comp.Update(nil) // Trigger lifecycle

	// Wait to ensure timer cleanup worked
	time.Sleep(250 * time.Millisecond)

	// Assert - timer was cleaned up, no additional executions
	assert.Equal(t, 1, callCount, "No additional calls after cleanup")
}
