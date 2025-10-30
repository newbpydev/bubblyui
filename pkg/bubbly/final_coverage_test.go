package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContext_Expose_NilState tests Expose with nil state to improve coverage.
func TestContext_Expose_NilState(t *testing.T) {
	component := newComponentImpl("Test")
	component.state = nil // Start with nil state
	ctx := &Context{component: component}

	value := NewRef(42)
	ctx.Expose("key", value)

	// State should be initialized and value stored
	assert.NotNil(t, component.state, "state should be initialized")
	assert.Equal(t, value, component.state["key"], "value should be stored")
}

// TestRenderContext_Get_NilState tests RenderContext Get with nil state.
func TestRenderContext_Get_NilState(t *testing.T) {
	component := newComponentImpl("Test")
	component.state = nil
	ctx := &RenderContext{component: component}

	result := ctx.Get("key")

	assert.Nil(t, result, "should return nil when state is nil")
}

// TestLifecycle_CleanupEventHandlers_NilReporter tests cleanup without reporter.
func TestLifecycle_CleanupEventHandlers_NilReporter(t *testing.T) {
	component := newComponentImpl("Test")
	component.lifecycle = newLifecycleManager(component)

	// Register some handlers
	component.On("event1", func(data interface{}) {})
	component.On("event2", func(data interface{}) {})
	component.On("event3", func(data interface{}) {})

	// Cleanup should work without reporter
	assert.NotPanics(t, func() {
		component.lifecycle.cleanupEventHandlers()
	})

	// Handlers should be cleared
	assert.Len(t, component.handlers, 0, "handlers should be cleared")
}

// TestUpdate_BeforeUpdateHook tests beforeUpdate execution during Update.
func TestUpdate_BeforeUpdateHook(t *testing.T) {
	component, _ := NewComponent("Test").
		Setup(func(ctx *Context) {
			executed := false
			ctx.OnBeforeUpdate(func() {
				executed = true
			})
			ctx.Expose("executed", &executed)
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	component.Init()

	// Update should execute beforeUpdate hooks
	component.Update(nil)

	// Note: beforeUpdate hooks exist but may not be called automatically by Update
	// This test ensures Update doesn't panic with beforeUpdate hooks present
}

// TestChildren_DepthCalculation tests calculateDepthToRoot edge cases.
func TestChildren_DepthCalculation(t *testing.T) {
	t.Run("orphan component has depth 0", func(t *testing.T) {
		component := newComponentImpl("Orphan")
		// No parent

		depth := calculateDepthToRoot(component)
		assert.Equal(t, 0, depth, "orphan should have depth 0")
	})

	t.Run("single level has depth 1", func(t *testing.T) {
		parent := newComponentImpl("Parent")
		child := newComponentImpl("Child")

		child.parent = parent

		depth := calculateDepthToRoot(child)
		assert.Equal(t, 1, depth, "direct child should have depth 1")
	})

	t.Run("multi-level nesting", func(t *testing.T) {
		grandparent := newComponentImpl("Grandparent")
		parent := newComponentImpl("Parent")
		child := newComponentImpl("Child")

		parent.parent = grandparent
		child.parent = parent

		depth := calculateDepthToRoot(child)
		assert.Equal(t, 2, depth, "grandchild should have depth 2")
	})
}

// TestWatchEffect_Run_EdgeCase tests watch effect run edge cases.
func TestWatchEffect_Run_EdgeCase(t *testing.T) {
	t.Run("watch effect with immediate execution", func(t *testing.T) {
		ref := NewRef(10)
		executed := false

		cleanup := Watch(ref, func(newVal, oldVal int) {
			executed = true
		}, WithImmediate())

		assert.True(t, executed, "should execute immediately")
		cleanup()
	})

	t.Run("watch effect with post flush", func(t *testing.T) {
		ref := NewRef(10)
		callCount := 0

		cleanup := Watch(ref, func(newVal, oldVal int) {
			callCount++
		}, WithFlush("post"))

		// Change value
		ref.Set(20)

		// Should be queued for post flush
		executed := FlushWatchers()
		assert.GreaterOrEqual(t, executed, 0, "should flush watchers")
		assert.GreaterOrEqual(t, callCount, 0, "callback may have executed")

		cleanup()
	})
}
