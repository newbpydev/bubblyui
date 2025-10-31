package composables

import (
	"sync"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestUseEventListener_RegistersHandler verifies that UseEventListener registers an event handler
func TestUseEventListener_RegistersHandler(t *testing.T) {
	// Arrange
	executed := false
	handler := func() {
		executed = true
	}

	// Act
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			UseEventListener(ctx, "click", handler)
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)
	comp.Init()
	comp.Emit("click", nil)

	// Assert
	assert.True(t, executed, "Handler should be executed when event is emitted")
}

// TestUseEventListener_HandlerExecutesOnEvent verifies handler executes when event is emitted
func TestUseEventListener_HandlerExecutesOnEvent(t *testing.T) {
	// Arrange
	callCount := 0
	handler := func() {
		callCount++
	}

	// Act
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			UseEventListener(ctx, "update", handler)
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)
	comp.Init()
	comp.Emit("update", nil)
	comp.Emit("update", nil)
	comp.Emit("update", nil)

	// Assert
	assert.Equal(t, 3, callCount, "Handler should execute 3 times for 3 events")
}

// TestUseEventListener_ManualCleanup verifies manual cleanup prevents handler execution
func TestUseEventListener_ManualCleanup(t *testing.T) {
	// Arrange
	var cleanup func()
	callCount := 0
	handler := func() {
		callCount++
	}

	// Act
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			cleanup = UseEventListener(ctx, "click", handler)
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)
	comp.Init()
	comp.Emit("click", nil) // Should execute
	cleanup()               // Manual cleanup
	comp.Emit("click", nil) // Should NOT execute

	// Assert
	assert.Equal(t, 1, callCount, "Handler should only execute once before cleanup")
}

// TestUseEventListener_AutoCleanupOnUnmount verifies automatic cleanup on unmount
func TestUseEventListener_AutoCleanupOnUnmount(t *testing.T) {
	// Arrange
	callCount := 0
	handler := func() {
		callCount++
	}

	// Act
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			UseEventListener(ctx, "click", handler)
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)
	comp.Init()
	comp.Emit("click", nil) // Should execute

	// Unmount using type assertion
	if impl, ok := comp.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	comp.Emit("click", nil) // Should NOT execute

	// Assert
	assert.Equal(t, 1, callCount, "Handler should only execute once before unmount")
}

// TestUseEventListener_MultipleListeners verifies multiple listeners work independently
func TestUseEventListener_MultipleListeners(t *testing.T) {
	// Arrange
	var cleanup1, cleanup2 func()
	count1 := 0
	count2 := 0
	handler1 := func() { count1++ }
	handler2 := func() { count2++ }

	// Act
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			cleanup1 = UseEventListener(ctx, "event1", handler1)
			cleanup2 = UseEventListener(ctx, "event2", handler2)
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)
	comp.Init()

	comp.Emit("event1", nil)
	comp.Emit("event2", nil)
	comp.Emit("event1", nil)

	cleanup1() // Cleanup first listener

	comp.Emit("event1", nil) // Should NOT execute handler1
	comp.Emit("event2", nil) // Should execute handler2

	cleanup2() // Cleanup second listener

	comp.Emit("event2", nil) // Should NOT execute handler2

	// Assert
	assert.Equal(t, 2, count1, "Handler1 should execute 2 times before cleanup")
	assert.Equal(t, 2, count2, "Handler2 should execute 2 times before cleanup")
}

// TestUseEventListener_DifferentEvents verifies listeners for different events
func TestUseEventListener_DifferentEvents(t *testing.T) {
	// Arrange
	clickCount := 0
	submitCount := 0

	// Act
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			UseEventListener(ctx, "click", func() { clickCount++ })
			UseEventListener(ctx, "submit", func() { submitCount++ })
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)
	comp.Init()

	comp.Emit("click", nil)
	comp.Emit("submit", nil)
	comp.Emit("click", nil)
	comp.Emit("click", nil)
	comp.Emit("submit", nil)

	// Assert
	assert.Equal(t, 3, clickCount, "Click handler should execute 3 times")
	assert.Equal(t, 2, submitCount, "Submit handler should execute 2 times")
}

// TestUseEventListener_ThreadSafety verifies thread-safe concurrent access
func TestUseEventListener_ThreadSafety(t *testing.T) {
	// Arrange
	var cleanup func()
	var mu sync.Mutex
	callCount := 0
	handler := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	// Act
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			cleanup = UseEventListener(ctx, "concurrent", handler)
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)
	comp.Init()

	// Emit events concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comp.Emit("concurrent", nil)
		}()
	}
	wg.Wait()

	// Cleanup and emit more events
	cleanup()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comp.Emit("concurrent", nil)
		}()
	}
	wg.Wait()

	// Assert
	mu.Lock()
	assert.Equal(t, 10, callCount, "Handler should execute 10 times before cleanup")
	mu.Unlock()
}

// TestUseEventListener_NilContext verifies graceful handling of nil context
func TestUseEventListener_NilContext(t *testing.T) {
	// Arrange
	executed := false
	handler := func() {
		executed = true
	}

	// Act - should not panic
	cleanup := UseEventListener(nil, "click", handler)

	// Assert
	assert.NotNil(t, cleanup, "Should return cleanup function even with nil context")
	assert.False(t, executed, "Handler should not execute without component")
}

// TestUseEventListener_CleanupIdempotent verifies cleanup can be called multiple times
func TestUseEventListener_CleanupIdempotent(t *testing.T) {
	// Arrange
	var cleanup func()
	callCount := 0
	handler := func() {
		callCount++
	}

	// Act
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			cleanup = UseEventListener(ctx, "click", handler)
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)
	comp.Init()
	comp.Emit("click", nil)

	// Call cleanup multiple times - should not panic
	cleanup()
	cleanup()
	cleanup()

	comp.Emit("click", nil)

	// Assert
	assert.Equal(t, 1, callCount, "Handler should only execute once")
}

// TestUseEventListener_WithComponentLifecycle verifies integration with component lifecycle
func TestUseEventListener_WithComponentLifecycle(t *testing.T) {
	// Arrange
	var mountCount int
	comp, err := bubbly.NewComponent("Test").
		Setup(func(ctx *bubbly.Context) {
			UseEventListener(ctx, "mount", func() {
				mountCount++
			})
		}).
		Template(func(rc bubbly.RenderContext) string { return "" }).
		Build()
	assert.NoError(t, err)

	// Act
	comp.Init()
	comp.Emit("mount", nil)
	firstCount := mountCount

	// Unmount using type assertion
	if impl, ok := comp.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	comp.Emit("mount", nil)
	secondCount := mountCount

	// Assert
	assert.Equal(t, 1, firstCount, "Handler should execute before unmount")
	assert.Equal(t, 1, secondCount, "Handler should not execute after unmount")
}
