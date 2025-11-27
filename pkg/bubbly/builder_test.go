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

// TestComponentBuilder_WithMultiKeyBindings tests the variadic WithMultiKeyBindings method.
// This tests Task 2.1: Multi-key binding helper that accepts variadic keys.
func TestComponentBuilder_WithMultiKeyBindings(t *testing.T) {
	tests := []struct {
		name        string
		event       string
		description string
		keys        []string
		wantCount   int
	}{
		{
			name:        "registers all keys correctly",
			event:       "increment",
			description: "Increment counter",
			keys:        []string{"up", "k", "+"},
			wantCount:   3,
		},
		{
			name:        "single key works (equivalent to WithKeyBinding)",
			event:       "decrement",
			description: "Decrement counter",
			keys:        []string{"down"},
			wantCount:   1,
		},
		{
			name:        "empty keys list is no-op",
			event:       "noop",
			description: "No operation",
			keys:        []string{},
			wantCount:   0,
		},
		{
			name:        "multiple keys all emit same event",
			event:       "save",
			description: "Save data",
			keys:        []string{"ctrl+s", "F2", "ctrl+enter"},
			wantCount:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			builder := NewComponent("TestComponent").
				WithMultiKeyBindings(tt.event, tt.description, tt.keys...).
				Template(func(ctx RenderContext) string { return "test" })

			component, err := builder.Build()

			// Assert
			require.NoError(t, err, "Build should succeed")
			require.NotNil(t, component)

			impl := component.(*componentImpl)

			// Verify all keys are registered
			totalBindings := 0
			for _, key := range tt.keys {
				bindings, exists := impl.keyBindings[key]
				if tt.wantCount > 0 {
					assert.True(t, exists, "Key %q should be registered", key)
					assert.Len(t, bindings, 1, "Key %q should have exactly one binding", key)
					if len(bindings) > 0 {
						assert.Equal(t, tt.event, bindings[0].Event, "Key %q should emit event %q", key, tt.event)
						assert.Equal(t, tt.description, bindings[0].Description, "Key %q should have description %q", key, tt.description)
					}
				}
				totalBindings += len(bindings)
			}

			assert.Equal(t, tt.wantCount, totalBindings, "Total bindings should match expected count")
		})
	}
}

// TestComponentBuilder_WithMultiKeyBindings_ChainWithOthers tests chaining with other methods.
func TestComponentBuilder_WithMultiKeyBindings_ChainWithOthers(t *testing.T) {
	t.Run("works with existing WithKeyBinding in same builder", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("TestComponent").
			WithKeyBinding("esc", "cancel", "Cancel operation").
			WithMultiKeyBindings("increment", "Increment", "up", "k", "+").
			WithKeyBinding("enter", "submit", "Submit form").
			Template(func(ctx RenderContext) string { return "test" })

		component, err := builder.Build()

		// Assert
		require.NoError(t, err)
		impl := component.(*componentImpl)

		// Verify multi-key bindings
		assert.Contains(t, impl.keyBindings, "up")
		assert.Contains(t, impl.keyBindings, "k")
		assert.Contains(t, impl.keyBindings, "+")

		// Verify single-key bindings
		assert.Contains(t, impl.keyBindings, "esc")
		assert.Contains(t, impl.keyBindings, "enter")

		// Verify all emit correct events
		assert.Equal(t, "increment", impl.keyBindings["up"][0].Event)
		assert.Equal(t, "increment", impl.keyBindings["k"][0].Event)
		assert.Equal(t, "increment", impl.keyBindings["+"][0].Event)
		assert.Equal(t, "cancel", impl.keyBindings["esc"][0].Event)
		assert.Equal(t, "submit", impl.keyBindings["enter"][0].Event)
	})
}

// TestComponentBuilder_WithMultiKeyBindings_ReturnsBuilder tests method chaining.
func TestComponentBuilder_WithMultiKeyBindings_ReturnsBuilder(t *testing.T) {
	t.Run("returns builder for chaining", func(t *testing.T) {
		// Arrange
		builder1 := NewComponent("TestComponent")

		// Act
		builder2 := builder1.WithMultiKeyBindings("increment", "Increment", "up", "k")

		// Assert
		assert.Same(t, builder1, builder2, "Should return same builder instance for chaining")
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

// TestComponentBuilder_Build tests the Build method.
func TestComponentBuilder_Build(t *testing.T) {
	t.Run("builds valid component with template", func(t *testing.T) {
		// Arrange
		builder := NewComponent("Test").
			Template(func(ctx RenderContext) string {
				return "Hello"
			})

		// Act
		component, err := builder.Build()

		// Assert
		require.NoError(t, err, "Build should succeed with template")
		require.NotNil(t, component, "Component should not be nil")
		assert.Equal(t, "Test", component.Name())
		assert.NotEmpty(t, component.ID())
	})

	t.Run("fails without template", func(t *testing.T) {
		// Arrange
		builder := NewComponent("Test")

		// Act
		component, err := builder.Build()

		// Assert
		require.Error(t, err, "Build should fail without template")
		assert.Nil(t, component, "Component should be nil on error")
		assert.Contains(t, err.Error(), "template is required")
	})

	t.Run("builds with all configuration", func(t *testing.T) {
		// Arrange
		type TestProps struct {
			Label string
		}
		props := TestProps{Label: "Click"}
		setupFn := func(ctx *Context) {}
		templateFn := func(ctx RenderContext) string { return "test" }
		child, childErr := NewComponent("Child").Template(func(ctx RenderContext) string { return "child" }).Build()
		require.NoError(t, childErr)

		// Act
		component, err := NewComponent("Test").
			Props(props).
			Setup(setupFn).
			Template(templateFn).
			Children(child).
			Build()

		// Assert
		require.NoError(t, err)
		require.NotNil(t, component)
		assert.Equal(t, props, component.Props())
	})

	t.Run("returns component implementing Component interface", func(t *testing.T) {
		// Arrange & Act
		component, err := NewComponent("Test").
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		// Assert
		require.NoError(t, err)
		require.NotNil(t, component)

		// Verify it implements Component interface
		var _ Component = component
		assert.Equal(t, "Test", component.Name())
		assert.NotEmpty(t, component.ID())
		assert.NotNil(t, component.Props)
		assert.NotNil(t, component.Emit)
		assert.NotNil(t, component.On)
	})

	t.Run("error message is clear and descriptive", func(t *testing.T) {
		// Arrange
		builder := NewComponent("MyButton")

		// Act
		_, err := builder.Build()

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "template is required")
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("builds component with only template", func(t *testing.T) {
		// Arrange - minimal valid configuration
		builder := NewComponent("Minimal").
			Template(func(ctx RenderContext) string {
				return "minimal"
			})

		// Act
		component, err := builder.Build()

		// Assert
		require.NoError(t, err)
		require.NotNil(t, component)
		assert.Nil(t, component.Props(), "Props should be nil when not set")
	})
}

// TestComponentBuilder_BuildValidation tests validation logic.
func TestComponentBuilder_BuildValidation(t *testing.T) {
	t.Run("accumulates multiple errors", func(t *testing.T) {
		// Arrange - builder with no template
		builder := NewComponent("Test")

		// Act
		_, err := builder.Build()

		// Assert
		require.Error(t, err)
		// Should mention template requirement
		assert.Contains(t, err.Error(), "template")
	})

	t.Run("validates before returning component", func(t *testing.T) {
		// Arrange
		builder := NewComponent("Test")

		// Act
		component, err := builder.Build()

		// Assert - validation happens before component is returned
		assert.Error(t, err)
		assert.Nil(t, component, "Component should be nil when validation fails")
	})
}

// TestComponentBuilder_BuildIntegration tests Build with Bubbletea.
func TestComponentBuilder_BuildIntegration(t *testing.T) {
	t.Run("built component works with Bubbletea", func(t *testing.T) {
		// Arrange
		component, err := NewComponent("Test").
			Template(func(ctx RenderContext) string {
				return "Hello World"
			}).
			Build()

		require.NoError(t, err)

		// Act - use as tea.Model
		_ = component.Init()
		_, _ = component.Update(nil)
		view := component.View()

		// Assert
		assert.Equal(t, "Hello World", view)
	})

	t.Run("built component with setup executes correctly", func(t *testing.T) {
		// Arrange
		setupCalled := false
		component, err := NewComponent("Test").
			Setup(func(ctx *Context) {
				setupCalled = true
			}).
			Template(func(ctx RenderContext) string {
				return "test"
			}).
			Build()

		require.NoError(t, err)

		// Act
		_ = component.Init()

		// Assert
		assert.True(t, setupCalled, "Setup should be called during Init")
	})
}

// TestComponentBuilder_WithAutoCommands tests the WithAutoCommands builder method.
func TestComponentBuilder_WithAutoCommands(t *testing.T) {
	tests := []struct {
		name          string
		autoCommands  bool
		wantEnabled   bool
		wantQueueInit bool
		wantGenInit   bool
	}{
		{
			name:          "enables auto commands when true",
			autoCommands:  true,
			wantEnabled:   true,
			wantQueueInit: true,
			wantGenInit:   true,
		},
		{
			name:          "disables auto commands when false",
			autoCommands:  false,
			wantEnabled:   false,
			wantQueueInit: false,
			wantGenInit:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			component, err := NewComponent("Test").
				WithAutoCommands(tt.autoCommands).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()

			// Assert
			require.NoError(t, err, "Build should succeed")
			require.NotNil(t, component, "Component should not be nil")

			// Type assert to access internal fields
			impl, ok := component.(*componentImpl)
			require.True(t, ok, "Component should be *componentImpl")

			// Verify autoCommands flag
			assert.Equal(t, tt.wantEnabled, impl.autoCommands, "autoCommands flag should match")

			// Verify command queue initialization
			if tt.wantQueueInit {
				assert.NotNil(t, impl.commandQueue, "Command queue should be initialized when auto commands enabled")
			} else {
				assert.Nil(t, impl.commandQueue, "Command queue should be nil when auto commands disabled")
			}

			// Verify command generator initialization
			if tt.wantGenInit {
				assert.NotNil(t, impl.commandGen, "Command generator should be initialized when auto commands enabled")
			} else {
				assert.Nil(t, impl.commandGen, "Command generator should be nil when auto commands disabled")
			}
		})
	}
}

// TestComponentBuilder_WithAutoCommands_FluentAPI tests fluent API chaining.
func TestComponentBuilder_WithAutoCommands_FluentAPI(t *testing.T) {
	t.Run("returns builder for method chaining", func(t *testing.T) {
		// Arrange & Act
		builder := NewComponent("Test").
			WithAutoCommands(true).
			Props("test props").
			Setup(func(ctx *Context) {}).
			Template(func(ctx RenderContext) string { return "test" })

		// Assert
		require.NotNil(t, builder, "Builder should not be nil")
		component, err := builder.Build()
		require.NoError(t, err)
		require.NotNil(t, component)
	})

	t.Run("can be called in any order", func(t *testing.T) {
		// Arrange & Act - call WithAutoCommands at different positions
		component1, err1 := NewComponent("Test1").
			WithAutoCommands(true).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		component2, err2 := NewComponent("Test2").
			Template(func(ctx RenderContext) string { return "test" }).
			WithAutoCommands(true).
			Build()

		component3, err3 := NewComponent("Test3").
			Props("props").
			WithAutoCommands(true).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		// Assert - all should succeed
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
		require.NotNil(t, component1)
		require.NotNil(t, component2)
		require.NotNil(t, component3)

		// All should have auto commands enabled
		impl1 := component1.(*componentImpl)
		impl2 := component2.(*componentImpl)
		impl3 := component3.(*componentImpl)
		assert.True(t, impl1.autoCommands)
		assert.True(t, impl2.autoCommands)
		assert.True(t, impl3.autoCommands)
	})

	t.Run("last call wins when called multiple times", func(t *testing.T) {
		// Arrange & Act
		component, err := NewComponent("Test").
			WithAutoCommands(true).
			WithAutoCommands(false).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		// Assert
		require.NoError(t, err)
		impl := component.(*componentImpl)
		assert.False(t, impl.autoCommands, "Last WithAutoCommands call should win")
	})
}

// TestComponentBuilder_WithAutoCommands_DefaultBehavior tests default behavior without WithAutoCommands.
func TestComponentBuilder_WithAutoCommands_DefaultBehavior(t *testing.T) {
	t.Run("auto commands disabled by default", func(t *testing.T) {
		// Arrange & Act - build without calling WithAutoCommands
		component, err := NewComponent("Test").
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		// Assert
		require.NoError(t, err)
		impl := component.(*componentImpl)
		assert.False(t, impl.autoCommands, "Auto commands should be disabled by default")
		assert.Nil(t, impl.commandQueue, "Command queue should be nil by default")
		assert.Nil(t, impl.commandGen, "Command generator should be nil by default")
	})
}

// TestComponentBuilder_WithAutoCommands_CommandInfrastructure tests command infrastructure initialization.
func TestComponentBuilder_WithAutoCommands_CommandInfrastructure(t *testing.T) {
	t.Run("initializes command queue when enabled", func(t *testing.T) {
		// Arrange & Act
		component, err := NewComponent("Test").
			WithAutoCommands(true).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		// Assert
		require.NoError(t, err)
		impl := component.(*componentImpl)
		require.NotNil(t, impl.commandQueue, "Command queue should be initialized")

		// Verify queue is functional
		assert.Equal(t, 0, impl.commandQueue.Len(), "Queue should start empty")
	})

	t.Run("initializes default command generator when enabled", func(t *testing.T) {
		// Arrange & Act
		component, err := NewComponent("Test").
			WithAutoCommands(true).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		// Assert
		require.NoError(t, err)
		impl := component.(*componentImpl)
		require.NotNil(t, impl.commandGen, "Command generator should be initialized")

		// Verify generator is functional (generates non-nil command)
		cmd := impl.commandGen.Generate("test-id", "ref-1", "old", "new")
		assert.NotNil(t, cmd, "Generator should create non-nil command")
	})

	t.Run("does not initialize infrastructure when disabled", func(t *testing.T) {
		// Arrange & Act
		component, err := NewComponent("Test").
			WithAutoCommands(false).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		// Assert
		require.NoError(t, err)
		impl := component.(*componentImpl)
		assert.Nil(t, impl.commandQueue, "Command queue should not be initialized when disabled")
		assert.Nil(t, impl.commandGen, "Command generator should not be initialized when disabled")
	})
}
