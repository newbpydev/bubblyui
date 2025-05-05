package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookContext(t *testing.T) {
	t.Run("Context Creation and Access", func(t *testing.T) {
		// Create a new hook context
		ctx := NewHookContext()
		assert.NotNil(t, ctx, "HookContext should not be nil")

		// Set and get values
		ctx.Set("stringKey", "string value")
		ctx.Set("intKey", 42)
		ctx.Set("boolKey", true)

		// Test value retrieval
		val, exists := ctx.Get("stringKey")
		assert.True(t, exists, "Key should exist")
		assert.Equal(t, "string value", val, "Value should match")

		val, exists = ctx.Get("intKey")
		assert.True(t, exists, "Key should exist")
		assert.Equal(t, 42, val, "Value should match")

		val, exists = ctx.Get("nonExistentKey")
		assert.False(t, exists, "Key should not exist")
		assert.Nil(t, val, "Value should be nil for non-existent key")
	})

	t.Run("Context Inheritance", func(t *testing.T) {
		// Create parent context
		parentCtx := NewHookContext()
		parentCtx.Set("parentKey", "parent value")

		// Create child context with parent
		childCtx := NewHookContextWithParent(parentCtx)
		childCtx.Set("childKey", "child value")

		// Child should access its own values
		val, exists := childCtx.Get("childKey")
		assert.True(t, exists, "Child key should exist in child context")
		assert.Equal(t, "child value", val, "Child value should match")

		// Child should access parent values
		val, exists = childCtx.Get("parentKey")
		assert.True(t, exists, "Parent key should be accessible from child context")
		assert.Equal(t, "parent value", val, "Parent value should match")

		// Parent should not access child values
		val, exists = parentCtx.Get("childKey")
		assert.False(t, exists, "Child key should not be accessible from parent context")
	})
}

func TestHookContextIntegration(t *testing.T) {
	t.Run("Hook Manager With Context", func(t *testing.T) {
		// Create hook manager
		hm := NewHookManager("test-component")

		// Create and set context
		ctx := NewHookContext()
		ctx.Set("componentType", "test")
		hm.SetContext(ctx)

		// Get context from manager
		retrievedCtx := hm.GetContext()
		assert.NotNil(t, retrievedCtx, "Retrieved context should not be nil")

		// Check value in retrieved context
		val, exists := retrievedCtx.Get("componentType")
		assert.True(t, exists, "Key should exist in retrieved context")
		assert.Equal(t, "test", val, "Value should match")
	})

	t.Run("Context Propagation Between Managers", func(t *testing.T) {
		// Create parent hook manager with context
		parentHM := NewHookManager("parent")
		parentCtx := NewHookContext()
		parentCtx.Set("level", "parent")
		parentHM.SetContext(parentCtx)

		// Create child hook manager
		childHM := NewHookManager("child")
		childHM.SetParentManager(parentHM)

		// Child should access parent context
		childCtx := childHM.GetContext()
		val, exists := childCtx.Get("level")
		assert.True(t, exists, "Parent context key should be accessible from child")
		assert.Equal(t, "parent", val, "Value should match")

		// Test context override in child
		childCtx.Set("level", "child")
		val, exists = childCtx.Get("level")
		assert.True(t, exists, "Key should exist in child context")
		assert.Equal(t, "child", val, "Child value should override parent value")

		// Parent context should remain unchanged
		val, exists = parentCtx.Get("level")
		assert.True(t, exists, "Key should still exist in parent context")
		assert.Equal(t, "parent", val, "Parent value should remain unchanged")
	})
}
