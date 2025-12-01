package bubblyui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui"
)

// TestRootPackageTypes verifies that all core types are accessible from root package.
func TestRootPackageTypes(t *testing.T) {
	t.Run("types are accessible", func(t *testing.T) {
		// Verify types can be used in variable declarations
		// These won't compile if types aren't exported properly
		var _ bubblyui.Component
		var _ bubblyui.ComponentBuilder
		var _ bubblyui.Context
		var _ bubblyui.RenderContext
		var _ bubblyui.Ref[int]
		var _ bubblyui.Computed[int]
		var _ bubblyui.RunOption
	})
}

// TestRootPackageFunctions verifies that all core functions are accessible.
func TestRootPackageFunctions(t *testing.T) {
	t.Run("NewComponent is not nil", func(t *testing.T) {
		assert.NotNil(t, bubblyui.NewComponent, "NewComponent should be exported and not nil")
	})

	t.Run("WatchEffect is not nil", func(t *testing.T) {
		assert.NotNil(t, bubblyui.WatchEffect, "WatchEffect should be exported and not nil")
	})

	t.Run("Run is not nil", func(t *testing.T) {
		assert.NotNil(t, bubblyui.Run, "Run should be exported and not nil")
	})

	// Generic functions are tested via usage in other tests
	t.Run("Generic functions work", func(t *testing.T) {
		// NewRef
		ref := bubblyui.NewRef(42)
		assert.NotNil(t, ref)

		// NewComputed
		computed := bubblyui.NewComputed(func() int { return 10 })
		assert.NotNil(t, computed)

		// Watch - returns cleanup function
		cleanup := bubblyui.Watch(ref, func(n, o int) {})
		assert.NotNil(t, cleanup)
		cleanup()
	})
}

// TestRootPackageRunOptions verifies that run option functions are accessible.
func TestRootPackageRunOptions(t *testing.T) {
	tests := []struct {
		name   string
		option bubblyui.RunOption
	}{
		{"WithAltScreen", bubblyui.WithAltScreen()},
		{"WithMouseAllMotion", bubblyui.WithMouseAllMotion()},
		{"WithMouseCellMotion", bubblyui.WithMouseCellMotion()},
		{"WithFPS", bubblyui.WithFPS(60)},
		{"WithReportFocus", bubblyui.WithReportFocus()},
		{"WithoutBracketedPaste", bubblyui.WithoutBracketedPaste()},
		{"WithoutSignalHandler", bubblyui.WithoutSignalHandler()},
		{"WithoutCatchPanics", bubblyui.WithoutCatchPanics()},
		{"WithInputTTY", bubblyui.WithInputTTY()},
	}

	for _, tt := range tests {
		t.Run(tt.name+" returns valid option", func(t *testing.T) {
			assert.NotNil(t, tt.option, "%s() should return a non-nil RunOption", tt.name)
		})
	}
}

// TestNewComponentFromRoot verifies NewComponent works via root package.
func TestNewComponentFromRoot(t *testing.T) {
	builder := bubblyui.NewComponent("TestComponent")
	require.NotNil(t, builder, "NewComponent should return a builder")

	// Build a minimal component
	component, err := builder.
		Template(func(ctx bubblyui.RenderContext) string {
			return "Hello from root package"
		}).
		Build()

	require.NoError(t, err, "Build should not return error")
	assert.NotNil(t, component, "Component should be created")
	assert.Equal(t, "TestComponent", component.Name())
}

// TestNewRefFromRoot verifies NewRef works via root package.
func TestNewRefFromRoot(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		expected int
	}{
		{"zero value", 0, 0},
		{"positive value", 42, 42},
		{"negative value", -10, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubblyui.NewRef(tt.initial)
			require.NotNil(t, ref, "NewRef should return a ref")
			assert.Equal(t, tt.expected, ref.Get(), "Ref should hold initial value")
		})
	}
}

// TestNewRefSetFromRoot verifies Ref.Set works via root package.
func TestNewRefSetFromRoot(t *testing.T) {
	ref := bubblyui.NewRef(10)
	ref.Set(20)
	assert.Equal(t, 20, ref.Get(), "Ref should update value after Set")
}

// TestNewComputedFromRoot verifies NewComputed works via root package.
func TestNewComputedFromRoot(t *testing.T) {
	ref := bubblyui.NewRef(10)
	computed := bubblyui.NewComputed(func() int {
		return ref.GetTyped() * 2
	})

	require.NotNil(t, computed, "NewComputed should return a computed value")
	assert.Equal(t, 20, computed.GetTyped(), "Computed should derive from ref")

	// Update ref and verify computed updates
	ref.Set(15)
	assert.Equal(t, 30, computed.GetTyped(), "Computed should react to ref changes")
}

// TestWatchFromRoot verifies Watch works via root package.
func TestWatchFromRoot(t *testing.T) {
	ref := bubblyui.NewRef(0)
	var watchedValues []int

	cleanup := bubblyui.Watch(ref, func(newVal, oldVal int) {
		watchedValues = append(watchedValues, newVal)
	})
	defer cleanup()

	ref.Set(1)
	ref.Set(2)
	ref.Set(3)

	assert.Equal(t, []int{1, 2, 3}, watchedValues, "Watch should capture all value changes")
}

// TestWatchEffectFromRoot verifies WatchEffect works via root package.
func TestWatchEffectFromRoot(t *testing.T) {
	ref := bubblyui.NewRef(0)
	effectCount := 0

	cleanup := bubblyui.WatchEffect(func() {
		_ = ref.Get() // Access the ref to establish dependency
		effectCount++
	})
	defer cleanup()

	// Effect runs immediately
	assert.GreaterOrEqual(t, effectCount, 1, "WatchEffect should run immediately")

	initialCount := effectCount
	ref.Set(1)
	assert.Greater(t, effectCount, initialCount, "WatchEffect should run on ref change")
}

// TestGenericTypesWithDifferentTypes verifies generics work with various types.
func TestGenericTypesWithDifferentTypes(t *testing.T) {
	t.Run("Ref with string", func(t *testing.T) {
		ref := bubblyui.NewRef("hello")
		assert.Equal(t, "hello", ref.GetTyped())
		ref.Set("world")
		assert.Equal(t, "world", ref.GetTyped())
	})

	t.Run("Ref with struct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		ref := bubblyui.NewRef(Person{Name: "Alice", Age: 30})
		assert.Equal(t, "Alice", ref.GetTyped().Name)
		ref.Set(Person{Name: "Bob", Age: 25})
		assert.Equal(t, "Bob", ref.GetTyped().Name)
	})

	t.Run("Computed with float64", func(t *testing.T) {
		ref := bubblyui.NewRef(3.14)
		computed := bubblyui.NewComputed(func() float64 {
			return ref.GetTyped() * 2
		})
		assert.InDelta(t, 6.28, computed.GetTyped(), 0.01)
	})
}
