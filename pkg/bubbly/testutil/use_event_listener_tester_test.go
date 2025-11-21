package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/stretchr/testify/assert"
)

// TestUseEventListenerTester_BasicListener tests basic event listening
func TestUseEventListenerTester_BasicListener(t *testing.T) {
	callCount := 0

	comp, err := bubbly.NewComponent("TestEventListener").
		Setup(func(ctx *bubbly.Context) {
			cleanup := composables.UseEventListener(ctx, "click", func() {
				callCount++
			})

			ctx.Expose("callCount", &callCount)
			ctx.Expose("cleanup", cleanup)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseEventListenerTester(comp)

	// Initially not called
	assert.Equal(t, 0, callCount)

	// Emit event
	tester.EmitEvent("click", nil)
	assert.Equal(t, 1, callCount)

	// Emit again
	tester.EmitEvent("click", nil)
	assert.Equal(t, 2, callCount)
}

// TestUseEventListenerTester_ManualCleanup tests manual cleanup
func TestUseEventListenerTester_ManualCleanup(t *testing.T) {
	callCount := 0

	comp, err := bubbly.NewComponent("TestEventListener").
		Setup(func(ctx *bubbly.Context) {
			cleanup := composables.UseEventListener(ctx, "click", func() {
				callCount++
			})

			ctx.Expose("callCount", &callCount)
			ctx.Expose("cleanup", cleanup)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseEventListenerTester(comp)

	// Emit event
	tester.EmitEvent("click", nil)
	assert.Equal(t, 1, callCount)

	// Manual cleanup
	tester.TriggerCleanup()

	// Emit after cleanup - should not execute
	tester.EmitEvent("click", nil)
	assert.Equal(t, 1, callCount)
}

// TestUseEventListenerTester_MultipleEvents tests listening to multiple event types
func TestUseEventListenerTester_MultipleEvents(t *testing.T) {
	clickCount := 0
	submitCount := 0

	comp, err := bubbly.NewComponent("TestEventListener").
		Setup(func(ctx *bubbly.Context) {
			// Register click listener
			composables.UseEventListener(ctx, "click", func() {
				clickCount++
			})

			// Register submit listener (separate call)
			cleanup := composables.UseEventListener(ctx, "submit", func() {
				submitCount++
			})

			ctx.Expose("clickCount", &clickCount)
			ctx.Expose("submitCount", &submitCount)
			ctx.Expose("cleanup", cleanup) // Expose one cleanup for testing
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseEventListenerTester(comp)

	// Emit click
	tester.EmitEvent("click", nil)
	assert.Equal(t, 1, clickCount)
	assert.Equal(t, 0, submitCount)

	// Emit submit
	tester.EmitEvent("submit", nil)
	assert.Equal(t, 1, clickCount)
	assert.Equal(t, 1, submitCount)

	// Emit both
	tester.EmitEvent("click", nil)
	tester.EmitEvent("submit", nil)
	assert.Equal(t, 2, clickCount)
	assert.Equal(t, 2, submitCount)
}

// TestUseEventListenerTester_GetCallCount tests call count tracking
func TestUseEventListenerTester_GetCallCount(t *testing.T) {
	callCount := 0

	comp, err := bubbly.NewComponent("TestEventListener").
		Setup(func(ctx *bubbly.Context) {
			cleanup := composables.UseEventListener(ctx, "test", func() {
				callCount++
			})

			ctx.Expose("callCount", &callCount)
			ctx.Expose("cleanup", cleanup)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseEventListenerTester(comp)

	// Get call count via tester
	count := tester.GetCallCount("callCount")
	assert.Equal(t, 0, count)

	// Emit events
	tester.EmitEvent("test", nil)
	tester.EmitEvent("test", nil)
	tester.EmitEvent("test", nil)

	// Check updated count
	count = tester.GetCallCount("callCount")
	assert.Equal(t, 3, count)
}

// TestUseEventListenerTester_EventWithData tests event with data payload
func TestUseEventListenerTester_EventWithData(t *testing.T) {
	// Note: UseEventListener doesn't pass data to handler (it's func())
	// This test verifies that emitting with data doesn't break anything
	callCount := 0

	comp, err := bubbly.NewComponent("TestEventListener").
		Setup(func(ctx *bubbly.Context) {
			cleanup := composables.UseEventListener(ctx, "click", func() {
				callCount++
			})

			ctx.Expose("callCount", &callCount)
			ctx.Expose("cleanup", cleanup)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseEventListenerTester(comp)

	// Emit with data (data is ignored by UseEventListener)
	tester.EmitEvent("click", "some data")
	assert.Equal(t, 1, callCount)

	tester.EmitEvent("click", map[string]interface{}{"key": "value"})
	assert.Equal(t, 2, callCount)
}

// TestUseEventListenerTester_MissingRefs tests panic when required refs not exposed
func TestUseEventListenerTester_MissingRefs(t *testing.T) {
	comp, err := bubbly.NewComponent("TestEventListener").
		Setup(func(ctx *bubbly.Context) {
			// Don't expose cleanup function
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	assert.Panics(t, func() {
		NewUseEventListenerTester(comp)
	})
}
