package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewProvideKey tests that NewProvideKey creates a key with the correct string
func TestNewProvideKey(t *testing.T) {
	tests := []struct {
		name     string
		keyStr   string
		typeDesc string
	}{
		{
			name:     "string key",
			keyStr:   "theme",
			typeDesc: "string type",
		},
		{
			name:     "int key",
			keyStr:   "count",
			typeDesc: "int type",
		},
		{
			name:     "complex key name",
			keyStr:   "app.config.database.url",
			typeDesc: "string type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act - create different typed keys
			stringKey := NewProvideKey[string](tt.keyStr)
			intKey := NewProvideKey[int](tt.keyStr)

			// Assert - verify internal key is set correctly
			assert.Equal(t, tt.keyStr, stringKey.key, "String key should match")
			assert.Equal(t, tt.keyStr, intKey.key, "Int key should match")
		})
	}
}

// TestProvideTyped tests that ProvideTyped stores typed values
func TestProvideTyped(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Context) (string, interface{})
	}{
		{
			name: "provide string value",
			setup: func(ctx *Context) (string, interface{}) {
				key := NewProvideKey[string]("theme")
				ProvideTyped(ctx, key, "dark")
				return "theme", "dark"
			},
		},
		{
			name: "provide int value",
			setup: func(ctx *Context) (string, interface{}) {
				key := NewProvideKey[int]("count")
				ProvideTyped(ctx, key, 42)
				return "count", 42
			},
		},
		{
			name: "provide bool value",
			setup: func(ctx *Context) (string, interface{}) {
				key := NewProvideKey[bool]("enabled")
				ProvideTyped(ctx, key, true)
				return "enabled", true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			ctx := &Context{component: c}

			// Act
			keyStr, expectedValue := tt.setup(ctx)

			// Assert - verify value is in provides map
			c.providesMu.RLock()
			stored, exists := c.provides[keyStr]
			c.providesMu.RUnlock()

			assert.True(t, exists, "Key should exist in provides map")
			assert.Equal(t, expectedValue, stored, "Stored value should match")
		})
	}
}

// TestInjectTyped_FromSelf tests injecting from same component
func TestInjectTyped_FromSelf(t *testing.T) {
	tests := []struct {
		name         string
		provideValue interface{}
		defaultValue interface{}
		keyName      string
	}{
		{
			name:         "inject string",
			provideValue: "dark",
			defaultValue: "light",
			keyName:      "theme",
		},
		{
			name:         "inject int",
			provideValue: 100,
			defaultValue: 0,
			keyName:      "count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			ctx := &Context{component: c}

			// Act - provide and inject based on type
			switch v := tt.provideValue.(type) {
			case string:
				key := NewProvideKey[string](tt.keyName)
				ProvideTyped(ctx, key, v)
				result := InjectTyped(ctx, key, tt.defaultValue.(string))
				// Assert
				assert.Equal(t, v, result, "Should return provided value")
			case int:
				key := NewProvideKey[int](tt.keyName)
				ProvideTyped(ctx, key, v)
				result := InjectTyped(ctx, key, tt.defaultValue.(int))
				// Assert
				assert.Equal(t, v, result, "Should return provided value")
			}
		})
	}
}

// TestInjectTyped_NotFound tests inject with no provider returns default
func TestInjectTyped_NotFound(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue interface{}
		keyName      string
	}{
		{
			name:         "default string",
			defaultValue: "light",
			keyName:      "theme",
		},
		{
			name:         "default int",
			defaultValue: 42,
			keyName:      "count",
		},
		{
			name:         "default bool",
			defaultValue: false,
			keyName:      "enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			c := newComponentImpl("TestComponent")
			ctx := &Context{component: c}

			// Act & Assert based on type
			switch v := tt.defaultValue.(type) {
			case string:
				key := NewProvideKey[string](tt.keyName)
				result := InjectTyped(ctx, key, v)
				assert.Equal(t, v, result, "Should return default value")
			case int:
				key := NewProvideKey[int](tt.keyName)
				result := InjectTyped(ctx, key, v)
				assert.Equal(t, v, result, "Should return default value")
			case bool:
				key := NewProvideKey[bool](tt.keyName)
				result := InjectTyped(ctx, key, v)
				assert.Equal(t, v, result, "Should return default value")
			}
		})
	}
}

// TestInjectTyped_FromParent tests injecting from parent component
func TestInjectTyped_FromParent(t *testing.T) {
	// Arrange
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")
	child.parent = parent

	parentCtx := &Context{component: parent}
	childCtx := &Context{component: child}

	// Create typed key
	themeKey := NewProvideKey[string]("theme")

	// Act - parent provides, child injects
	ProvideTyped(parentCtx, themeKey, "dark")
	result := InjectTyped(childCtx, themeKey, "light")

	// Assert
	assert.Equal(t, "dark", result, "Child should receive parent's provided value")
}

// TestInjectTyped_ComplexTypes tests with complex types like Ref
func TestInjectTyped_ComplexTypes(t *testing.T) {
	t.Run("inject Ref[int]", func(t *testing.T) {
		// Arrange
		parent := newComponentImpl("Parent")
		child := newComponentImpl("Child")
		child.parent = parent

		parentCtx := &Context{component: parent}
		childCtx := &Context{component: child}

		// Create key for Ref[int]
		countKey := NewProvideKey[*Ref[int]]("count")

		// Act - parent provides Ref, child injects
		countRef := NewRef(42)
		ProvideTyped(parentCtx, countKey, countRef)
		injectedRef := InjectTyped(childCtx, countKey, NewRef(0))

		// Assert
		require.NotNil(t, injectedRef, "Injected ref should not be nil")
		assert.Equal(t, countRef, injectedRef, "Should be same Ref instance")
		assert.Equal(t, 42, injectedRef.Get(), "Ref value should be 42")

		// Verify reactivity - changes propagate
		countRef.Set(100)
		assert.Equal(t, 100, injectedRef.Get(), "Child should see reactive changes")
	})

	t.Run("inject struct", func(t *testing.T) {
		// Arrange
		type Config struct {
			Theme string
			Debug bool
		}

		parent := newComponentImpl("Parent")
		child := newComponentImpl("Child")
		child.parent = parent

		parentCtx := &Context{component: parent}
		childCtx := &Context{component: child}

		configKey := NewProvideKey[Config]("config")

		// Act
		cfg := Config{Theme: "dark", Debug: true}
		ProvideTyped(parentCtx, configKey, cfg)
		injectedCfg := InjectTyped(childCtx, configKey, Config{})

		// Assert
		assert.Equal(t, "dark", injectedCfg.Theme)
		assert.True(t, injectedCfg.Debug)
	})
}

// TestInjectTyped_MultipleKeys tests multiple keys with different types
func TestInjectTyped_MultipleKeys(t *testing.T) {
	// Arrange
	parent := newComponentImpl("Parent")
	child := newComponentImpl("Child")
	child.parent = parent

	parentCtx := &Context{component: parent}
	childCtx := &Context{component: child}

	// Create multiple typed keys
	themeKey := NewProvideKey[string]("theme")
	countKey := NewProvideKey[int]("count")
	enabledKey := NewProvideKey[bool]("enabled")

	// Act - provide multiple values
	ProvideTyped(parentCtx, themeKey, "dark")
	ProvideTyped(parentCtx, countKey, 42)
	ProvideTyped(parentCtx, enabledKey, true)

	// Inject all values
	theme := InjectTyped(childCtx, themeKey, "light")
	count := InjectTyped(childCtx, countKey, 0)
	enabled := InjectTyped(childCtx, enabledKey, false)

	// Assert - verify correct types and values
	assert.Equal(t, "dark", theme)
	assert.Equal(t, 42, count)
	assert.True(t, enabled)
}

// TestProvideKey_TypeSafety is a compile-time test
// This test verifies that the type system works correctly
func TestProvideKey_TypeSafety(t *testing.T) {
	// This test primarily validates compile-time type safety
	// If this compiles, type safety is working

	c := newComponentImpl("TestComponent")
	ctx := &Context{component: c}

	// Create typed keys
	stringKey := NewProvideKey[string]("str")
	intKey := NewProvideKey[int]("num")

	// These should compile fine
	ProvideTyped(ctx, stringKey, "hello")
	ProvideTyped(ctx, intKey, 42)

	result1 := InjectTyped(ctx, stringKey, "default")
	result2 := InjectTyped(ctx, intKey, 0)

	// Verify types are correct
	var _ string = result1 // Should compile - result1 is string
	var _ int = result2    // Should compile - result2 is int

	assert.Equal(t, "hello", result1)
	assert.Equal(t, 42, result2)
}
