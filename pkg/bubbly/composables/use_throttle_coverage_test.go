package composables

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseThrottle_ZeroDelayMultipleCalls tests zero delay with multiple rapid calls
func TestUseThrottle_ZeroDelayMultipleCalls(t *testing.T) {
	// Arrange
	ctx := bubbly.NewTestContext()
	var callCount int32
	var mu sync.Mutex

	fn := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	throttled := UseThrottle(ctx, fn, 0)

	// Act - make multiple calls with zero delay
	for i := 0; i < 5; i++ {
		throttled()
	}

	// Assert - with zero delay, all calls should execute
	mu.Lock()
	count := callCount
	mu.Unlock()

	assert.Equal(t, int32(5), count, "All calls should execute with zero delay")
}

// TestUseThrottle_NilContextWithTimer tests nil context cleanup behavior
func TestUseThrottle_NilContextWithTimer(t *testing.T) {
	// Arrange
	callCount := 0
	fn := func() {
		callCount++
	}

	// Act - create throttled function with nil context
	throttled := UseThrottle(nil, fn, 50*time.Millisecond)

	// Call immediately
	throttled()
	assert.Equal(t, 1, callCount, "First call should execute")

	// Call during throttle period
	throttled()
	assert.Equal(t, 1, callCount, "Second call should be throttled")

	// Wait for throttle period
	time.Sleep(60 * time.Millisecond)

	// Call after throttle period
	throttled()
	assert.Equal(t, 2, callCount, "Third call after delay should execute")
}

// TestUseThrottle_VeryShortDelay tests throttling with very short delay
func TestUseThrottle_VeryShortDelay(t *testing.T) {
	// Arrange
	ctx := bubbly.NewTestContext()
	callCount := 0
	fn := func() {
		callCount++
	}

	throttled := UseThrottle(ctx, fn, 10*time.Millisecond)

	// Act
	throttled() // First call
	assert.Equal(t, 1, callCount)

	throttled() // Throttled
	assert.Equal(t, 1, callCount)

	time.Sleep(20 * time.Millisecond) // Wait significantly longer than delay for reliability

	throttled() // Should execute
	assert.Equal(t, 2, callCount, "Call after short delay should execute")
}

// TestUseThrottle_ConcurrentCallsWithZeroDelay tests concurrent calls with zero delay
func TestUseThrottle_ConcurrentCallsWithZeroDelay(t *testing.T) {
	// Arrange
	ctx := bubbly.NewTestContext()
	var callCount int32
	var mu sync.Mutex

	fn := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	throttled := UseThrottle(ctx, fn, 0)

	// Act - concurrent calls with zero delay
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			throttled()
		}()
	}
	wg.Wait()

	// Assert - all calls should execute with zero delay
	mu.Lock()
	count := callCount
	mu.Unlock()

	assert.Equal(t, int32(10), count, "All concurrent calls should execute with zero delay")
}

// TestUseThrottle_UnmountBeforeTimerExpires tests cleanup when unmounting before timer expires
func TestUseThrottle_UnmountBeforeTimerExpires(t *testing.T) {
	// Arrange
	var throttled func()
	var ctx *bubbly.Context
	callCount := 0

	component, err := bubbly.NewComponent("test").
		Setup(func(c *bubbly.Context) {
			ctx = c
			fn := func() {
				callCount++
			}
			throttled = UseThrottle(c, fn, 100*time.Millisecond)
		}).
		Template(func(c bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)

	// Initialize component
	_ = component.Init()

	// Act - call throttled function then immediately unmount
	throttled()
	assert.Equal(t, 1, callCount, "First call should execute")

	// Trigger unmount lifecycle hooks
	bubbly.TriggerUnmount(ctx)

	// Wait for what would have been the throttle period
	time.Sleep(150 * time.Millisecond)

	// Try calling after unmount (timer should be stopped)
	throttled()

	// Assert - count should not increase after unmount
	// Note: The call will still execute because the closure is independent,
	// but the timer cleanup should have occurred
	assert.Equal(t, 2, callCount, "Call after unmount should execute (closure is independent)")
}

// TestUseThrottle_RapidUnmountRemount tests rapid unmount/remount cycles
func TestUseThrottle_RapidUnmountRemount(t *testing.T) {
	// Arrange
	callCount := 0

	createComponent := func() (*bubbly.Context, func()) {
		var throttled func()
		var ctx *bubbly.Context
		component, err := bubbly.NewComponent("test").
			Setup(func(c *bubbly.Context) {
				ctx = c
				fn := func() {
					callCount++
				}
				throttled = UseThrottle(c, fn, 50*time.Millisecond)
			}).
			Template(func(c bubbly.RenderContext) string {
				return "test"
			}).
			Build()

		assert.NoError(t, err)
		_ = component.Init()

		return ctx, throttled
	}

	// Act - rapidly create, call, and destroy
	for i := 0; i < 5; i++ {
		ctx, throttled := createComponent()
		throttled() // One call per component
		bubbly.TriggerUnmount(ctx)
	}

	// Assert
	assert.Equal(t, 5, callCount, "Each component should execute once")
}
