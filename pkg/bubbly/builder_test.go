package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewComponent tests the NewComponent constructor function.
func TestNewComponent(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		wantName      string
	}{
		{
			name:          "creates builder with simple name",
			componentName: "Button",
			wantName:      "Button",
		},
		{
			name:          "creates builder with compound name",
			componentName: "FormInput",
			wantName:      "FormInput",
		},
		{
			name:          "creates builder with empty name",
			componentName: "",
			wantName:      "",
		},
		{
			name:          "creates builder with special characters",
			componentName: "My-Component_123",
			wantName:      "My-Component_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			builder := NewComponent(tt.componentName)

			// Assert
			require.NotNil(t, builder, "NewComponent should return non-nil builder")
			assert.NotNil(t, builder.component, "Builder should have component reference")
			assert.Equal(t, tt.wantName, builder.component.name, "Component name should match")
			assert.NotEmpty(t, builder.component.id, "Component should have unique ID")
			assert.NotNil(t, builder.errors, "Builder should have errors slice")
			assert.Empty(t, builder.errors, "Builder should start with no errors")
		})
	}
}

// TestComponentBuilder_Structure tests the ComponentBuilder struct fields.
func TestComponentBuilder_Structure(t *testing.T) {
	t.Run("builder stores component reference", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test")

		// Assert
		require.NotNil(t, builder.component)
		assert.IsType(t, &componentImpl{}, builder.component)
	})

	t.Run("builder initializes error tracking", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test")

		// Assert
		require.NotNil(t, builder.errors)
		assert.Empty(t, builder.errors)
		assert.Equal(t, 0, len(builder.errors))
	})

	t.Run("component has initialized fields", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test")

		// Assert
		c := builder.component
		assert.NotEmpty(t, c.id, "Component should have ID")
		assert.NotNil(t, c.state, "State map should be initialized")
		assert.NotNil(t, c.handlers, "Handlers map should be initialized")
		assert.NotNil(t, c.children, "Children slice should be initialized")
	})
}

// TestComponentBuilder_UniqueIDs tests that each builder creates components with unique IDs.
func TestComponentBuilder_UniqueIDs(t *testing.T) {
	t.Run("generates unique IDs for multiple components", func(t *testing.T) {
		// Arrange
		count := 10
		ids := make(map[string]bool)

		// Act
		for i := 0; i < count; i++ {
			builder := NewComponent("Test")
			id := builder.component.id

			// Assert uniqueness
			assert.False(t, ids[id], "ID %s should be unique", id)
			ids[id] = true
		}

		// Assert
		assert.Equal(t, count, len(ids), "Should have %d unique IDs", count)
	})

	t.Run("IDs follow expected format", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test")
		id := builder.component.id

		// Assert
		assert.Regexp(t, `^component-\d+$`, id, "ID should match format 'component-N'")
	})
}

// TestComponentBuilder_ErrorTracking tests error tracking functionality.
func TestComponentBuilder_ErrorTracking(t *testing.T) {
	t.Run("errors slice is mutable", func(t *testing.T) {
		// Arrange
		builder := NewComponent("Test")

		// Act - manually add error for testing
		builder.errors = append(builder.errors, assert.AnError)

		// Assert
		assert.Len(t, builder.errors, 1)
		assert.Equal(t, assert.AnError, builder.errors[0])
	})

	t.Run("multiple errors can be tracked", func(t *testing.T) {
		// Arrange
		builder := NewComponent("Test")

		// Act
		builder.errors = append(builder.errors, assert.AnError)
		builder.errors = append(builder.errors, assert.AnError)

		// Assert
		assert.Len(t, builder.errors, 2)
	})
}

// TestComponentBuilder_Concurrency tests thread-safety of component creation.
func TestComponentBuilder_Concurrency(t *testing.T) {
	t.Run("concurrent component creation is safe", func(t *testing.T) {
		// Arrange
		count := 100
		done := make(chan string, count)

		// Act - create components concurrently
		for i := 0; i < count; i++ {
			go func() {
				builder := NewComponent("Test")
				done <- builder.component.id
			}()
		}

		// Collect IDs
		ids := make(map[string]bool)
		for i := 0; i < count; i++ {
			id := <-done
			ids[id] = true
		}

		// Assert
		assert.Equal(t, count, len(ids), "All IDs should be unique even with concurrent creation")
	})
}

// TestComponentBuilder_Props tests the Props method.
func TestComponentBuilder_Props(t *testing.T) {
	type TestProps struct {
		Label    string
		Disabled bool
	}

	tests := []struct {
		name  string
		props interface{}
	}{
		{
			name:  "sets struct props",
			props: TestProps{Label: "Click me", Disabled: false},
		},
		{
			name:  "sets string props",
			props: "simple string",
		},
		{
			name:  "sets int props",
			props: 42,
		},
		{
			name:  "sets nil props",
			props: nil,
		},
		{
			name:  "sets map props",
			props: map[string]interface{}{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			builder := NewComponent("Test").Props(tt.props)

			// Assert
			require.NotNil(t, builder, "Props should return builder")
			assert.Equal(t, tt.props, builder.component.props, "Props should be stored")
		})
	}
}

// TestComponentBuilder_Setup tests the Setup method.
func TestComponentBuilder_Setup(t *testing.T) {
	t.Run("sets setup function", func(t *testing.T) {
		// Arrange
		setupCalled := false
		setupFn := func(ctx *Context) {
			setupCalled = true
		}

		// Act
		builder := NewComponent("Test").Setup(setupFn)

		// Assert
		require.NotNil(t, builder, "Setup should return builder")
		assert.NotNil(t, builder.component.setup, "Setup function should be stored")

		// Verify it's the right function by calling it
		builder.component.setup(&Context{})
		assert.True(t, setupCalled, "Setup function should be callable")
	})

	t.Run("sets nil setup function", func(t *testing.T) {
		// Act
		builder := NewComponent("Test").Setup(nil)

		// Assert
		require.NotNil(t, builder)
		assert.Nil(t, builder.component.setup)
	})
}

// TestComponentBuilder_Template tests the Template method.
func TestComponentBuilder_Template(t *testing.T) {
	t.Run("sets template function", func(t *testing.T) {
		// Arrange
		expectedOutput := "Hello World"
		templateFn := func(ctx RenderContext) string {
			return expectedOutput
		}

		// Act
		builder := NewComponent("Test").Template(templateFn)

		// Assert
		require.NotNil(t, builder, "Template should return builder")
		assert.NotNil(t, builder.component.template, "Template function should be stored")

		// Verify it's the right function by calling it
		output := builder.component.template(RenderContext{})
		assert.Equal(t, expectedOutput, output, "Template function should be callable")
	})

	t.Run("sets nil template function", func(t *testing.T) {
		// Act
		builder := NewComponent("Test").Template(nil)

		// Assert
		require.NotNil(t, builder)
		assert.Nil(t, builder.component.template)
	})
}

// TestComponentBuilder_Children tests the Children method.
func TestComponentBuilder_Children(t *testing.T) {
	t.Run("sets single child", func(t *testing.T) {
		// Arrange
		child := NewComponent("Child").component

		// Act
		builder := NewComponent("Parent").Children(child)

		// Assert
		require.NotNil(t, builder, "Children should return builder")
		assert.Len(t, builder.component.children, 1, "Should have 1 child")
		assert.Equal(t, child, builder.component.children[0], "Child should match")
	})

	t.Run("sets multiple children", func(t *testing.T) {
		// Arrange
		child1 := NewComponent("Child1").component
		child2 := NewComponent("Child2").component
		child3 := NewComponent("Child3").component

		// Act
		builder := NewComponent("Parent").Children(child1, child2, child3)

		// Assert
		require.NotNil(t, builder)
		assert.Len(t, builder.component.children, 3, "Should have 3 children")
		assert.Equal(t, child1, builder.component.children[0])
		assert.Equal(t, child2, builder.component.children[1])
		assert.Equal(t, child3, builder.component.children[2])
	})

	t.Run("sets no children", func(t *testing.T) {
		// Act
		builder := NewComponent("Parent").Children()

		// Assert
		require.NotNil(t, builder)
		assert.Empty(t, builder.component.children, "Should have no children")
	})
}

// TestComponentBuilder_MethodChaining tests fluent API chaining.
func TestComponentBuilder_MethodChaining(t *testing.T) {
	t.Run("all methods return builder for chaining", func(t *testing.T) {
		// Arrange
		type ButtonProps struct {
			Label string
		}
		props := ButtonProps{Label: "Click"}
		setupFn := func(ctx *Context) {}
		templateFn := func(ctx RenderContext) string { return "test" }
		child := NewComponent("Child").component

		// Act - chain all methods
		builder := NewComponent("Test").
			Props(props).
			Setup(setupFn).
			Template(templateFn).
			Children(child)

		// Assert
		require.NotNil(t, builder)
		assert.Equal(t, props, builder.component.props)
		assert.NotNil(t, builder.component.setup)
		assert.NotNil(t, builder.component.template)
		assert.Len(t, builder.component.children, 1)
	})

	t.Run("methods can be called in any order", func(t *testing.T) {
		// Arrange
		child := NewComponent("Child").component

		// Act - different order
		builder := NewComponent("Test").
			Children(child).
			Template(func(ctx RenderContext) string { return "test" }).
			Props("props").
			Setup(func(ctx *Context) {})

		// Assert
		require.NotNil(t, builder)
		assert.NotNil(t, builder.component.props)
		assert.NotNil(t, builder.component.setup)
		assert.NotNil(t, builder.component.template)
		assert.Len(t, builder.component.children, 1)
	})

	t.Run("methods can be called multiple times", func(t *testing.T) {
		// Act - call Props twice
		builder := NewComponent("Test").
			Props("first").
			Props("second")

		// Assert - last call wins
		require.NotNil(t, builder)
		assert.Equal(t, "second", builder.component.props)
	})
}

// TestComponentBuilder_TypeSafety tests type safety of builder methods.
func TestComponentBuilder_TypeSafety(t *testing.T) {
	t.Run("props accepts any type", func(t *testing.T) {
		// Different types should all work
		_ = NewComponent("Test").Props(struct{ Name string }{Name: "test"})
		_ = NewComponent("Test").Props("string")
		_ = NewComponent("Test").Props(123)
		_ = NewComponent("Test").Props([]int{1, 2, 3})
	})

	t.Run("setup accepts SetupFunc", func(t *testing.T) {
		// Valid SetupFunc
		var fn SetupFunc = func(ctx *Context) {}
		builder := NewComponent("Test").Setup(fn)
		assert.NotNil(t, builder)
	})

	t.Run("template accepts RenderFunc", func(t *testing.T) {
		// Valid RenderFunc
		var fn RenderFunc = func(ctx RenderContext) string { return "" }
		builder := NewComponent("Test").Template(fn)
		assert.NotNil(t, builder)
	})

	t.Run("children accepts Component variadic", func(t *testing.T) {
		// Valid Component arguments
		c1 := NewComponent("C1").component
		c2 := NewComponent("C2").component
		builder := NewComponent("Test").Children(c1, c2)
		assert.NotNil(t, builder)
	})
}
