package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContext_OnBeforeUpdate tests OnBeforeUpdate hook registration.
// This was previously untested (0% coverage).
func TestContext_OnBeforeUpdate(t *testing.T) {
	t.Run("register and execute onBeforeUpdate hook", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		executed := false
		ctx.OnBeforeUpdate(func() {
			executed = true
		})

		// Verify hook was registered
		assert.Len(t, component.lifecycle.hooks["beforeUpdate"], 1)

		// Execute the hook via the generic executeHooks
		component.lifecycle.executeHooks("beforeUpdate")

		// Verify it executed
		assert.True(t, executed, "onBeforeUpdate hook should execute")
	})

	t.Run("multiple onBeforeUpdate hooks execute in order", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		executionOrder := []int{}

		ctx.OnBeforeUpdate(func() {
			executionOrder = append(executionOrder, 1)
		})
		ctx.OnBeforeUpdate(func() {
			executionOrder = append(executionOrder, 2)
		})
		ctx.OnBeforeUpdate(func() {
			executionOrder = append(executionOrder, 3)
		})

		component.lifecycle.executeHooks("beforeUpdate")

		assert.Equal(t, []int{1, 2, 3}, executionOrder, "hooks should execute in registration order")
	})

	t.Run("onBeforeUpdate with panic recovery", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		executed := []bool{false, false, false}

		ctx.OnBeforeUpdate(func() {
			executed[0] = true
		})
		ctx.OnBeforeUpdate(func() {
			executed[1] = true
			panic("test panic")
		})
		ctx.OnBeforeUpdate(func() {
			executed[2] = true
		})

		// Should not panic
		assert.NotPanics(t, func() {
			component.lifecycle.executeHooks("beforeUpdate")
		})

		// All hooks should execute (panic recovered)
		assert.True(t, executed[0], "first hook should execute")
		assert.True(t, executed[1], "second hook should execute before panic")
		assert.True(t, executed[2], "third hook should execute after panic recovery")
	})
}

// TestContext_OnBeforeUnmount tests OnBeforeUnmount hook registration.
// This was previously untested (0% coverage).
func TestContext_OnBeforeUnmount(t *testing.T) {
	t.Run("register and execute onBeforeUnmount hook", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		executed := false
		ctx.OnBeforeUnmount(func() {
			executed = true
		})

		// Verify hook was registered
		assert.Len(t, component.lifecycle.hooks["beforeUnmount"], 1)

		// Execute the hook via the generic executeHooks
		component.lifecycle.executeHooks("beforeUnmount")

		// Verify it executed
		assert.True(t, executed, "onBeforeUnmount hook should execute")
	})

	t.Run("multiple onBeforeUnmount hooks execute in order", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		executionOrder := []int{}

		ctx.OnBeforeUnmount(func() {
			executionOrder = append(executionOrder, 1)
		})
		ctx.OnBeforeUnmount(func() {
			executionOrder = append(executionOrder, 2)
		})
		ctx.OnBeforeUnmount(func() {
			executionOrder = append(executionOrder, 3)
		})

		component.lifecycle.executeHooks("beforeUnmount")

		assert.Equal(t, []int{1, 2, 3}, executionOrder, "hooks should execute in registration order")
	})

	t.Run("onBeforeUnmount with panic recovery", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		executed := []bool{false, false, false}

		ctx.OnBeforeUnmount(func() {
			executed[0] = true
		})
		ctx.OnBeforeUnmount(func() {
			executed[1] = true
			panic("test panic")
		})
		ctx.OnBeforeUnmount(func() {
			executed[2] = true
		})

		// Should not panic
		assert.NotPanics(t, func() {
			component.lifecycle.executeHooks("beforeUnmount")
		})

		// All hooks should execute (panic recovered)
		assert.True(t, executed[0], "first hook should execute")
		assert.True(t, executed[1], "second hook should execute before panic")
		assert.True(t, executed[2], "third hook should execute after panic recovery")
	})

	t.Run("onBeforeUnmount hook is registered correctly", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		hookRegistered := false

		ctx.OnBeforeUnmount(func() {
			hookRegistered = true
		})

		// Verify hook is registered
		assert.Len(t, component.lifecycle.hooks["beforeUnmount"], 1, "hook should be registered")

		// Execute the hook directly to verify it works
		component.lifecycle.executeHooks("beforeUnmount")

		assert.True(t, hookRegistered, "beforeUnmount hook should execute when called")
	})
}

// TestEvent_StopPropagation tests the StopPropagation method.
// This was previously untested (0% coverage).
func TestEvent_StopPropagation(t *testing.T) {
	t.Run("StopPropagation sets Stopped flag", func(t *testing.T) {
		event := &Event{
			Name:    "test",
			Stopped: false,
		}

		assert.False(t, event.Stopped, "event should not be stopped initially")

		event.StopPropagation()

		assert.True(t, event.Stopped, "event should be stopped after calling StopPropagation")
	})

	t.Run("StopPropagation can be called multiple times", func(t *testing.T) {
		event := &Event{Name: "test"}

		event.StopPropagation()
		assert.True(t, event.Stopped)

		event.StopPropagation() // Should not cause issues
		assert.True(t, event.Stopped)
	})
}

// TestContext_Expose_EdgeCases tests Expose edge cases to improve coverage from 66.7%.
func TestContext_Expose_EdgeCases(t *testing.T) {
	t.Run("Expose stores value in state", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		ctx := &Context{component: component}

		value := NewRef(42)
		ctx.Expose("value", value)

		// Value should be stored in state
		assert.Equal(t, value, component.state["value"], "value should be stored")
	})

	t.Run("Expose overwrites existing value", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		value1 := NewRef(1)
		value2 := NewRef(2)

		ctx.Expose("key", value1)
		assert.Equal(t, value1, component.state["key"])

		ctx.Expose("key", value2)
		assert.Equal(t, value2, component.state["key"], "should overwrite")
	})

	t.Run("Expose with empty key", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		value := NewRef(42)
		ctx.Expose("", value)

		assert.Equal(t, value, component.state[""], "should accept empty key")
	})
}

// TestContext_Get_EdgeCases tests Get edge cases to improve coverage from 66.7%.
func TestContext_Get_EdgeCases(t *testing.T) {
	t.Run("Get returns value from state", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		value := NewRef(42)
		component.state["value"] = value
		ctx := &Context{component: component}

		result := ctx.Get("value")

		assert.Equal(t, value, result, "should return the value from state")
	})

	t.Run("Get non-existent key returns nil", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		result := ctx.Get("nonexistent")

		assert.Nil(t, result, "should return nil for non-existent key")
	})

	t.Run("Get with empty key", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		component.state[""] = NewRef(42)
		ctx := &Context{component: component}

		result := ctx.Get("")

		assert.NotNil(t, result, "should handle empty key")
	})
}

// TestRenderContext_Get_EdgeCases tests RenderContext Get edge cases to improve coverage from 66.7%.
func TestRenderContext_Get_EdgeCases(t *testing.T) {
	t.Run("Get non-existent key returns nil", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		ctx := &RenderContext{component: component}

		result := ctx.Get("nonexistent")

		assert.Nil(t, result, "should return nil for non-existent key")
	})

	t.Run("Get with empty key", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.state[""] = NewRef(99)
		ctx := &RenderContext{component: component}

		result := ctx.Get("")

		assert.NotNil(t, result, "should handle empty key")
	})

	t.Run("Get after state modification", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		value := NewRef(1)
		component.state["key"] = value
		ctx := &RenderContext{component: component}

		result1 := ctx.Get("key")
		value.Set(2)
		result2 := ctx.Get("key")

		assert.Equal(t, result1, result2, "should return same reference")
	})
}

// TestContext_OnCleanup_EdgeCases tests OnCleanup edge cases to improve coverage from 66.7%.
func TestContext_OnCleanup_EdgeCases(t *testing.T) {
	t.Run("OnCleanup with nil lifecycle creates lifecycle", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = nil
		ctx := &Context{component: component}

		executed := false
		ctx.OnCleanup(func() {
			executed = true
		})

		assert.NotNil(t, component.lifecycle, "lifecycle should be created")
		assert.Len(t, component.lifecycle.cleanups, 1, "cleanup should be registered")

		component.lifecycle.executeCleanups()
		assert.True(t, executed, "cleanup should execute")
	})

	t.Run("OnCleanup with multiple cleanups", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		executionOrder := []int{}

		ctx.OnCleanup(func() {
			executionOrder = append(executionOrder, 1)
		})
		ctx.OnCleanup(func() {
			executionOrder = append(executionOrder, 2)
		})
		ctx.OnCleanup(func() {
			executionOrder = append(executionOrder, 3)
		})

		component.lifecycle.executeCleanups()

		// Cleanups execute in LIFO order (reverse)
		assert.Equal(t, []int{3, 2, 1}, executionOrder, "cleanups should execute in LIFO order")
	})

	t.Run("OnCleanup with panic recovery", func(t *testing.T) {
		component := newComponentImpl("TestComponent")
		component.lifecycle = newLifecycleManager(component)
		ctx := &Context{component: component}

		executed := []bool{false, false, false}

		ctx.OnCleanup(func() {
			executed[0] = true
		})
		ctx.OnCleanup(func() {
			executed[1] = true
			panic("test panic")
		})
		ctx.OnCleanup(func() {
			executed[2] = true
		})

		// Should not panic
		assert.NotPanics(t, func() {
			component.lifecycle.executeCleanups()
		})

		// All should execute (LIFO order: 3, 2, 1)
		assert.True(t, executed[0], "first cleanup should execute")
		assert.True(t, executed[1], "second cleanup should execute")
		assert.True(t, executed[2], "third cleanup should execute")
	})
}
