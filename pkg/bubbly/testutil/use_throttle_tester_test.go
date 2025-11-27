package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// TestUseThrottleTester_BasicThrottling tests that throttle delays subsequent calls
func TestUseThrottleTester_BasicThrottling(t *testing.T) {
	tests := []struct {
		name          string
		delay         time.Duration
		calls         int
		expectedCalls int
		waitBetween   time.Duration
	}{
		{
			name:          "immediate execution on first call",
			delay:         100 * time.Millisecond,
			calls:         1,
			expectedCalls: 1,
			waitBetween:   0,
		},
		{
			name:          "subsequent calls within delay ignored",
			delay:         100 * time.Millisecond,
			calls:         5,
			expectedCalls: 1, // Only first call executes
			waitBetween:   10 * time.Millisecond,
		},
		{
			name:          "calls after delay execute",
			delay:         50 * time.Millisecond,
			calls:         3,
			expectedCalls: 3, // All execute because we wait between calls
			waitBetween:   60 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component with UseThrottle
			comp, err := bubbly.NewComponent("TestThrottle").
				Setup(func(ctx *bubbly.Context) {
					callCount := ctx.Ref(0)

					throttledFn := composables.UseThrottle(ctx, func() {
						count := callCount.Get().(int)
						callCount.Set(count + 1)
					}, tt.delay)

					ctx.Expose("callCount", callCount)
					ctx.Expose("throttledFn", throttledFn)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)
			comp.Init()

			tester := NewUseThrottleTester(comp)

			// Trigger calls
			for i := 0; i < tt.calls; i++ {
				tester.TriggerThrottled()
				if tt.waitBetween > 0 && i < tt.calls-1 {
					tester.AdvanceTime(tt.waitBetween)
				}
			}

			// Wait for any pending executions
			if tt.waitBetween == 0 {
				tester.AdvanceTime(tt.delay + 10*time.Millisecond)
			}

			// Assert call count
			assert.Equal(t, tt.expectedCalls, tester.GetCallCount())
		})
	}
}

// TestUseThrottleTester_ZeroDelay tests throttle with zero delay (no throttling)
func TestUseThrottleTester_ZeroDelay(t *testing.T) {
	comp, err := bubbly.NewComponent("TestThrottle").
		Setup(func(ctx *bubbly.Context) {
			callCount := ctx.Ref(0)

			throttledFn := composables.UseThrottle(ctx, func() {
				count := callCount.Get().(int)
				callCount.Set(count + 1)
			}, 0)

			ctx.Expose("callCount", callCount)
			ctx.Expose("throttledFn", throttledFn)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseThrottleTester(comp)

	// All calls should execute immediately with zero delay
	for i := 0; i < 5; i++ {
		tester.TriggerThrottled()
	}

	assert.Equal(t, 5, tester.GetCallCount())
}

// TestUseThrottleTester_GetLastCallTime tests last call time tracking
func TestUseThrottleTester_GetLastCallTime(t *testing.T) {
	comp, err := bubbly.NewComponent("TestThrottle").
		Setup(func(ctx *bubbly.Context) {
			callCount := ctx.Ref(0)

			throttledFn := composables.UseThrottle(ctx, func() {
				count := callCount.Get().(int)
				callCount.Set(count + 1)
			}, 100*time.Millisecond)

			ctx.Expose("callCount", callCount)
			ctx.Expose("throttledFn", throttledFn)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseThrottleTester(comp)

	// First call
	before := time.Now()
	tester.TriggerThrottled()
	after := time.Now()

	lastCall := tester.GetLastCallTime()
	assert.True(t, lastCall.After(before) || lastCall.Equal(before))
	assert.True(t, lastCall.Before(after) || lastCall.Equal(after))
}

// TestUseThrottleTester_IsThrottled tests throttled state checking
func TestUseThrottleTester_IsThrottled(t *testing.T) {
	comp, err := bubbly.NewComponent("TestThrottle").
		Setup(func(ctx *bubbly.Context) {
			callCount := ctx.Ref(0)

			throttledFn := composables.UseThrottle(ctx, func() {
				count := callCount.Get().(int)
				callCount.Set(count + 1)
			}, 100*time.Millisecond)

			ctx.Expose("callCount", callCount)
			ctx.Expose("throttledFn", throttledFn)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseThrottleTester(comp)

	// Initially not throttled
	assert.False(t, tester.IsThrottled())

	// After first call, should be throttled
	tester.TriggerThrottled()
	assert.True(t, tester.IsThrottled())

	// After delay, should not be throttled
	tester.AdvanceTime(150 * time.Millisecond)
	assert.False(t, tester.IsThrottled())
}

// TestUseThrottleTester_MissingRefs tests panic when required refs not exposed
func TestUseThrottleTester_MissingRefs(t *testing.T) {
	comp, err := bubbly.NewComponent("TestThrottle").
		Setup(func(ctx *bubbly.Context) {
			// Don't expose required refs
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	assert.Panics(t, func() {
		NewUseThrottleTester(comp)
	})
}

// TestUseThrottleTester_RapidCalls tests many rapid calls
func TestUseThrottleTester_RapidCalls(t *testing.T) {
	comp, err := bubbly.NewComponent("TestThrottle").
		Setup(func(ctx *bubbly.Context) {
			callCount := ctx.Ref(0)

			throttledFn := composables.UseThrottle(ctx, func() {
				count := callCount.Get().(int)
				callCount.Set(count + 1)
			}, 50*time.Millisecond)

			ctx.Expose("callCount", callCount)
			ctx.Expose("throttledFn", throttledFn)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseThrottleTester(comp)

	// 100 rapid calls - only first should execute
	for i := 0; i < 100; i++ {
		tester.TriggerThrottled()
	}

	assert.Equal(t, 1, tester.GetCallCount())

	// Wait for throttle to reset
	tester.AdvanceTime(60 * time.Millisecond)

	// Next call should execute
	tester.TriggerThrottled()
	assert.Equal(t, 2, tester.GetCallCount())
}
