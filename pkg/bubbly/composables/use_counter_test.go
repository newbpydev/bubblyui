package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseCounter_InitialValueSetCorrectly tests that initial value is set correctly
func TestUseCounter_InitialValueSetCorrectly(t *testing.T) {
	tests := []struct {
		name    string
		initial int
	}{
		{"initial zero", 0},
		{"initial positive", 42},
		{"initial negative", -10},
		{"initial large", 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial)

			assert.NotNil(t, counter, "UseCounter should return non-nil")
			assert.NotNil(t, counter.Count, "Count should not be nil")
			assert.Equal(t, tt.initial, counter.Count.GetTyped(),
				"Initial value should be %d", tt.initial)
		})
	}
}

// TestUseCounter_IncrementByStep tests that Increment increases by step
func TestUseCounter_IncrementByStep(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		step     int
		expected int
	}{
		{"default step 1", 0, 0, 1}, // step 0 means default (1)
		{"custom step 5", 0, 5, 5},
		{"step from negative", -5, 3, -2},
		{"large step", 0, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			var counter *CounterReturn
			if tt.step == 0 {
				counter = UseCounter(ctx, tt.initial)
			} else {
				counter = UseCounter(ctx, tt.initial, WithStep(tt.step))
			}

			counter.Increment()

			assert.Equal(t, tt.expected, counter.Count.GetTyped(),
				"After Increment, count should be %d", tt.expected)
		})
	}
}

// TestUseCounter_DecrementByStep tests that Decrement decreases by step
func TestUseCounter_DecrementByStep(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		step     int
		expected int
	}{
		{"default step 1", 10, 0, 9}, // step 0 means default (1)
		{"custom step 5", 20, 5, 15},
		{"step to negative", 2, 5, -3},
		{"from zero", 0, 1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			var counter *CounterReturn
			if tt.step == 0 {
				counter = UseCounter(ctx, tt.initial)
			} else {
				counter = UseCounter(ctx, tt.initial, WithStep(tt.step))
			}

			counter.Decrement()

			assert.Equal(t, tt.expected, counter.Count.GetTyped(),
				"After Decrement, count should be %d", tt.expected)
		})
	}
}

// TestUseCounter_MinBoundEnforced tests that min bound is enforced
func TestUseCounter_MinBoundEnforced(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		min      int
		expected int
	}{
		{"decrement stops at min", 5, 0, 0},
		{"initial below min clamped", -10, 0, 0},
		{"negative min", 0, -5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial, WithMin(tt.min))

			// Try to decrement below min
			for i := 0; i < 20; i++ {
				counter.Decrement()
			}

			assert.GreaterOrEqual(t, counter.Count.GetTyped(), tt.min,
				"Count should not go below min %d", tt.min)
			assert.Equal(t, tt.expected, counter.Count.GetTyped(),
				"Count should be clamped to %d", tt.expected)
		})
	}
}

// TestUseCounter_MaxBoundEnforced tests that max bound is enforced
func TestUseCounter_MaxBoundEnforced(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		max      int
		expected int
	}{
		{"increment stops at max", 95, 100, 100},
		{"initial above max clamped", 150, 100, 100},
		{"negative max", -10, -5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial, WithMax(tt.max))

			// Try to increment above max
			for i := 0; i < 20; i++ {
				counter.Increment()
			}

			assert.LessOrEqual(t, counter.Count.GetTyped(), tt.max,
				"Count should not go above max %d", tt.max)
			assert.Equal(t, tt.expected, counter.Count.GetTyped(),
				"Count should be clamped to %d", tt.expected)
		})
	}
}

// TestUseCounter_IncrementByWorks tests that IncrementBy increases by n
func TestUseCounter_IncrementByWorks(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		by       int
		expected int
	}{
		{"increment by 10", 0, 10, 10},
		{"increment by 1", 5, 1, 6},
		{"increment by negative (decreases)", 10, -5, 5},
		{"increment by zero", 5, 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial)

			counter.IncrementBy(tt.by)

			assert.Equal(t, tt.expected, counter.Count.GetTyped(),
				"After IncrementBy(%d), count should be %d", tt.by, tt.expected)
		})
	}
}

// TestUseCounter_DecrementByWorks tests that DecrementBy decreases by n
func TestUseCounter_DecrementByWorks(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		by       int
		expected int
	}{
		{"decrement by 10", 20, 10, 10},
		{"decrement by 1", 5, 1, 4},
		{"decrement by negative (increases)", 10, -5, 15},
		{"decrement by zero", 5, 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial)

			counter.DecrementBy(tt.by)

			assert.Equal(t, tt.expected, counter.Count.GetTyped(),
				"After DecrementBy(%d), count should be %d", tt.by, tt.expected)
		})
	}
}

// TestUseCounter_SetClampsToBounds tests that Set clamps to bounds
func TestUseCounter_SetClampsToBounds(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		min      int
		max      int
		setValue int
		expected int
	}{
		{"set within bounds", 50, 0, 100, 75, 75},
		{"set below min clamped", 50, 0, 100, -10, 0},
		{"set above max clamped", 50, 0, 100, 150, 100},
		{"set to min", 50, 0, 100, 0, 0},
		{"set to max", 50, 0, 100, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial, WithMin(tt.min), WithMax(tt.max))

			counter.Set(tt.setValue)

			assert.Equal(t, tt.expected, counter.Count.GetTyped(),
				"After Set(%d), count should be clamped to %d", tt.setValue, tt.expected)
		})
	}
}

// TestUseCounter_ResetReturnsToInitial tests that Reset returns to initial value
func TestUseCounter_ResetReturnsToInitial(t *testing.T) {
	tests := []struct {
		name    string
		initial int
	}{
		{"reset to zero", 0},
		{"reset to positive", 50},
		{"reset to negative", -25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial)

			// Change the value
			counter.IncrementBy(100)
			assert.NotEqual(t, tt.initial, counter.Count.GetTyped(),
				"Value should have changed from initial")

			// Reset
			counter.Reset()

			assert.Equal(t, tt.initial, counter.Count.GetTyped(),
				"After Reset, count should be %d", tt.initial)
		})
	}
}

// TestUseCounter_DefaultStepIsOne tests that default step is 1
func TestUseCounter_DefaultStepIsOne(t *testing.T) {
	ctx := createTestContext()
	counter := UseCounter(ctx, 0)

	counter.Increment()
	assert.Equal(t, 1, counter.Count.GetTyped(), "Default step should be 1")

	counter.Decrement()
	assert.Equal(t, 0, counter.Count.GetTyped(), "Default step should be 1")
}

// TestUseCounter_BoundsWithStep tests bounds enforcement with custom step
func TestUseCounter_BoundsWithStep(t *testing.T) {
	ctx := createTestContext()
	counter := UseCounter(ctx, 95, WithMin(0), WithMax(100), WithStep(10))

	// Increment should stop at max
	counter.Increment()
	assert.Equal(t, 100, counter.Count.GetTyped(), "Should clamp to max 100")

	// Another increment should stay at max
	counter.Increment()
	assert.Equal(t, 100, counter.Count.GetTyped(), "Should stay at max 100")

	// Set to near min
	counter.Set(5)

	// Decrement should stop at min
	counter.Decrement()
	assert.Equal(t, 0, counter.Count.GetTyped(), "Should clamp to min 0")

	// Another decrement should stay at min
	counter.Decrement()
	assert.Equal(t, 0, counter.Count.GetTyped(), "Should stay at min 0")
}

// TestUseCounter_WorksWithCreateShared tests shared composable pattern
func TestUseCounter_WorksWithCreateShared(t *testing.T) {
	// Create shared instance
	sharedCounter := CreateShared(func(ctx *bubbly.Context) *CounterReturn {
		return UseCounter(ctx, 0)
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	counter1 := sharedCounter(ctx1)
	counter2 := sharedCounter(ctx2)

	// Both should be the same instance
	counter1.IncrementBy(10)

	assert.Equal(t, 10, counter2.Count.GetTyped(),
		"Shared instance should have same counter state")
}

// TestUseCounter_ValueIsReactive tests that Count ref is reactive
func TestUseCounter_ValueIsReactive(t *testing.T) {
	ctx := createTestContext()
	counter := UseCounter(ctx, 0)

	// Track changes
	changeCount := 0
	bubbly.Watch(counter.Count, func(newVal, oldVal int) {
		changeCount++
	})

	// Increment should trigger watcher
	counter.Increment()
	assert.Equal(t, 1, changeCount, "Increment should trigger watcher")

	// Decrement should trigger watcher
	counter.Decrement()
	assert.Equal(t, 2, changeCount, "Decrement should trigger watcher")

	// Set should trigger watcher
	counter.Set(50)
	assert.Equal(t, 3, changeCount, "Set should trigger watcher")

	// Reset should trigger watcher
	counter.Reset()
	assert.Equal(t, 4, changeCount, "Reset should trigger watcher")
}

// TestUseCounter_MultipleOptions tests combining multiple options
func TestUseCounter_MultipleOptions(t *testing.T) {
	ctx := createTestContext()
	counter := UseCounter(ctx, 50,
		WithMin(0),
		WithMax(100),
		WithStep(5))

	// Verify initial value
	assert.Equal(t, 50, counter.Count.GetTyped(), "Initial should be 50")

	// Increment by step
	counter.Increment()
	assert.Equal(t, 55, counter.Count.GetTyped(), "After increment: 55")

	// Decrement by step
	counter.Decrement()
	assert.Equal(t, 50, counter.Count.GetTyped(), "After decrement: 50")

	// Set to max
	counter.Set(100)
	assert.Equal(t, 100, counter.Count.GetTyped(), "Set to max: 100")

	// Increment should stay at max
	counter.Increment()
	assert.Equal(t, 100, counter.Count.GetTyped(), "Stay at max: 100")

	// Reset to initial
	counter.Reset()
	assert.Equal(t, 50, counter.Count.GetTyped(), "Reset to initial: 50")
}

// TestUseCounter_NoBoundsAllowsAnyValue tests counter without bounds
func TestUseCounter_NoBoundsAllowsAnyValue(t *testing.T) {
	ctx := createTestContext()
	counter := UseCounter(ctx, 0)

	// Should allow very large values
	counter.Set(1000000)
	assert.Equal(t, 1000000, counter.Count.GetTyped(), "Should allow large positive")

	// Should allow very small values
	counter.Set(-1000000)
	assert.Equal(t, -1000000, counter.Count.GetTyped(), "Should allow large negative")
}

// TestUseCounter_IncrementByWithBounds tests IncrementBy respects bounds
func TestUseCounter_IncrementByWithBounds(t *testing.T) {
	ctx := createTestContext()
	counter := UseCounter(ctx, 90, WithMin(0), WithMax(100))

	// IncrementBy should clamp to max
	counter.IncrementBy(50)
	assert.Equal(t, 100, counter.Count.GetTyped(), "IncrementBy should clamp to max")
}

// TestUseCounter_DecrementByWithBounds tests DecrementBy respects bounds
func TestUseCounter_DecrementByWithBounds(t *testing.T) {
	ctx := createTestContext()
	counter := UseCounter(ctx, 10, WithMin(0), WithMax(100))

	// DecrementBy should clamp to min
	counter.DecrementBy(50)
	assert.Equal(t, 0, counter.Count.GetTyped(), "DecrementBy should clamp to min")
}
