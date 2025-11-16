package testutil

import (
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/stretchr/testify/assert"
)

// TestUseDebounceTester_Creation tests creating a new UseDebounceTester
func TestUseDebounceTester_Creation(t *testing.T) {
	ts := NewTimeSimulator()

	// Create a simple component with debounced value (using short delay for fast tests)
	comp, err := bubbly.NewComponent("TestDebounce").
		Setup(func(ctx *bubbly.Context) {
			source := ctx.Ref("initial")
			debounced := composables.UseDebounce(ctx, source, 50*time.Millisecond)
			ctx.Expose("source", source)
			ctx.Expose("debounced", debounced)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()

	tester := NewUseDebounceTester(comp, ts)

	assert.NotNil(t, tester, "UseDebounceTester should not be nil")
	assert.NotNil(t, tester.timeSim, "TimeSimulator should be set")
	assert.NotNil(t, tester.component, "Component should be set")
	assert.NotNil(t, tester.debounced, "Debounced ref should be extracted")
}

// TestUseDebounceTester_DebounceDelaysUpdate tests that debounce delays value updates
func TestUseDebounceTester_DebounceDelaysUpdate(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		newValue      string
		advanceTime   time.Duration
		debounceDelay time.Duration
		shouldUpdate  bool
	}{
		{
			name:          "no update before delay",
			initialValue:  "initial",
			newValue:      "changed",
			advanceTime:   20 * time.Millisecond,
			debounceDelay: 50 * time.Millisecond,
			shouldUpdate:  false,
		},
		{
			name:          "update after delay",
			initialValue:  "initial",
			newValue:      "changed",
			advanceTime:   60 * time.Millisecond,
			debounceDelay: 50 * time.Millisecond,
			shouldUpdate:  true,
		},
		{
			name:          "update after exceeding delay",
			initialValue:  "initial",
			newValue:      "changed",
			advanceTime:   100 * time.Millisecond,
			debounceDelay: 50 * time.Millisecond,
			shouldUpdate:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimeSimulator()

			comp, err := bubbly.NewComponent("TestDebounce").
				Setup(func(ctx *bubbly.Context) {
					source := ctx.Ref(tt.initialValue)
					debounced := composables.UseDebounce(ctx, source, tt.debounceDelay)
					ctx.Expose("source", source)
					ctx.Expose("debounced", debounced)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			comp.Init()

			tester := NewUseDebounceTester(comp, ts)

			// Trigger change
			tester.TriggerChange(tt.newValue)

			// Advance time
			tester.AdvanceTime(tt.advanceTime)

			// Check if debounced value updated
			actual := tester.debounced.Get().(string)

			if tt.shouldUpdate {
				assert.Equal(t, tt.newValue, actual, "debounced value should update after delay")
			} else {
				assert.Equal(t, tt.initialValue, actual, "debounced value should not update before delay")
			}
		})
	}
}

// TestUseDebounceTester_MultipleCancels tests that multiple changes cancel previous timers
func TestUseDebounceTester_MultipleCancels(t *testing.T) {
	ts := NewTimeSimulator()

	comp, err := bubbly.NewComponent("TestDebounce").
		Setup(func(ctx *bubbly.Context) {
			source := ctx.Ref("initial")
			debounced := composables.UseDebounce(ctx, source, 50*time.Millisecond)
			ctx.Expose("source", source)
			ctx.Expose("debounced", debounced)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()

	tester := NewUseDebounceTester(comp, ts)

	// Trigger multiple rapid changes
	tester.TriggerChange("change1")
	tester.AdvanceTime(20 * time.Millisecond)

	tester.TriggerChange("change2")
	tester.AdvanceTime(20 * time.Millisecond)

	tester.TriggerChange("change3")
	tester.AdvanceTime(20 * time.Millisecond)

	// Still shouldn't have updated (only 60ms total, but timer keeps resetting)
	actual := tester.debounced.Get().(string)
	assert.Equal(t, "initial", actual, "debounced value should not update with rapid changes")

	// Now advance past the delay from last change
	tester.AdvanceTime(40 * time.Millisecond) // Total 100ms from last change

	// Should now have the final value
	actual = tester.debounced.Get().(string)
	assert.Equal(t, "change3", actual, "debounced value should have final value after delay")
}

// TestUseDebounceTester_FinalValueCorrect tests that final value is correct
func TestUseDebounceTester_FinalValueCorrect(t *testing.T) {
	tests := []struct {
		name          string
		changes       []string
		delayBetween  time.Duration
		finalDelay    time.Duration
		debounceDelay time.Duration
		expectedFinal string
	}{
		{
			name:          "single change",
			changes:       []string{"value1"},
			delayBetween:  0,
			finalDelay:    60 * time.Millisecond,
			debounceDelay: 50 * time.Millisecond,
			expectedFinal: "value1",
		},
		{
			name:          "multiple rapid changes",
			changes:       []string{"value1", "value2", "value3"},
			delayBetween:  10 * time.Millisecond,
			finalDelay:    60 * time.Millisecond,
			debounceDelay: 50 * time.Millisecond,
			expectedFinal: "value3",
		},
		{
			name:          "changes with long delay between",
			changes:       []string{"value1", "value2"},
			delayBetween:  70 * time.Millisecond,
			finalDelay:    70 * time.Millisecond,
			debounceDelay: 50 * time.Millisecond,
			expectedFinal: "value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimeSimulator()

			comp, err := bubbly.NewComponent("TestDebounce").
				Setup(func(ctx *bubbly.Context) {
					source := ctx.Ref("initial")
					debounced := composables.UseDebounce(ctx, source, tt.debounceDelay)
					ctx.Expose("source", source)
					ctx.Expose("debounced", debounced)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()

			assert.NoError(t, err)
			comp.Init()

			tester := NewUseDebounceTester(comp, ts)

			// Apply all changes
			for _, change := range tt.changes {
				tester.TriggerChange(change)
				if tt.delayBetween > 0 {
					tester.AdvanceTime(tt.delayBetween)
				}
			}

			// Advance final delay
			tester.AdvanceTime(tt.finalDelay)

			// Check final value
			actual := tester.debounced.Get().(string)
			assert.Equal(t, tt.expectedFinal, actual, "final debounced value should be correct")
		})
	}
}

// TestUseDebounceTester_TypeSafety tests type safety with different types
func TestUseDebounceTester_TypeSafety(t *testing.T) {
	t.Run("integer values", func(t *testing.T) {
		ts := NewTimeSimulator()

		comp, err := bubbly.NewComponent("TestDebounce").
			Setup(func(ctx *bubbly.Context) {
				source := ctx.Ref(0)
				debounced := composables.UseDebounce(ctx, source, 50*time.Millisecond)
				ctx.Expose("source", source)
				ctx.Expose("debounced", debounced)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()

		assert.NoError(t, err)
		comp.Init()

		tester := NewUseDebounceTester(comp, ts)

		tester.TriggerChange(42)
		tester.AdvanceTime(60 * time.Millisecond)

		actual := tester.debounced.Get().(int)
		assert.Equal(t, 42, actual)
	})

	t.Run("boolean values", func(t *testing.T) {
		ts := NewTimeSimulator()

		comp, err := bubbly.NewComponent("TestDebounce").
			Setup(func(ctx *bubbly.Context) {
				source := ctx.Ref(false)
				debounced := composables.UseDebounce(ctx, source, 50*time.Millisecond)
				ctx.Expose("source", source)
				ctx.Expose("debounced", debounced)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()

		assert.NoError(t, err)
		comp.Init()

		tester := NewUseDebounceTester(comp, ts)

		tester.TriggerChange(true)
		tester.AdvanceTime(60 * time.Millisecond)

		actual := tester.debounced.Get().(bool)
		assert.True(t, actual)
	})
}

// TestUseDebounceTester_ZeroDelay tests debounce with zero delay
func TestUseDebounceTester_ZeroDelay(t *testing.T) {
	ts := NewTimeSimulator()

	comp, err := bubbly.NewComponent("TestDebounce").
		Setup(func(ctx *bubbly.Context) {
			source := ctx.Ref("initial")
			debounced := composables.UseDebounce(ctx, source, 0)
			ctx.Expose("source", source)
			ctx.Expose("debounced", debounced)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()

	tester := NewUseDebounceTester(comp, ts)

	tester.TriggerChange("immediate")
	// Even with zero delay, time.AfterFunc schedules asynchronously
	// so we need a tiny wait for the goroutine to execute
	tester.AdvanceTime(5 * time.Millisecond)

	// With zero delay, should update almost immediately
	actual := tester.debounced.Get().(string)
	assert.Equal(t, "immediate", actual, "zero delay should update almost immediately")
}

// TestUseDebounceTester_GetDebouncedValue tests GetDebouncedValue method
func TestUseDebounceTester_GetDebouncedValue(t *testing.T) {
	ts := NewTimeSimulator()

	comp, err := bubbly.NewComponent("TestDebounce").
		Setup(func(ctx *bubbly.Context) {
			source := ctx.Ref("initial")
			debounced := composables.UseDebounce(ctx, source, 50*time.Millisecond)
			ctx.Expose("source", source)
			ctx.Expose("debounced", debounced)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()

	tester := NewUseDebounceTester(comp, ts)

	// Test initial value
	value := tester.GetDebouncedValue()
	assert.Equal(t, "initial", value, "should return initial debounced value")

	// Change source value
	tester.TriggerChange("new value")
	
	// Should still return old value (debounced)
	value = tester.GetDebouncedValue()
	assert.Equal(t, "initial", value, "should still return initial value before debounce")

	// Advance time past debounce delay
	tester.AdvanceTime(60 * time.Millisecond)
	
	// Now should return new value
	value = tester.GetDebouncedValue()
	assert.Equal(t, "new value", value, "should return debounced value after delay")
}

// TestUseDebounceTester_GetSourceValue tests GetSourceValue method
func TestUseDebounceTester_GetSourceValue(t *testing.T) {
	ts := NewTimeSimulator()

	comp, err := bubbly.NewComponent("TestDebounce").
		Setup(func(ctx *bubbly.Context) {
			source := ctx.Ref("initial")
			debounced := composables.UseDebounce(ctx, source, 50*time.Millisecond)
			ctx.Expose("source", source)
			ctx.Expose("debounced", debounced)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()

	tester := NewUseDebounceTester(comp, ts)

	// Test initial value
	value := tester.GetSourceValue()
	assert.Equal(t, "initial", value, "should return initial source value")

	// Change source value
	tester.TriggerChange("new value")
	
	// Should immediately return new value (not debounced)
	value = tester.GetSourceValue()
	assert.Equal(t, "new value", value, "should immediately return new source value")
}
