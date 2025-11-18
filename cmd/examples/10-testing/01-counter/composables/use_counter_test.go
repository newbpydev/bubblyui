package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// NOTE: This file demonstrates TWO testing approaches for composables:
// 1. Direct unit testing (for custom composables like UseCounter)
// 2. Component-wrapper testing (using testutil patterns)

// ============================================================================
// APPROACH 1: Direct Unit Testing (for custom composables)
// ============================================================================

// createTestContext creates a minimal context for direct composable testing
func createTestContext() *bubbly.Context {
	var ctx *bubbly.Context
	component, _ := bubbly.NewComponent("Test").
		Setup(func(c *bubbly.Context) {
			ctx = c
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()
	// CRITICAL: Must call Init() to execute Setup and get the context
	component.Init()
	return ctx
}

// ============================================================================
// APPROACH 2: Component-Wrapper Testing (using testutil patterns)
// ============================================================================

// createCounterComponent wraps UseCounter in a component for testutil-based testing
// This demonstrates how to test composables using BubblyUI's testutil patterns
func createCounterComponent(initial int) (bubbly.Component, error) {
	return bubbly.NewComponent("CounterTest").
		Setup(func(ctx *bubbly.Context) {
			counter := UseCounter(ctx, initial)

			// Expose all counter fields for testing
			ctx.Expose("count", counter.Count)
			ctx.Expose("history", counter.History)
			ctx.Expose("doubled", counter.Doubled)
			ctx.Expose("isEven", counter.IsEven)
			ctx.Expose("increment", counter.Increment)
			ctx.Expose("decrement", counter.Decrement)
			ctx.Expose("reset", counter.Reset)
			ctx.Expose("setValue", counter.SetValue)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
}

// TestUseCounter_CreatesWithInitialValue demonstrates basic composable creation
// Shows: Composable initialization, initial state verification
func TestUseCounter_CreatesWithInitialValue(t *testing.T) {
	tests := []struct {
		name    string
		initial int
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := createTestContext()

			// Act
			counter := UseCounter(ctx, tt.initial)

			// Assert
			require.NotNil(t, counter, "Counter should not be nil")
			assert.NotNil(t, counter.Count, "Count ref should not be nil")
			assert.Equal(t, tt.initial, counter.Count.Get(), "Initial count should match")
		})
	}
}

// TestUseCounter_Increment demonstrates increment functionality
// Shows: Testing composable methods, state mutations
func TestUseCounter_Increment(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		times    int
		expected int
	}{
		{"single increment from zero", 0, 1, 1},
		{"multiple increments", 5, 3, 8},
		{"increment from negative", -5, 2, -3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial)

			// Act
			for i := 0; i < tt.times; i++ {
				counter.Increment()
			}

			// Assert
			assert.Equal(t, tt.expected, counter.Count.Get(), "Count should match expected")
		})
	}
}

// TestUseCounter_Decrement demonstrates decrement functionality
func TestUseCounter_Decrement(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		times    int
		expected int
	}{
		{"single decrement from positive", 10, 1, 9},
		{"multiple decrements", 5, 3, 2},
		{"decrement to negative", 0, 2, -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial)

			// Act
			for i := 0; i < tt.times; i++ {
				counter.Decrement()
			}

			// Assert
			assert.Equal(t, tt.expected, counter.Count.Get())
		})
	}
}

// TestUseCounter_Reset demonstrates reset functionality
func TestUseCounter_Reset(t *testing.T) {
	tests := []struct {
		name    string
		initial int
		changes []func(*CounterComposable)
	}{
		{
			name:    "reset after increment",
			initial: 0,
			changes: []func(*CounterComposable){
				func(c *CounterComposable) { c.Increment() },
				func(c *CounterComposable) { c.Increment() },
			},
		},
		{
			name:    "reset after decrement",
			initial: 10,
			changes: []func(*CounterComposable){
				func(c *CounterComposable) { c.Decrement() },
				func(c *CounterComposable) { c.Decrement() },
			},
		},
		{
			name:    "reset after mixed operations",
			initial: 5,
			changes: []func(*CounterComposable){
				func(c *CounterComposable) { c.Increment() },
				func(c *CounterComposable) { c.Decrement() },
				func(c *CounterComposable) { c.Increment() },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial)

			// Act: Make changes
			for _, change := range tt.changes {
				change(counter)
			}

			// Act: Reset
			counter.Reset()

			// Assert: Should be back to initial
			assert.Equal(t, tt.initial, counter.Count.Get(), "Reset should restore initial value")
		})
	}
}

// TestUseCounter_SetValue demonstrates setValue functionality
func TestUseCounter_SetValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		setValue int
	}{
		{"set to positive", 0, 100},
		{"set to negative", 10, -5},
		{"set to zero", 42, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.initial)

			// Act
			counter.SetValue(tt.setValue)

			// Assert
			assert.Equal(t, tt.setValue, counter.Count.Get())
		})
	}
}

// TestUseCounter_Doubled demonstrates computed value
func TestUseCounter_Doubled(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"zero", 0, 0},
		{"positive", 5, 10},
		{"negative", -3, -6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.value)

			// Assert
			assert.Equal(t, tt.expected, counter.Doubled.Get().(int))
		})
	}
}

// TestUseCounter_IsEven demonstrates computed boolean
func TestUseCounter_IsEven(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected bool
	}{
		{"zero is even", 0, true},
		{"positive even", 4, true},
		{"positive odd", 5, false},
		{"negative even", -2, true},
		{"negative odd", -3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := createTestContext()
			counter := UseCounter(ctx, tt.value)

			// Assert
			assert.Equal(t, tt.expected, counter.IsEven.Get().(bool))
		})
	}
}

// TestUseCounter_History demonstrates history tracking
func TestUseCounter_History(t *testing.T) {
	t.Run("initial value in history", func(t *testing.T) {
		// Arrange
		ctx := createTestContext()
		counter := UseCounter(ctx, 5)

		// Assert
		history := counter.History.Get().([]int)
		assert.Equal(t, []int{5}, history, "History should contain initial value")
	})

	t.Run("history tracks increments", func(t *testing.T) {
		// Arrange
		ctx := createTestContext()
		counter := UseCounter(ctx, 0)

		// Act
		counter.Increment() // 0 -> 1
		counter.Increment() // 1 -> 2

		// Assert
		history := counter.History.Get().([]int)
		assert.Contains(t, history, 1, "History should contain 1")
		assert.Contains(t, history, 2, "History should contain 2")
	})

	t.Run("history limited to 5 values", func(t *testing.T) {
		// Arrange
		ctx := createTestContext()
		counter := UseCounter(ctx, 0)

		// Act: Perform many increments
		for i := 0; i < 10; i++ {
			counter.Increment()
		}

		// Assert
		history := counter.History.Get().([]int)
		assert.LessOrEqual(t, len(history), 5, "History should be limited to 5 values")
	})
}

// Note: Concurrent access test removed - BubblyUI composables are not designed
// to be thread-safe. They should be used within the single-threaded
// Bubbletea model Update loop.

// ============================================================================
// Component-Wrapper Tests (Using BubblyUI testutil)
// These tests demonstrate the recommended pattern for testing composables
// within components, using the full testutil harness.
// ============================================================================

// TestUseCounter_WithTestutil_BasicMounting demonstrates testutil-based testing
// Shows: Component harness, state inspection via ComponentTest
func TestUseCounter_WithTestutil_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	component, err := createCounterComponent(5)
	require.NoError(t, err)

	// Act: Mount component
	ct := harness.Mount(component)

	// Assert: Use ComponentTest.State() to inspect state
	ct.AssertRefEquals("count", 5)
}

// TestUseCounter_WithTestutil_StateInspection demonstrates state inspection via testutil
// Shows: Accessing refs through ComponentTest.State() API
func TestUseCounter_WithTestutil_StateInspection(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	component, err := createCounterComponent(0)
	require.NoError(t, err)

	ct := harness.Mount(component)

	// Assert: Use AssertRefEquals for type-safe assertions
	ct.AssertRefEquals("count", 0)

	// Access count ref and modify it directly
	countRef := ct.State().GetRef("count")
	countRef.Set(5)

	// Verify state updated
	ct.AssertRefEquals("count", 5)

	// Access history ref
	historyRef := ct.State().GetRef("history")
	history := historyRef.Get().([]int)
	assert.Contains(t, history, 0, "History should contain initial value")
}

// TestUseCounter_WithTestutil_AssertRefChanged demonstrates change tracking
// Shows: Using AssertRefChanged to verify state mutations
func TestUseCounter_WithTestutil_AssertRefChanged(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	component, err := createCounterComponent(0)
	require.NoError(t, err)

	ct := harness.Mount(component)

	// Capture initial value
	initialCount := ct.State().GetRefValue("count")

	// Modify state
	countRef := ct.State().GetRef("count")
	countRef.Set(10)

	// Assert: State changed from initial
	ct.AssertRefChanged("count", initialCount)
	ct.AssertRefEquals("count", 10)
}
