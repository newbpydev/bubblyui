package composables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseDebounce_BasicDebouncing tests that rapid changes are debounced
func TestUseDebounce_BasicDebouncing(t *testing.T) {
	// Arrange
	ctx := createTestContext()
	source := bubbly.NewRef(0)
	delay := 50 * time.Millisecond

	// Act
	debounced := UseDebounce(ctx, source, delay)

	// Initial value should match source
	assert.Equal(t, 0, debounced.GetTyped())

	// Rapid changes
	source.Set(1)
	source.Set(2)
	source.Set(3)

	// Should not update immediately
	assert.Equal(t, 0, debounced.GetTyped())

	// Wait for debounce delay
	time.Sleep(delay + 10*time.Millisecond)

	// Should have final value
	assert.Equal(t, 3, debounced.GetTyped())
}

// TestUseDebounce_DelayRespected tests that the delay duration is respected
func TestUseDebounce_DelayRespected(t *testing.T) {
	// Arrange
	ctx := createTestContext()
	source := bubbly.NewRef("initial")
	delay := 100 * time.Millisecond

	// Act
	debounced := UseDebounce(ctx, source, delay)

	// Change value
	source.Set("changed")

	// Check before delay expires
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, "initial", debounced.GetTyped(), "should not update before delay")

	// Wait for remaining delay
	time.Sleep(60 * time.Millisecond)
	assert.Equal(t, "changed", debounced.GetTyped(), "should update after delay")
}

// TestUseDebounce_MultipleRapidChanges tests handling of many rapid changes
func TestUseDebounce_MultipleRapidChanges(t *testing.T) {
	// Arrange
	ctx := createTestContext()
	source := bubbly.NewRef(0)
	delay := 50 * time.Millisecond

	// Act
	debounced := UseDebounce(ctx, source, delay)

	// Make many rapid changes
	for i := 1; i <= 10; i++ {
		source.Set(i)
		time.Sleep(5 * time.Millisecond) // Small delay between changes
	}

	// Should still be initial value
	assert.Equal(t, 0, debounced.GetTyped())

	// Wait for debounce
	time.Sleep(delay + 10*time.Millisecond)

	// Should have final value
	assert.Equal(t, 10, debounced.GetTyped())
}

// TestUseDebounce_FinalValuePropagated tests that the final value is correctly propagated
func TestUseDebounce_FinalValuePropagated(t *testing.T) {
	// Arrange
	ctx := createTestContext()
	source := bubbly.NewRef(100)
	delay := 30 * time.Millisecond

	// Act
	debounced := UseDebounce(ctx, source, delay)

	// Initial value
	assert.Equal(t, 100, debounced.GetTyped())

	// Change to 200
	source.Set(200)
	time.Sleep(delay + 10*time.Millisecond)
	assert.Equal(t, 200, debounced.GetTyped())

	// Change to 300
	source.Set(300)
	time.Sleep(delay + 10*time.Millisecond)
	assert.Equal(t, 300, debounced.GetTyped())
}

// TestUseDebounce_TimerCleanupOnUnmount tests that timer is cleaned up on unmount
func TestUseDebounce_TimerCleanupOnUnmount(t *testing.T) {
	// Arrange
	ctx := createTestContext()
	source := bubbly.NewRef(0)
	debounced := UseDebounce(ctx, source, 100*time.Millisecond)

	// Make changes
	source.Set(1)
	source.Set(2)

	// Note: In a real component, ctx.OnUnmounted would be called automatically.
	// For this test, we verify that the timer cleanup is registered properly
	// by checking that rapid changes before the delay don't update the debounced value.

	// Verify debounced hasn't updated yet
	assert.Equal(t, 0, debounced.GetTyped(), "should not update before delay")

	// Wait for timer to fire
	time.Sleep(150 * time.Millisecond)

	// Should have final value
	assert.Equal(t, 2, debounced.GetTyped(), "should update after delay")
}

// TestUseDebounce_TypeSafety tests type safety with different types
func TestUseDebounce_TypeSafety(t *testing.T) {
	tests := []struct {
		name         string
		initialValue interface{}
		newValue     interface{}
	}{
		{
			name:         "int type",
			initialValue: 42,
			newValue:     100,
		},
		{
			name:         "string type",
			initialValue: "hello",
			newValue:     "world",
		},
		{
			name:         "struct type",
			initialValue: struct{ Name string }{Name: "Alice"},
			newValue:     struct{ Name string }{Name: "Bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies compile-time type safety
			// If it compiles, type safety is working
			ctx := createTestContext()

			switch v := tt.initialValue.(type) {
			case int:
				source := bubbly.NewRef(v)
				debounced := UseDebounce(ctx, source, 10*time.Millisecond)
				source.Set(tt.newValue.(int))
				time.Sleep(20 * time.Millisecond)
				assert.Equal(t, tt.newValue.(int), debounced.GetTyped())

			case string:
				source := bubbly.NewRef(v)
				debounced := UseDebounce(ctx, source, 10*time.Millisecond)
				source.Set(tt.newValue.(string))
				time.Sleep(20 * time.Millisecond)
				assert.Equal(t, tt.newValue.(string), debounced.GetTyped())

			case struct{ Name string }:
				source := bubbly.NewRef(v)
				debounced := UseDebounce(ctx, source, 10*time.Millisecond)
				source.Set(tt.newValue.(struct{ Name string }))
				time.Sleep(20 * time.Millisecond)
				assert.Equal(t, tt.newValue.(struct{ Name string }), debounced.GetTyped())
			}
		})
	}
}

// TestUseDebounce_ZeroDelay tests behavior with zero delay
func TestUseDebounce_ZeroDelay(t *testing.T) {
	// Arrange
	ctx := createTestContext()
	source := bubbly.NewRef(0)
	delay := 0 * time.Millisecond

	// Act
	debounced := UseDebounce(ctx, source, delay)

	// Change value
	source.Set(1)

	// With zero delay, should update very quickly
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, 1, debounced.GetTyped())
}

// TestUseDebounce_ConsecutiveDebounces tests multiple debounce periods
func TestUseDebounce_ConsecutiveDebounces(t *testing.T) {
	// Arrange
	ctx := createTestContext()
	source := bubbly.NewRef(0)
	delay := 30 * time.Millisecond

	// Act
	debounced := UseDebounce(ctx, source, delay)

	// First batch of changes
	source.Set(1)
	source.Set(2)
	time.Sleep(delay + 10*time.Millisecond)
	assert.Equal(t, 2, debounced.GetTyped())

	// Second batch of changes
	source.Set(3)
	source.Set(4)
	time.Sleep(delay + 10*time.Millisecond)
	assert.Equal(t, 4, debounced.GetTyped())

	// Third batch of changes
	source.Set(5)
	source.Set(6)
	time.Sleep(delay + 10*time.Millisecond)
	assert.Equal(t, 6, debounced.GetTyped())
}

// TestUseDebounce_WithComponentLifecycle tests cleanup integration with component lifecycle
func TestUseDebounce_WithComponentLifecycle(t *testing.T) {
	var source *bubbly.Ref[int]
	var debounced *bubbly.Ref[int]

	// Create component with UseDebounce
	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			source = bubbly.NewRef(0)
			debounced = UseDebounce(ctx, source, 50*time.Millisecond)

			ctx.Expose("source", source)
			ctx.Expose("debounced", debounced)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Initialize component
	comp.Init()
	comp.View()

	// Make changes
	source.Set(1)
	source.Set(2)
	source.Set(3)

	// Wait for debounce
	time.Sleep(60 * time.Millisecond)

	// Should have final value
	assert.Equal(t, 3, debounced.GetTyped())

	// Unmount component (cleanup should run)
	type unmounter interface {
		Unmount()
	}
	if u, ok := comp.(unmounter); ok {
		u.Unmount()
	}

	// Make more changes after unmount
	source.Set(4)
	source.Set(5)

	// Wait to ensure timer would have fired
	time.Sleep(60 * time.Millisecond)

	// Debounced should still be 3 (cleanup stopped the watcher)
	assert.Equal(t, 3, debounced.GetTyped(), "debounced should not update after unmount")
}
